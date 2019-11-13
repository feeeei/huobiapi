package ws

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/feeeei/huobiapi/debug"
	"github.com/feeeei/huobiapi/utils"

	"github.com/bitly/go-simplejson"
)

type HuobiWebSocket struct {
	WS                *SafeWebSocket
	endpoint          string
	Listeners         map[string]Listener
	impl              HuobiWebSocketService
	ListenerMutex     sync.Mutex
	SubscribedTopic   map[string]bool
	SubscribeResultCb map[string]JsonChan
	RequestResultCb   map[string]JsonChan
	autoReconnect     bool          // 掉线后是否自动重连，如果用户主动执行Close()则不自动重连
	LastPing          int64         // 上次接收到的ping时间戳
	HeartbeatInterval time.Duration // 主动发送心跳的时间间隔，默认5秒
	ReceiveTimeout    time.Duration // 接收消息超时时间，默认10秒
}

type JsonChan = chan *simplejson.Json

// Listener 订阅事件监听器
type Listener = func(topic string, json *simplejson.Json)

func NewHuobiWebSocket(endpoint string, service HuobiWebSocketService) (m *HuobiWebSocket, err error) {
	m = &HuobiWebSocket{
		endpoint:          endpoint,
		HeartbeatInterval: 5 * time.Second,
		ReceiveTimeout:    10 * time.Second,
		WS:                nil,
		autoReconnect:     true,
		impl:              service,
		Listeners:         make(map[string]Listener),
		SubscribeResultCb: make(map[string]JsonChan),
		RequestResultCb:   make(map[string]JsonChan),
		SubscribedTopic:   make(map[string]bool),
	}

	if err := m.connect(); err != nil {
		return nil, err
	}

	return m, nil
}

// connect 连接
func (m *HuobiWebSocket) connect() error {
	debug.Println("connecting")
	ws, err := NewSafeWebSocket(m.endpoint)
	if err != nil {
		return err
	}
	m.WS = ws
	m.LastPing = utils.GetUinxMillisecond()
	debug.Println("connected")

	m.handleMessageLoop()
	m.keepAlive()

	return nil
}

// reconnect 重新连接
func (m *HuobiWebSocket) reconnect() error {
	debug.Println("reconnecting after 1s")
	time.Sleep(time.Second)

	if err := m.connect(); err != nil {
		debug.Println(err)
		return err
	}

	// 重新订阅
	m.ListenerMutex.Lock()
	var Listeners = make(map[string]Listener)
	for k, v := range m.Listeners {
		Listeners[k] = v
	}
	m.ListenerMutex.Unlock()

	for topic, listener := range Listeners {
		delete(m.SubscribedTopic, topic)
		m.Subscribe(topic, listener)
	}
	return nil
}

// sendMessage 发送消息
func (m *HuobiWebSocket) SendMessage(data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	debug.Println("sendMessage", string(b))
	m.WS.Send(b)
	return nil
}

// handleMessageLoop 处理消息循环
func (m *HuobiWebSocket) handleMessageLoop() {
	m.WS.Listen(func(buf []byte) {
		msg, err := utils.UnGzipData(buf)
		debug.Println("readMessage", string(msg))
		if err != nil {
			debug.Println(err)
			return
		}
		json, err := simplejson.NewJson(msg)
		if err != nil {
			debug.Println(err)
			return
		}
		m.impl.HandleMessage(json)
	})
}

// keepAlive 保持活跃
func (m *HuobiWebSocket) keepAlive() {
	m.WS.KeepAlive(m.HeartbeatInterval, func() {
		var t = utils.GetUinxMillisecond()
		m.SendMessage(m.impl.HandlePing(t))

		// 检查上次ping时间，如果超过20秒无响应，重新连接
		tr := time.Duration(math.Abs(float64(t - m.LastPing)))
		if tr >= m.HeartbeatInterval*2 {
			debug.Println("no ping max delay", tr, m.HeartbeatInterval*2, t, m.LastPing)
			if m.autoReconnect {
				err := m.reconnect()
				if err != nil {
					debug.Println(err)
				}
			}
		}
	})
}

// Subscribe 订阅
func (m *HuobiWebSocket) Subscribe(topic string, listener Listener) error {
	debug.Println("subscribe", topic)

	var isNew = false

	// 如果未曾发送过订阅指令，则发送，并等待订阅操作结果，否则直接返回
	if _, ok := m.SubscribedTopic[topic]; !ok {
		m.SubscribeResultCb[topic] = make(JsonChan)
		m.SendMessage(subData{ID: topic, Sub: topic})
		isNew = true
	} else {
		debug.Println("send subscribe before, reset listener only")
	}

	m.ListenerMutex.Lock()
	m.Listeners[topic] = listener
	m.ListenerMutex.Unlock()
	m.SubscribedTopic[topic] = true

	if isNew {
		var json = <-m.SubscribeResultCb[topic]
		// 判断订阅结果，如果出错则返回出错信息
		if msg, err := json.Get("err-msg").String(); err == nil {
			return fmt.Errorf(msg)
		}
	}
	return nil
}

// Unsubscribe 取消订阅
func (m *HuobiWebSocket) Unsubscribe(topic string) {
	debug.Println("unSubscribe", topic)

	m.ListenerMutex.Lock()
	// 火币网没有提供取消订阅的接口，只能删除监听器
	delete(m.Listeners, topic)
	m.ListenerMutex.Unlock()
}

// ReConnect 重新连接
func (m *HuobiWebSocket) ReConnect() (err error) {
	debug.Println("reconnect")
	m.autoReconnect = true
	if err = m.WS.Destroy(); err != nil {
		return err
	}
	return m.reconnect()
}

// Close 关闭连接
func (m *HuobiWebSocket) Close() error {
	debug.Println("close")
	m.autoReconnect = false
	if err := m.WS.Destroy(); err != nil {
		return err
	}
	return nil
}

// Loop 进入循环
func (m *HuobiWebSocket) Loop() {
	debug.Println("startLoop")
	for {
		err := m.WS.Loop()
		if err != nil {
			debug.Println(err)
			if err == SafeWebSocketDestroyError {
				break
			} else if m.autoReconnect {
				m.reconnect()
			} else {
				break
			}
		}
	}
	debug.Println("endLoop")
}

package trade

import (
	"fmt"
	"huobiapi/debug"
	"huobiapi/ws"
	"net/url"

	"github.com/bitly/go-simplejson"
)

// MarketEndpoint 行情的Websocket入口
var Endpoint = "wss://api.huobi.pro/ws/v1"

type Trade struct {
	*ws.HuobiWebSocket
	sign *Sign
}

func NewTrade(accessKeyId, accessKeySecret string) (m *Trade, err error) {
	trade := Trade{
		sign: NewSign(accessKeyId, accessKeySecret),
	}
	huobiWebSocket, err := ws.NewHuobiWebSocket(Endpoint, &trade)
	if err != nil {
		return nil, err
	}
	trade.HuobiWebSocket = huobiWebSocket
	return &trade, nil
}

// Auth 资产订单ws需要鉴权验证
func (m *Trade) Auth() error {
	urlInfo, err := url.Parse(Endpoint)
	if err != nil {
		return err
	}

	authMapping := m.sign.Get("GET", urlInfo.Host, urlInfo.Path)
	return m.SendMessage(authMapping)
}

// Subscribe 订阅
func (m *Trade) Subscribe(topic string, listener ws.Listener) error {
	debug.Println("subscribe", topic)

	var isNew = false

	// 如果未曾发送过订阅指令，则发送，并等待订阅操作结果，否则直接返回
	if _, ok := m.SubscribedTopic[topic]; !ok {
		m.SubscribeResultCb[topic] = make(ws.JsonChan)
		m.SendMessage(map[string]interface{}{"op": "sub", "topic": topic})
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

func (m *Trade) HandleMessage(json *simplejson.Json) {
	// 处理ping消息
	if op := json.Get("op").MustString(); op == "ping" {
		m.SendMessage(m.HandlePing(json.Get("ts").MustInt64()))
		return
	}

	// 处理pong消息
	if op := json.Get("op").MustString(); op == "pong" {
		m.LastPing = json.Get("ts").MustInt64()
		return
	}

	// 处理订阅成功通知
	if op := json.Get("op").MustString(); op == "sub" {
		c, ok := m.SubscribeResultCb[json.Get("topic").MustString()]
		if ok {
			c <- json
		}
		return
	}
}

// HandlePing 处理 Ping
func (m *Trade) HandlePing(ping int64) map[string]interface{} {
	debug.Println("handlePing", ping)
	m.LastPing = ping
	return map[string]interface{}{"op": "pong", "ts": ping}
}

// BuildPing 构造 Ping
func (m *Trade) BuildPing(ping int64) map[string]interface{} {
	debug.Println("BuildPing", ping)
	return map[string]interface{}{"op": "ping", "ts": ping}
}

package market

import (
	"fmt"
	"huobiapi/debug"
	"huobiapi/utils"
	"huobiapi/ws"

	"github.com/bitly/go-simplejson"
)

// MarketEndpoint 行情的Websocket入口
var Endpoint = "wss://api.huobi.pro/ws"

type Market struct {
	*ws.HuobiWebSocket
}

// NewMarket 创建Market实例
func NewMarket() (m *Market, err error) {
	var market Market
	huobiWebSocket, err := ws.NewHuobiWebSocket(Endpoint, &market)
	if err != nil {
		return nil, err
	}
	market.HuobiWebSocket = huobiWebSocket
	return &market, nil
}

func (m *Market) HandleMessage(json *simplejson.Json) {
	// 处理ping消息
	if ping := json.Get("ping").MustInt64(); ping > 0 {
		m.SendMessage(m.HandlePing(ping))
		return
	}

	// 处理pong消息
	if pong := json.Get("pong").MustInt64(); pong > 0 {
		m.LastPing = pong
		return
	}

	// 处理订阅消息
	if ch := json.Get("ch").MustString(); ch != "" {
		m.ListenerMutex.Lock()
		listener, ok := m.Listeners[ch]
		m.ListenerMutex.Unlock()
		if ok {
			debug.Println("handleSubscribe", json)
			listener(ch, json)
		}
		return
	}

	// 处理订阅成功通知
	if subbed := json.Get("subbed").MustString(); subbed != "" {
		c, ok := m.SubscribeResultCb[subbed]
		if ok {
			c <- json
		}
		return
	}

	// 请求行情结果
	if rep, id := json.Get("rep").MustString(), json.Get("id").MustString(); rep != "" && id != "" {
		c, ok := m.RequestResultCb[id]
		if ok {
			c <- json
		}
		return
	}

	// 处理错误消息
	if status := json.Get("status").MustString(); status == "error" {
		// 判断是否为订阅失败
		id := json.Get("id").MustString()
		c, ok := m.SubscribeResultCb[id]
		if ok {
			c <- json
		}
		return
	}
}

// Request 请求行情信息
func (m *Market) Request(req string) (*simplejson.Json, error) {
	var id = utils.GetRandomString(10)
	m.RequestResultCb[id] = make(ws.JsonChan)

	if err := m.SendMessage(ws.ReqData{Req: req, ID: id}); err != nil {
		return nil, err
	}
	var json = <-m.RequestResultCb[id]

	delete(m.RequestResultCb, id)

	// 判断是否出错
	if msg := json.Get("err-msg").MustString(); msg != "" {
		return json, fmt.Errorf(msg)
	}
	return json, nil
}

// HandlePing 处理 Ping
func (m *Market) HandlePing(ping int64) map[string]interface{} {
	debug.Println("handlePing", ping)
	m.LastPing = ping
	return map[string]interface{}{"pong": ping}
}

// BuildPing 构造 Ping
func (m *Market) BuildPing(ping int64) map[string]interface{} {
	debug.Println("BuildPing", ping)
	return map[string]interface{}{"ping": ping}
}

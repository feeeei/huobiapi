package ws

import "github.com/bitly/go-simplejson"

type HuobiWebSocketService interface {
	HandlePing(int64) map[string]interface{}
	BuildPing(int64) map[string]interface{}
	HandleMessage(message *simplejson.Json)
}

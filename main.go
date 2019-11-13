package main

import (
	"github.com/leizongmin/huobiapi/client"
	"github.com/leizongmin/huobiapi/market"
	"github.com/leizongmin/huobiapi/trade"
	"github.com/leizongmin/huobiapi/ws"

	"github.com/bitly/go-simplejson"
)

type JSON = simplejson.Json

type ParamsData = client.ParamData
type Market = market.Market
type Listener = ws.Listener
type Client = client.Client

/// 创建WebSocket版Market客户端
func NewMarket() (*market.Market, error) {
	return market.NewMarket()
}

/// 创建WebSoceket版Trade客户端
func NewTrade(accessKeyId, accessKeySecret string) (*trade.Trade, error) {
	return trade.NewTrade(accessKeyId, accessKeySecret)
}

/// 创建RESTFul客户端
func NewClient(accessKeyId, accessKeySecret string) (*client.Client, error) {
	return client.NewClient(client.Endpoint, accessKeyId, accessKeySecret)
}

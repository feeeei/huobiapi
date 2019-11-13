package ws

type ReqData struct {
	Req string `json:"req"`
	ID  string `json:"id"`
}

type subData struct {
	Sub string `json:"sub"`
	ID  string `json:"id"`
}

type pongData struct {
	Pong int64 `json:"pong"`
}

type pingData struct {
	Ping int64 `json:"ping"`
}

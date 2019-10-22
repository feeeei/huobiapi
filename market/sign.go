package market

import (
	"time"

	"github.com/leizongmin/huobiapi/sign"
)

type Sign struct {
	*sign.Sign
}

func NewSign(accessKeyId, accessKeySecret string) *Sign {
	return &Sign{sign.NewSign(accessKeyId, accessKeySecret)}
}

func (s *Sign) Get(method, host, path string) map[string]string {
	var str = method + "\n" + host + "\n" + path + "\n"
	params := map[string]string{
		"AccessKeyId":      s.AccessKeyId,
		"SignatureMethod":  s.SignatureMethod,
		"SignatureVersion": s.SignatureVersion,
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05"),
	}
	str += s.EncodeQueryString(params)
	params["op"] = "auth"
	params["Signature"] = s.ComputeHmac256(str, s.AccessKeySecret)
	return params
}

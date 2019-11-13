package client

import "huobiapi/sign"

type Sign struct {
	*sign.Sign
}

func NewSign(accessKeyId, accessKeySecret string) *Sign {
	return &Sign{sign.NewSign(accessKeyId, accessKeySecret)}
}

func (s *Sign) Get(method, host, path, timestamp string, params map[string]string) (string, error) {
	var str = method + "\n" + host + "\n" + path + "\n"
	params["AccessKeyId"] = s.AccessKeyId
	params["SignatureMethod"] = s.SignatureMethod
	params["SignatureVersion"] = s.SignatureVersion
	params["Timestamp"] = timestamp
	str += s.EncodeQueryString(params)
	return s.ComputeHmac256(str, s.AccessKeySecret), nil
}

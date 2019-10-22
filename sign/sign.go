package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"sort"
	"strings"
)

func getMapKeys(m map[string]string) (keys []string) {
	for k, _ := range m {
		keys = append(keys, k)
	}
	return keys
}

func sortKeys(keys []string) []string {
	sort.Strings(keys)
	return keys
}

func (s *Sign) ComputeHmac256(data string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

/// 拼接query字符串
func (s *Sign) EncodeQueryString(query map[string]string) string {
	var keys = sortKeys(getMapKeys(query))
	var len = len(keys)
	var lines = make([]string, len)
	for i := 0; i < len; i++ {
		var k = keys[i]
		lines[i] = url.QueryEscape(k) + "=" + url.QueryEscape(query[k])
	}
	return strings.Join(lines, "&")
}

type Sign struct {
	AccessKeyId      string
	AccessKeySecret  string
	SignatureMethod  string
	SignatureVersion string
}

func NewSign(accessKeyId, accessKeySecret string) *Sign {
	return &Sign{
		AccessKeyId:      accessKeyId,
		AccessKeySecret:  accessKeySecret,
		SignatureMethod:  "HmacSHA256",
		SignatureVersion: "2",
	}
}

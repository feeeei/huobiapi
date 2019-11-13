package utils

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"math/rand"
	"time"
)

var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// GetRandomString 返回随机字符串
func GetRandomString(n uint) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// GetUinxMillisecond 取毫秒时间戳
func GetUinxMillisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// UnGzipData 解压gzip的数据
func UnGzipData(buf []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}

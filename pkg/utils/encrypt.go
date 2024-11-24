package utils

import (
	"github.com/jxskiss/base62"
)

const encodeStd = "Vsc2D3nebguQ4IAp8XtLyO9iExjPM5mJCUwqr0ThoZ71GNdBKHzYFlkWSafR6v" // 打乱顺序防止被解码

func init() {
	base62.StdEncoding = base62.NewEncoding(encodeStd)
}

func Encrypt(src string) string {
	return string(base62.Encode([]byte(src)))
}

func Decrypt(src string) (string, error) {
	b, err := base62.Decode([]byte(src))
	return string(b), err
}

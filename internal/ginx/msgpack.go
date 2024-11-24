package ginx

import (
	"bytes"
	"compress/zlib"
	"eggServer/internal/config"
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/shamaton/msgpack/v2"
	"io"
	"net/http"
)

var MsgPack binding.BindingBody = msgpackBinding{}

type msgpackBinding struct{}

func (msgpackBinding) Name() string {
	return "msgpack"
}

func (msgpackBinding) Bind(req *http.Request, obj any) error {
	return decodeMsgPack(req.Body, obj)
}

func (msgpackBinding) BindBody(body []byte, obj any) error {
	return decodeMsgPack(bytes.NewReader(body), obj)
}

func decodeMsgPack(r io.Reader, obj any) error {
	// 读取原始数据流
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// 数据解密
	for i := range data {
		data[i] ^= config.C.DataKey
	}

	// 创建 zlib 解压缩器
	zlibReader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating zlib reader: %w", err)
	}
	defer zlibReader.Close()

	// 将解压缩的数据写入一个缓冲区
	var decompressedData bytes.Buffer
	if _, err := io.Copy(&decompressedData, zlibReader); err != nil {
		return fmt.Errorf("error reading decompressed data: %w", err)
	}

	// 使用解压缩后的数据进行解码
	if err := msgpack.Unmarshal(decompressedData.Bytes(), obj); err != nil {
		return fmt.Errorf("error unmarshaling data: %w", err)
	}

	// 验证对象（如果有验证器）
	if binding.Validator != nil {
		if err := binding.Validator.ValidateStruct(obj); err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
	}

	return nil
}

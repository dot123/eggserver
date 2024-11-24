package ginx

import (
	"bytes"
	"compress/zlib"
	"eggServer/internal/config"
	"eggServer/internal/constant"
	"eggServer/pkg/errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/shamaton/msgpack/v2"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

const (
	prefix     = "egg"                // 自定义前缀
	ReqBodyKey = prefix + "/req-body" // 请求体的键
	ResBodyKey = prefix + "/res-body" // 响应体的键
)

// ResponseData 数据返回结构体
type ResponseData struct {
	Code int         `json:"code" msgpack:"code"`           // 状态码
	Data interface{} `json:"data,omitempty" msgpack:"data"` // 数据内容
}

// ResponseFail 返回失败的结构体
type ResponseFail struct {
	Code int    `json:"code" msgpack:"code"` // 状态码
	Msg  string `json:"msg" msgpack:"msg"`   // 错误信息
}

// 验证器
// 如果对象有一个名为 "Verify" 的方法，它会调用该方法进行验证。
// 如果 "Verify" 方法返回非空的字符串，则返回错误。
func verify(obj interface{}) error {
	t := reflect.TypeOf(obj)
	m, ok := t.MethodByName("Verify")
	if ok {
		args := []reflect.Value{reflect.ValueOf(obj)}
		resultList := m.Func.Call(args)
		msg := resultList[0].String()
		if msg != "" {
			return errors.New(msg)
		}
	}
	return nil
}

// ResData 返回数据的接口
// 用于返回带有数据的成功响应
func ResData(c *gin.Context, code int, data interface{}) {
	resp := ResponseData{
		Code: code,
		Data: data,
	}
	ResJSON(c, http.StatusOK, resp)
}

// ResOk 返回操作成功的响应
// 不带数据的成功响应
func ResOk(c *gin.Context) {
	resp := ResponseData{
		Code: constant.OK,
	}
	ResJSON(c, http.StatusOK, resp)
}

// ResJSON 返回 JSON 数据
// 将数据序列化为 msgpack 格式，然后使用 zlib 进行压缩，最后将压缩后的数据作为响应返回
func ResJSON(c *gin.Context, httpCode int, resp interface{}) {
	contentType := c.ContentType()

	switch contentType {
	case "application/json":
		if config.C.IsDebugMode() {
			c.JSON(httpCode, resp)
			c.Abort()
		}
	case "application/x-1I9EK5kMNs":
		data, err := msgpack.Marshal(resp)
		if err != nil {
			fmt.Println("Error marshaling data:", err)
			return
		}

		// 创建一个缓冲区来保存压缩后的数据
		var compressedData bytes.Buffer

		// 创建一个新的 zlib 写入器
		w := zlib.NewWriter(&compressedData)

		// 写入数据进行压缩
		if _, err := w.Write(data); err != nil {
			fmt.Println("Error writing compressed data:", err)
			return
		}

		// 关闭写入器以确保所有数据都被写入
		if err := w.Close(); err != nil {
			fmt.Println("Error closing zlib writer:", err)
			return
		}

		// 加密数据
		b := compressedData.Bytes()
		for i, v := range b {
			b[i] = v ^ config.C.DataKey
		}

		// 发送压缩后的数据
		c.Data(httpCode, "application/x-1I9EK5kMNs", b)
		c.Abort()
	}
}

// ResError 返回错误响应
// 根据错误类型生成对应的错误响应
func ResError(c *gin.Context, status int, err error) {
	if err != nil {
		var e *errors.ResponseError
		if errors.As(err, &e) {
			ResJSON(c, status, ResponseFail{Code: e.Code, Msg: e.Error()})
		} else {
			ResJSON(c, status, ResponseFail{Code: constant.UnknownError, Msg: err.Error()})
		}
	}
}

// GetPage 获取分页参数
// 从查询参数中获取页码和每页条数
func GetPage(c *gin.Context) (pageNum, pageSize int) {
	pageNum, _ = strconv.Atoi(c.Query("page"))
	pageSize, _ = strconv.Atoi(c.Query("limit"))
	if pageSize == 0 {
		pageSize = 10 // 默认每页条数为 10
	}
	if pageNum == 0 {
		pageNum = 1 // 默认页码为 1
	}
	return
}

// GetToken 从请求头中获取 JWT 令牌
// 从 "Authorization" 头部中提取 "Bearer" 开头的令牌
func GetToken(c *gin.Context) string {
	token := c.GetHeader("Authorization")
	prefix := "Bearer "
	if token != "" && strings.HasPrefix(token, prefix) {
		token = token[len(prefix):]
	}
	return token
}

// GetBodyData 从上下文中获取请求体数据
// 返回存储在上下文中的请求体数据
func GetBodyData(c *gin.Context) []byte {
	if v, ok := c.Get(ReqBodyKey); ok {
		if b, ok := v.([]byte); ok {
			return b
		}
	}
	return nil
}

// ParseParamID 从 URL 参数中解析 ID
// 将 URL 参数解析为 uint64 类型的 ID
func ParseParamID(c *gin.Context, key string) uint64 {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// ParseJSON 解析请求体中的 JSON 数据到结构体
// 将请求体的 JSON 数据绑定到指定的结构体，并进行验证
func ParseJSON(c *gin.Context, obj interface{}) error {
	contentType := c.ContentType()

	switch contentType {
	case "application/json":
		if config.C.IsDebugMode() {
			c.ShouldBindJSON(&obj)
		}
	case "application/x-1I9EK5kMNs":
		c.ShouldBindBodyWith(obj, MsgPack)
	}

	return verify(obj)
}

// ParseQuery 解析查询参数到结构体
// 将查询参数绑定到指定的结构体，并进行验证
func ParseQuery(c *gin.Context, obj interface{}) error {
	c.ShouldBindQuery(obj)
	return verify(obj)
}

// Bind 根据请求方法和内容类型自动选择绑定引擎
// 将请求数据绑定到指定的结构体，并进行验证
func Bind(c *gin.Context, obj interface{}) error {
	c.Bind(obj)
	return verify(obj)
}

// ParseForm 解析表单数据到结构体
// 将表单数据绑定到指定的结构体，并进行验证
func ParseForm(c *gin.Context, obj interface{}) error {
	c.ShouldBindWith(obj, binding.Form)
	return verify(obj)
}

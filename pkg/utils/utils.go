package utils

import (
	"github.com/bwmarrin/snowflake"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

// IsDifferentDays 判断两个时间是否在不同的日历天上，根据指定时区
func IsDifferentDays(t1, t2 time.Time, name string) bool {
	location, _ := time.LoadLocation(name)
	// 将两个时间都转换到指定时区
	t1InLoc := t1.In(location)
	t2InLoc := t2.In(location)

	// 比较日期部分（年、月、日）
	return t1InLoc.Year() != t2InLoc.Year() ||
		t1InLoc.Month() != t2InLoc.Month() ||
		t1InLoc.Day() != t2InLoc.Day()
}

// IsSameDays 判断两个时间是否在相同的日历天上，根据指定时区
func IsSameDays(t1, t2 time.Time, name string) bool {
	location, _ := time.LoadLocation(name)
	// 将两个时间都转换到指定时区
	t1InLoc := t1.In(location)
	t2InLoc := t2.In(location)

	// 比较日期部分（年、月、日）
	return t1InLoc.Year() == t2InLoc.Year() &&
		t1InLoc.Month() == t2InLoc.Month() &&
		t1InLoc.Day() == t2InLoc.Day()
}

// CalculateDaysDifference 计算两个时间戳之间相差的天数
func CalculateDaysDifference(timestamp1, timestamp2 int64) int {
	t1 := time.Unix(timestamp1, 0).Truncate(24 * time.Hour) // 只保留日期部分
	t2 := time.Unix(timestamp2, 0).Truncate(24 * time.Hour) // 只保留日期部分

	diff := t2.Sub(t1)            // 计算差值
	return int(diff.Hours() / 24) // 转换为整天
}

// ReverseStr 反转字符串
func ReverseStr(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// ShuffleString 打乱字符串
func ShuffleString(s string) string {
	r := []rune(s)
	for i := len(r) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

var node *snowflake.Node

func init() {
	var err error
	// 从环境变量获取节点 ID
	nodeID := os.Getenv("NODE_ID")
	node, err = snowflake.NewNode(cast.ToInt64(nodeID)) // 节点 ID
	if err != nil {
		log.Fatalf("Failed to create snowflake node: %v", err)
	}
}

// GenerateID 生成一个唯一的 Snowflake ID
func GenerateID() int64 {
	return int64(node.Generate())
}

// ToInt 将字符串转换为整数
func ToInt(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		log.Println("Conversion error:", err)
		return 0
	}
	return num
}

// Copy 复制结构体
func Copy(toValue interface{}, fromValue interface{}) {
	copier.Copy(toValue, fromValue)
}

// Request http请求
func Request(method, url string, body io.Reader) ([]byte, error) {
	// 创建HTTP客户端
	client := &http.Client{}

	// 创建HTTP请求
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Println("Request error:", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送HTTP请求
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request error:", err)
		return nil, err
	}

	defer resp.Body.Close() // 确保关闭响应体

	// 读取响应体
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Request error:", err)
		return nil, err
	}

	return data, nil
}

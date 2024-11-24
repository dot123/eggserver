package utils

import (
	"fmt"
	"math/rand"
)

// InArray 检查数组中是否包含指定的值
func InArray[T comparable](arr []T, val T) bool {
	for _, element := range arr {
		if element == val {
			return true
		}
	}
	return false
}

// IndexOf 返回指定值在数组中的索引位置，若值不存在则返回 -1
func IndexOf[T comparable](arr []T, val T) int {
	for i, element := range arr {
		if element == val {
			return i
		}
	}
	return -1
}

// RemoveElement 从数组中移除指定的元素，返回新的数组和一个布尔值表示是否成功移除
func RemoveElement[T comparable](arr []T, val T) ([]T, bool) {
	ret := false
	result := make([]T, 0, len(arr)) // Using len(arr) to optimize capacity
	for _, element := range arr {
		if element != val {
			result = append(result, element)
		} else {
			ret = true
		}
	}
	return result, ret
}

// RandomElement 从数组中随机选择一个元素
func RandomElement[T any](slice []T) (T, error) {
	var zero T // 零值，用于在切片为空时返回
	if len(slice) == 0 {
		return zero, fmt.Errorf("slice is empty")
	}
	randomIndex := rand.Intn(len(slice))
	return slice[randomIndex], nil
}

// DeepCopyArray 数组深拷贝函数
func DeepCopyArray[T any](src []T) []T {
	dest := make([]T, len(src))
	copy(dest, src)
	return dest
}

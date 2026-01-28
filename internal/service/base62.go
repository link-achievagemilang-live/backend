package service

import (
	"errors"
	"strings"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var base62Map map[rune]int64

func init() {
	base62Map = make(map[rune]int64)
	for i, c := range base62Chars {
		base62Map[c] = int64(i)
	}
}

// Encode converts a 64-bit integer to a Base62 string
func Encode(num int64) string {
	if num == 0 {
		return string(base62Chars[0])
	}

	var result strings.Builder
	base := int64(len(base62Chars))

	for num > 0 {
		remainder := num % base
		result.WriteByte(base62Chars[remainder])
		num = num / base
	}

	// Reverse the string
	encoded := result.String()
	runes := []rune(encoded)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// Decode converts a Base62 string back to a 64-bit integer
func Decode(str string) (int64, error) {
	if str == "" {
		return 0, errors.New("empty string cannot be decoded")
	}

	var num int64
	base := int64(len(base62Chars))

	for _, c := range str {
		val, ok := base62Map[c]
		if !ok {
			return 0, errors.New("invalid character in base62 string")
		}
		num = num*base + val
	}

	return num, nil
}

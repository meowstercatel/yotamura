package common

import (
	"fmt"
	"math/rand"
	"strings"
)

func GetType(variable any) string {
	typeName := strings.Split(fmt.Sprintf("%T", variable), ".")
	return typeName[len(typeName)-1]
}

func CreateMessage(data any) Message {
	return Message{Type: GetType(data), Data: data}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

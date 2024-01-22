package util

import (
	"math/rand"
	"time"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func GetRandomString(stringLen int) string {
	alphabet := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	randomString := make([]rune, stringLen)
	for i := range randomString {
		randomString[i] = alphabet[rnd.Intn(len(alphabet))]
	}

	return string(randomString)
}

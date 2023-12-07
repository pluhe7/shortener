package util

import (
	"math/rand"
)

var rnd = rand.New(rand.NewSource(1))

func GetRandomString(stringLen int) string {
	alphabet := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	randomString := make([]rune, stringLen)
	for i := range randomString {
		randomString[i] = alphabet[rnd.Intn(len(alphabet))]
	}

	return string(randomString)
}

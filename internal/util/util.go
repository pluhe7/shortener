package util

import "math/rand"

func GetRandomString(stringLen int) string {
	alphabet := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	randomString := make([]rune, stringLen)
	for i := range randomString {
		randomString[i] = alphabet[rand.Intn(len(alphabet))]
	}

	return string(randomString)
}

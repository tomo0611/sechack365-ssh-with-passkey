package main

import (
	"math/rand"
	"time"
)

func generateRandomLoginToken() string {
	// generate a random string
	rand.New(rand.NewSource(time.Now().UnixNano()))
	// Generate a random string of length 10
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	length := 32
	randomString := ""
	for i := 0; i < length; i++ {
		randomString += string(letters[rand.Intn(len(letters))])
	}
	return randomString
}

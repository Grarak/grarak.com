package jodirect

import (
	"crypto/rand"
	"../utils"
	"time"
)

const TOKEN_SYMBOLS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Message struct {
	Content string
	Date    time.Time
}

type Token struct {
	Value    string `json:"token"`
	Usable   bool `json:"-"`
	Password string `json:"password,omitempty"`

	Messages       map[string]Message `json:"-"`
	Expiration     time.Time `json:"-"`
	Timeleft       float64 `json:"timeleft,omitempty"`
	SortedMessages []string `json:"messages,omitempty"`
}

func NewToken() *Token {
	token := &Token{
		Value:    "00000",
		Usable:   false,
		Password: generatePassword(),
		Messages: make(map[string]Message),
	}
	return token
}

func generatePassword() string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	utils.Panic(err)

	return utils.ToURLBase64(buf)
}

func (token *Token) Increment() {
	token._increment(len(token.Value) - 1)
}

func (token *Token) _increment(position int) {
	if position < 0 {
		token.Value = string(TOKEN_SYMBOLS[1]) + token.Value
		return
	}

	tokChars := []byte(token.Value)
	charsLen := len(TOKEN_SYMBOLS)
	for index, char := range TOKEN_SYMBOLS {
		if char == int32(tokChars[position]) {
			if index >= charsLen-1 {
				tokChars[position] = TOKEN_SYMBOLS[0]
				token.Value = string(tokChars)
				token._increment(position - 1)
			} else {
				tokChars[position] = TOKEN_SYMBOLS[index+1]
				token.Value = string(tokChars)
			}
			break
		}
	}
}

func (token Token) String() string {
	return token.Value
}

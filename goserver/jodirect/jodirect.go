package jodirect

import (
	"../utils"
	"time"
	"sort"
)

type loginCount struct {
	lastTime time.Time
	count    int
}

type JoDirectData struct {
	tokens []*Token

	tokenGenAddrs map[string]time.Time
	loginAddrs    map[string]*loginCount
}

func NewJoDirectData() *JoDirectData {
	return &JoDirectData{
		tokenGenAddrs: make(map[string]time.Time),
		loginAddrs:    make(map[string]*loginCount),
	}
}

type JoDirectError string

func (e JoDirectError) Error() string {
	return string(e)
}

func GetGreater(token1, token2 string) string {
	return _getGreater(token1, token2, 0)
}

func _getGreater(token1, token2 string, position int) string {
	token1len := len(token1)
	token2len := len(token2)

	if token1 == token2 || (position >= token1len && position >= token2len) {
		return token1
	}

	if token2len > token1len {
		return _getGreater(token2, token1, position)
	}
	if token1len > token2len {
		for i := 0; i < token1len-token2len; i++ {
			token2 = string(TOKEN_SYMBOLS[0]) + token2
		}
	}

	position1 := -1
	position2 := -1
	for index, char := range TOKEN_SYMBOLS {
		if position1 == -1 && char == int32(token1[position]) {
			position1 = index
		}
		if position2 == -1 && char == int32(token2[position]) {
			position2 = index
		}
		if position1 >= 0 && position2 >= 0 {
			break
		}
	}

	if position1 > position2 {
		return token1
	} else if position1 < position2 {
		return token2
	}

	return _getGreater(token1, token2, position+1)
}

func (data *JoDirectData) LoginIPUsable(ipAddr string) float64 {
	currentTime := time.Now()
	if loginCount, ok := data.loginAddrs[ipAddr]; ok {
		if loginCount.count <= 5 {
			loginCount.count++
			loginCount.lastTime = currentTime
			return 0
		}
		if timeleft := loginCount.lastTime.Add(time.Minute * 5).Sub(
			currentTime).Seconds(); timeleft > 0 {
			return timeleft
		}
	}

	data.loginAddrs[ipAddr] = &loginCount{currentTime, 1}
	return 0
}

func (data *JoDirectData) TokenGenIPUsable(ipAddr string) float64 {
	currentTime := time.Now()
	if ipaddrTime, ok := data.tokenGenAddrs[ipAddr]; ok {
		if timeleft := ipaddrTime.Sub(currentTime).Seconds(); timeleft > 0 {
			return timeleft
		}
	}

	data.tokenGenAddrs[ipAddr] = currentTime.Add(time.Minute * 5)
	return 0
}

func (data *JoDirectData) findToken(token string) (int, *Token) {
	tokens := data.tokens
	bufSlice := make([]interface{}, len(tokens))
	for index := range tokens {
		bufSlice[index] = tokens[index]
	}

	index, err := utils.FindinSortedList(bufSlice, func(i, j interface{}) bool {
		return i.(string) == j.(*Token).Value
	}, func(i, j interface{}) bool {
		return i.(string) == GetGreater(i.(string), j.(*Token).Value)
	}, token, false)
	if err == nil {
		return index, bufSlice[index].(*Token)
	}
	return -1, nil
}

func (data *JoDirectData) GetToken(tokenValue, password, ipAddr string) (Token, JoDirectErrorCode) {
	_, token := data.findToken(tokenValue)
	if token == nil {
		return Token{}, INVALID_TOKEN
	}

	currentTime := time.Now()
	timeleft := token.Expiration.Sub(currentTime).Seconds()
	if timeleft <= 0 {
		return Token{}, TOKEN_EXPIRED
	}
	if token.Password != password {
		return Token{}, WRONG_PASSWORD
	}

	var messages []Message
	for _, message := range token.Messages {
		messages = append(messages, message)
	}
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Date.Sub(messages[j].Date).Seconds() > 0
	})

	ret := *token
	ret.Password = ""
	ret.Timeleft = timeleft
	ret.SortedMessages = make([]string, 0)
	for _, message := range messages {
		ret.SortedMessages = append(ret.SortedMessages, message.Content)
	}

	delete(data.loginAddrs, ipAddr)

	return ret, NO_ERROR
}

func (data *JoDirectData) AddTokenMessage(tokenValue, ipaddr, message string) JoDirectErrorCode {
	_, token := data.findToken(tokenValue)
	if token == nil {
		return INVALID_TOKEN
	}

	currentTime := time.Now()
	if token.Expiration.Sub(currentTime).Seconds() <= 0 {
		return TOKEN_EXPIRED
	}

	if _, ok := token.Messages[ipaddr]; ok {
		return MESSAGE_ALREADY_SENT
	}
	token.Messages[ipaddr] = Message{message, currentTime}

	return NO_ERROR
}

func (data *JoDirectData) GenerateToken() Token {
	data.tokens = append(data.tokens, NewToken())
	index := len(data.tokens) - 1
	token := data.tokens[index]

	if index != 0 {
		var usableToken Token
		incrementCount := 1
		for i := index - 1; i >= 0; i-- {
			bufToken := data.tokens[i]
			if bufToken.Usable {
				usableToken = *bufToken
				break
			}
			incrementCount++
		}
		token.Value = usableToken.Value
		for i := 0; i < incrementCount; i++ {
			token.Increment()
		}
	}

	token.Usable = true
	token.Expiration = time.Now().Add(time.Minute * 20)
	return *token
}

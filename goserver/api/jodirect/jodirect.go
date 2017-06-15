package api_jodirect

import (
	"../../miniserver"
	"../../utils"
	"../../jodirect"
	"net/http"
	"encoding/json"
)

type JoDirectApi struct {
	client       *miniserver.Client
	path         string
	version      string
	joDirectData *jodirect.JoDirectData
}

func (joDirectApi JoDirectApi) GetResponse() *miniserver.Response {
	switch joDirectApi.version {
	case "v1":
		return joDirectApi.joDirectApiv1()
	}
	return nil
}

func NewJoDirectApi(
	client *miniserver.Client,
	path,
	version string,
	joDirectData *jodirect.JoDirectData,
) JoDirectApi {
	return JoDirectApi{
		client,
		path,
		version,
		joDirectData,
	}
}

func createStatusResponse(status jodirect.JoDirectErrorCode) interface{} {
	type response struct {
		Status int `json:"status"`
	}
	return response{int(status)}
}

func (joDirectApi JoDirectApi) createResponse(content interface{}, success bool) *miniserver.Response {
	type message struct {
		Content interface{} `json:"content,omitempty"`
		Success bool `json:"success"`
	}

	m := message{Success: success}
	if content != nil {
		m.Content = content
	}
	buf, err := json.Marshal(m)
	utils.Panic(err)
	return joDirectApi.client.ResponseBody(string(buf))
}

func (joDirectApi JoDirectApi) tokenGenerate() (interface{}, bool) {
	if timeleft := joDirectApi.joDirectData.TokenGenIPUsable(
		joDirectApi.client.IPAddr); timeleft > 0 {
		return struct{ Timeleft float64 `json:"timeleft"` }{timeleft}, false
	}

	return joDirectApi.joDirectData.GenerateToken(), true
}

func (joDirectApi JoDirectApi) messageReceived() (interface{}, bool) {
	if timeleft := joDirectApi.joDirectData.LoginIPUsable(
		joDirectApi.client.IPAddr); timeleft > 0 {
		return struct{ Timeleft float64 `json:"timeleft"` }{timeleft}, false
	}

	tokenValue, tokenOk := joDirectApi.client.Queries["token"]
	password, passwordOk := joDirectApi.client.Queries["password"]

	if !tokenOk || !passwordOk {
		return nil, false
	}

	token, status := joDirectApi.joDirectData.GetToken(tokenValue[0], password[0], joDirectApi.client.IPAddr)
	if status != jodirect.NO_ERROR {
		return createStatusResponse(status), false
	}
	return token, true
}

func (joDirectApi JoDirectApi) messageSend() (interface{}, bool) {

	type message struct {
		Token   string `json:"token"`
		Content string `json:"content"`
	}

	var mes message
	err := json.Unmarshal(joDirectApi.client.Request, &mes)
	if err != nil || utils.StringEmpty(mes.Token) || utils.StringEmpty(mes.Token) {
		return nil, false
	}

	status := joDirectApi.joDirectData.AddTokenMessage(mes.Token,
		joDirectApi.client.IPAddr, mes.Content)
	return createStatusResponse(status), status == jodirect.NO_ERROR
}

func (joDirectApi JoDirectApi) joDirectApiv1() *miniserver.Response {
	var response interface{}
	success := true

	switch joDirectApi.path {
	case "token/generate":
		if joDirectApi.client.Method == http.MethodGet {
			response, success = joDirectApi.tokenGenerate()
		}
		break

	case "message/received":
		if joDirectApi.client.Method == http.MethodGet {
			response, success = joDirectApi.messageReceived()
		}
		break
	case "message/send":
		if joDirectApi.client.Method == http.MethodPost && len(joDirectApi.client.Request) > 0 {
			response, success = joDirectApi.messageSend()
		}
		break
	}

	if response == nil {
		res := joDirectApi.createResponse(nil, false)
		res.SetStatusCode(http.StatusNotFound)
		return res
	}
	return joDirectApi.createResponse(response, success)
}

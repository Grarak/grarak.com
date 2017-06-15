package miniserver

import (
	"net/http"
	"io/ioutil"
)

type Client struct {
	Url, Method, IPAddr string
	Request             []byte
	Queries             map[string][]string
}

func newClient(request *http.Request) *Client {
	defer request.Body.Close()

	body, _ := ioutil.ReadAll(request.Body)

	// Extract the ip address
	addChar := false
	addrChars := []byte(request.RemoteAddr)
	var ipAddrBuf []byte
	for i := len(addrChars) - 1; i >= 0; i-- {
		if addChar {
			ipAddrBuf = append([]byte{addrChars[i]}, ipAddrBuf...)
		} else if addrChars[i] == ':' {
			addChar = true
		}
	}

	return &Client{
		request.URL.Path,
		request.Method,
		string(ipAddrBuf),
		body,
		request.Form,
	}
}

func (client *Client) ResponseBody(body string) *Response {
	return newResponseBody(body)
}

func (client *Client) ResponseFile(file string) *Response {
	return newResponseFile(file)
}

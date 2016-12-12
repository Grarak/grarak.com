package miniserver

import (
	"io"
	"net/http"
)

type Client struct {
	Url, Method string
	Request     []byte
	Queries     map[string][]string
}

func newClient(request *http.Request) *Client {
	defer request.Body.Close()

	buf := make([]byte, 1)
	body := make([]byte, 0)
	for {
		_, err := request.Body.Read(buf)
		body = append(body, buf[0])
		if err == io.EOF {
			break
		}
	}

	return &Client{
		Url:     request.URL.Path,
		Method:  request.Method,
		Request: body,
		Queries: request.Form,
	}
}

func (client *Client) ResponseBody(body string) *Response {
	return newResponseBody(body)
}

func (client *Client) ResponseFile(file string) *Response {
	return newResponseFile(file)
}

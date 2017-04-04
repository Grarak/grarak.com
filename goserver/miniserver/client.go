package miniserver

import (
        "net/http"
        "io/ioutil"
)

type Client struct {
        Url, Method string
        Request     []byte
        Queries     map[string][]string
}

func newClient(request *http.Request) *Client {
        defer request.Body.Close()

        body, _ := ioutil.ReadAll(request.Body)

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

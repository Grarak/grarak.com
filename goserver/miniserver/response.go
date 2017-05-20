package miniserver

import (
	"net/http"
)

type Response struct {
	body, file, contentType, serverDescription string
	statusCode                                 int
}

func newResponseBody(body string) *Response {
	var response *Response = newResponse()
	response.body = body
	return response
}

func newResponseFile(file string) *Response {
	var response *Response = newResponse()
	response.file = file
	return response
}

func newResponse() *Response {
	return &Response{
		contentType:       "text/html",
		serverDescription: "Go MiniServer",
		statusCode:        http.StatusOK,
	}
}

func (response *Response) SetContentType(contentType string) {
	response.contentType = contentType
}

func (response *Response) SetStatusCode(code int) {
	response.statusCode = code
}

func (response *Response) SetServerDescription(description string) {
	response.serverDescription = description
}

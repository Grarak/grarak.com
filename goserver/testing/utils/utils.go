package utils

import (
        "net/http"
        "bytes"
        "io/ioutil"
        "encoding/base64"
)

func HttpGet(url string, buf []byte) ([]byte, error) {
        var request *http.Request
        var err error
        if buf == nil {
                request, err = http.NewRequest(http.MethodGet, url, nil)
        } else {
                request, err = http.NewRequest(http.MethodGet, url, bytes.NewBuffer(buf))
        }

        if err != nil {
                return nil, err
        }

        var client *http.Client = &http.Client{}
        response, err := client.Do(request)
        if err != nil {
                return nil, err
        }
        defer response.Body.Close()

        return ioutil.ReadAll(response.Body)
}

func HttpPost(url string, buf []byte) ([]byte, error) {
        var request *http.Request
        var err error
        if buf == nil {
                request, err = http.NewRequest(http.MethodPost, url, nil)
        } else {
                request, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(buf))
        }

        if err != nil {
                return nil, err
        }

        var client *http.Client = &http.Client{}
        response, err := client.Do(request)
        if err != nil {
                return nil, err
        }
        defer response.Body.Close()

        return ioutil.ReadAll(response.Body)
}

func Encode(text string) string {
        return base64.StdEncoding.EncodeToString([]byte(text))
}

func Decode(text string) ([]byte, error) {
        return base64.StdEncoding.DecodeString(text)
}

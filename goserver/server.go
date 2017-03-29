package main

import (
	"os"
	"strings"

	"fmt"

	"strconv"

	"./api"
	"./miniserver"
	"./utils"
)

const SERVER_TAG = "Server"

var supportedContentTypes [][]string = [][]string{
	{miniserver.ContentJson, ".json"},
	{miniserver.ContentJavascript, ".js"},
	{miniserver.ContentHtml, ".htm", ".html", ".shtml"},
	{miniserver.ContentCss, ".css"},
	{miniserver.ContentXIcon, ".ico"},
	{miniserver.ContentSVG, ".svg"},
}

var deviceData *api.DeviceData

func onConnect(client *miniserver.Client) *miniserver.Response {
	var response *miniserver.Response
	var url string = client.Url[1:]

	var urls []string = strings.Split(url, "/")

	var realPath string
	for i := range urls {
		if !utils.StringEmpty(urls[i]) &&
			(utils.DirExists(fmt.Sprintf("dist/%s", urls[i])) ||
				utils.FileExists(fmt.Sprintf("dist/%s", urls[i]))) {
			var realPathBuf []string = []string{"dist"}
			for x := i; x < len(urls); x++ {
				realPathBuf = append(realPathBuf, urls[x])
			}
			realPath = strings.Join(realPathBuf, "/")
			break
		}
	}

	if urls[0] != "/" && urls[0] != ".." && urls[0] != "serverdata" {
		if !utils.StringEmpty(realPath) {
			url = realPath
		}
		if utils.FileExists(url) {
			response = client.ResponseFile(url)
			response.SetContentType("text/plain")

		typesLoop:
			for _, contentType := range supportedContentTypes {
				for i := 1; i < len(contentType); i++ {
					var extension string = contentType[i]
					if len(url) > len(extension) &&
						url[len(url)-len(extension):] == extension {
						response.SetContentType(contentType[0])
						break typesLoop
					}
				}
			}
		}
	}

	if len(urls[len(urls)-1]) == 0 {
		urls = urls[:len(urls)-1]
	}
	if response == nil {
		if len(urls) >= 4 && urls[1] == "api" && len(realPath) == 0 {
			var resApi api.Api

			if urls[0] == "kerneladiutor" {
				resApi = api.NewKernelAdiutorApi(client,
					strings.Join(urls[3:], "/"), urls[2], deviceData)
			}

			if resApi != nil {
				response = resApi.GetResponse()
			}
		}

		if response == nil {
			response = client.ResponseFile("dist/index.html")
		}
	}
	return response
}

func main() {
	if _, err := os.Stat("serverdata"); err != nil {
		err = os.Mkdir("./serverdata", 0755)
		utils.Panic(err)
	}

	deviceData = api.NewDeviceData()
	if deviceData == nil {
		panic("Can't open devicedate db")
	}

	var port int = 3000
	if len(os.Args) == 2 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil {
			port = p
		}
	}

	utils.LogI(SERVER_TAG, fmt.Sprintf("Starting server at port %d\n", port))
	var server *miniserver.MiniServer = miniserver.NewServer(port)
	server.StartListening(onConnect)
}

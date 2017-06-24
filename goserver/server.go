package main

import (
	"os"
	"strings"

	"fmt"

	"strconv"

	"./api"
	"./api/kerneladiutor"
	"./api/mandy"
	"./api/jodirect"
	"./kerneladiutor"
	"./mandy"
	"./jodirect"
	"./miniserver"
	"./utils"
	"os/signal"
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

var kaDeviceData *kerneladiutor.DeviceData
var mandyUserData *mandy.UserData
var mandyStatus *mandy.MandyStatus
var joDirectData *jodirect.JoDirectData

func onConnect(client *miniserver.Client) *miniserver.Response {
	var response *miniserver.Response
	var url string = client.Url[1:]

	var urls []string = strings.Split(url, "/")

	var realPath string
	// Check if site requests an HTML file
	// They are all inside the dist directory
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

	// Make sure nobody is trying to access private data
	if urls[0] != "/" && urls[0] != ".." && urls[0] != "serverdata" {
		if !utils.StringEmpty(realPath) {
			// File belongs to the website inside the dist directory
			// Update the url accordingly
			url = realPath
		}
		if utils.FileExists(url) {
			response = client.ResponseFile(url)
			response.SetContentType("text/plain")

			// Make content type match the extension of the requested file
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
			var resApi api.Interface

			var path string = strings.Join(urls[3:], "/")
			var apiVersion string = urls[2]
			switch urls[0] {
			case "kerneladiutor":
				resApi = api_kerneladiutor.NewKernelAdiutorApi(
					client, path, apiVersion, kaDeviceData,
				)
				break
			case "mandy":
				resApi = api_mandy.NewMandyApi(
					client, path, apiVersion, mandyUserData,
					mandyStatus,
				)
				break
			case "jodirect":
				resApi = api_jodirect.NewJoDirectApi(client, path, apiVersion,
					joDirectData)
				break
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

func mkdirs(path string) {
	if utils.DirExists(path) {
		err := os.MkdirAll(path, 0755)
		utils.Panic(err)
	}
}

func showUsage() {
	fmt.Println("Usage:", os.Args[0], "<port> <mandy api key>")
}

func main() {
	if len(os.Args) != 3 {
		showUsage()
		return
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		showUsage()
		return
	}

	// Create necessary folders
	mkdirs(utils.KERNELADIUTOR)
	mkdirs(utils.MANDY)

	var server *miniserver.MiniServer = miniserver.NewServer(port)

	c := make(chan os.Signal, 1)
	cleanup := make(chan bool)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			utils.LogI(SERVER_TAG, fmt.Sprintf("Captured %s, killing...", sig))
			server.StopListening()
			mandyStatus.Kill()

			cleanup <- true
		}
	}()

	kaDeviceData = kerneladiutor.NewDeviceData()
	if kaDeviceData == nil {
		panic("Can't open kerneladiutor devicedata db")
	}

	mandyUserData = mandy.NewUserData()
	if mandyUserData == nil {
		panic("Can't open mandy userdata db")
	}

	mandyStatus = mandy.MandyInit(true, os.Args[2], mandyUserData)

	joDirectData = jodirect.NewJoDirectData()

	utils.LogI(SERVER_TAG, fmt.Sprintf("Starting server at port %d", port))
	go server.StartListening(onConnect)

	<-cleanup
}

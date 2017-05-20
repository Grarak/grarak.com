package api

import (
	"fmt"
	"strconv"

	"../utils"
	"encoding/json"
	"strings"
	"reflect"
	"time"
)

type KernelAdiutorTest struct{}

type deviceInfo struct {
	AndroidID      string    `json:"android_id"`
	AndroidVersion string    `json:"android_version"`
	KernelVersion  string    `json:"kernel_version"`
	AppVersion     string    `json:"app_version"`
	Board          string    `json:"board"`
	Model          string    `json:"model"`
	Vendor         string    `json:"vendor"`
	CpuInfo        string    `json:"cpuinfo"`
	Fingerprint    string    `json:"fingerprint"`
	Commands       []string  `json:"commands"`
	Times          []float64 `json:"times"`
	Cpu            float64   `json:"cpu"`
}

func (test KernelAdiutorTest) RunTest(args []string) {
	port, err := strconv.Atoi(args[0])
	if err != nil {
		utils.Failed(args[0], "is not a port")
		return
	}

	// Test port
	var url string = fmt.Sprintf("http://localhost:%d", port)
	buf, err := utils.HttpGet(url, nil)
	if err != nil {
		utils.Failed("Port", strconv.Itoa(port), "is not reachable")
		return
	}

	utils.Passed(fmt.Sprintf("%s returned:\n%s", url, string(buf)))

	var dInfo deviceInfo = deviceInfo{
		"0123456789abcdef",
		"7.1.1",
		"Linux version 4.4.16-abcdefgh",
		"1.0",
		"testboard",
		"Dummy",
		"Samtrung",
		utils.Encode("Generic cpu"),
		"grarak/testkey/123456/user",
		[]string{},
		[]float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
		123456789,
	}

	// Test device creation
	dInfo.deviceCreate(url)

	time.Sleep(time.Second)

	// Test device fetching
	dInfo.deviceGet(url)

	// Test device update
	dInfo.Board = "testboard2"
	dInfo.deviceCreate(url)

	dInfo.AndroidID = "0123456789abcdeg"
	dInfo.deviceCreate(url)

	dInfo.AndroidID = "0123456789abcdeh"
	dInfo.Cpu = 10
	dInfo.deviceCreate(url)
}

func (dInfo deviceInfo) deviceCreate(url string) {
	var urlDeviceCreate string = fmt.Sprintf("%s/kerneladiutor/api/v1/device/create", url)

	// Send invalid content
	_, err := utils.HttpPost(urlDeviceCreate, []byte("Random data"))
	if err != nil {
		utils.Failed("Create device failed")
		return
	}

	deviceJson, err := json.Marshal(dInfo)
	if err != nil {
		panic(err)
	}

	buf, err := utils.HttpPost(urlDeviceCreate, deviceJson)
	if err != nil || !strings.Contains(string(buf), "\"success\":true") {
		utils.Failed("Create device failed")
		return
	}

	utils.Passed("Create device")
}

func (dInfo deviceInfo) deviceGet(url string) {
	var urlDeviceGet string = fmt.Sprintf("%s/kerneladiutor/api/v1/device/get", url)

	buf, err := utils.HttpGet(urlDeviceGet, nil)

	var datas []map[string]interface{}
	err = json.Unmarshal(buf, &datas)
	if err != nil {
		utils.Failed(fmt.Sprintf("%s returned:\n%s ", urlDeviceGet, buf))
		return
	}

	var match bool = false
	deviceInfoField := reflect.TypeOf(dInfo)

	for _, device := range datas {
		decodeBuf, err := utils.Decode(device["id"].(string))
		if err != nil {
			utils.Failed("Can't decode ID:", device["id"])
			return
		}
		device["android_id"] = string(decodeBuf)

		for i := 0; i < deviceInfoField.NumField(); i++ {
			field := deviceInfoField.Field(0)
			dInfoValue := reflect.Indirect(reflect.ValueOf(dInfo)).FieldByName(field.Name)
			jsonName := field.Tag.Get("json")

			if reflect.DeepEqual(device[jsonName], dInfoValue) {
				break
			}
			match = true
		}
	}

	if !match {
		utils.Failed("Get returned wrong devices")
	}

	utils.Passed("Get device")
}

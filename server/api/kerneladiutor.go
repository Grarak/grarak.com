package api

import (
	"encoding/json"
	"net/http"
	"time"

	"fmt"

	"../miniserver"
	"../utils"
)

const KA_TAG = "kerneladiutor"

type KernelAdiutorApi struct {
	client      *miniserver.Client
	path        string
	version     string
	devicedata  *DeviceData
	deviceInfos map[string]DeviceInfo
}

func (kaAPi KernelAdiutorApi) GetResponse() *miniserver.Response {
	switch kaAPi.version {
	case "v1":
		return kaAPi.kernelAdiutorApiv1()
	default:
		return nil
	}
}

func NewKernelAdiutorApi(client *miniserver.Client,
	path, version string,
	dData *DeviceData,
	dInfos map[string]DeviceInfo) KernelAdiutorApi {
	return KernelAdiutorApi{
		client:      client,
		path:        path,
		version:     version,
		devicedata:  dData,
		deviceInfos: dInfos,
	}
}

func (kaAPi KernelAdiutorApi) kernelAdiutorApiv1() *miniserver.Response {
	var response *miniserver.Response

	switch kaAPi.path {
	case "device/create":
		if kaAPi.client.Method == http.MethodPost &&
			len(kaAPi.client.Request) > 0 {

			var data map[string]interface{}
			json.Unmarshal(kaAPi.client.Request, &data)

			var dInfo DeviceInfo = NewDeviceInfo(data)
			if dInfo.valid() {

				var updated bool = kaAPi.putDatabase(dInfo)
				if b, err := kaAPi.createStatus(true); err == nil {
					response = kaAPi.client.ResponseBody(string(b))
				}

				if updated {
					utils.LogI(KA_TAG, fmt.Sprintf("Updating device %s", dInfo.Model))
				} else {
					utils.LogI(KA_TAG, fmt.Sprintf("Inserting device %s", dInfo.Model))
				}
			}
		}

	}

	if response == nil {
		if b, err := kaAPi.createStatus(false); err == nil {
			response = kaAPi.client.ResponseBody(string(b))
		}
	}
	response.SetContentType(miniserver.ContentJson)

	return response
}

func (kaApi KernelAdiutorApi) createStatus(success bool) ([]byte, error) {
	var statusCode int = http.StatusOK
	if !success {
		statusCode = http.StatusNotFound
	}
	return json.Marshal(struct {
		Success bool   `json:"success"`
		Method  string `json:"method"`
		Request string `json:"request"`
		Version string `json:"version"`
		Status  int64  `json:"status"`
	}{success, kaApi.client.Method, kaApi.path,
		kaApi.version, int64(statusCode)})
}

type DeviceInfo struct {
	AndroidID      string    `json:"android_id"`
	AndroidVersion string    `json:"android_version"`
	KernelVersion  string    `json:"kernel_version"`
	Board          string    `json:"board"`
	Model          string    `json:"model"`
	Vendor         string    `json:"vendor"`
	Commands       []string  `json:"commands"`
	Times          []float64 `json:"times"`
	Cpu            float64   `json:"cpu"`
	Date           string    `json:"date"`
}

func NewDeviceInfo(data map[string]interface{}) DeviceInfo {
	var j utils.Json = utils.Json{data}

	var date string = j.GetString("date")
	if utils.StringEmpty(date) {
		date = time.Now().Format(time.RFC3339)
	}

	return DeviceInfo{
		AndroidID:      j.GetString("android_id"),
		AndroidVersion: j.GetString("android_version"),
		KernelVersion:  j.GetString("kernel_version"),
		Board:          j.GetString("board"),
		Model:          j.GetString("model"),
		Vendor:         j.GetString("vendor"),
		Commands:       j.GetStringArray("commands"),
		Times:          j.GetFloatArray("times"),
		Cpu:            j.GetFloat("cpu"),
		Date:           date,
	}
}

func (dInfo DeviceInfo) valid() bool {
	return !utils.StringEmpty(dInfo.AndroidID) &&
		!utils.StringEmpty(dInfo.AndroidVersion) &&
		!utils.StringEmpty(dInfo.KernelVersion) &&
		!utils.StringEmpty(dInfo.Board) &&
		!utils.StringEmpty(dInfo.Model) &&
		!utils.StringEmpty(dInfo.Vendor) &&
		dInfo.Commands != nil && len(dInfo.Commands) >= 10 &&
		dInfo.Times != nil && len(dInfo.Times) >= 15 &&
		dInfo.Cpu != 0
}

func (dInfo DeviceInfo) Json() ([]byte, error) {
	return json.Marshal(dInfo)
}

func (kaApi KernelAdiutorApi) putDatabase(dInfo DeviceInfo) bool {
	_, exists := kaApi.deviceInfos[dInfo.AndroidID]
	if exists {
		kaApi.devicedata.Update(dInfo, kaApi.deviceInfos)
	} else {
		kaApi.devicedata.Insert(dInfo, kaApi.deviceInfos)
	}
	return exists
}

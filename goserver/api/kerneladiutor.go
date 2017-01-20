package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"strconv"

	"../miniserver"
	"../utils"
)

const KA_TAG = "kerneladiutor"

type KernelAdiutorApi struct {
	client     *miniserver.Client
	path       string
	version    string
	devicedata *DeviceData
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
	dData *DeviceData) KernelAdiutorApi {
	return KernelAdiutorApi{
		client:     client,
		path:       path,
		version:    version,
		devicedata: dData,
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

			var dInfo *DeviceInfo = NewDeviceInfo(data, true)
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
	case "device/get":
		if kaAPi.client.Method == http.MethodGet {

			// Get all
			if page, pageok := kaAPi.client.Queries["page"]; (pageok && len(kaAPi.client.Queries) == 1) ||
				len(kaAPi.client.Queries) == 0 {

				var pageNumber int = 1
				if pageok {
					if num, err := strconv.Atoi(page[0]); err == nil {
						if num > 0 {
							pageNumber = num
						}
					}
				}

				responses := make([]DeviceInfo, 0)
				for i := (pageNumber - 1) * 10; i < pageNumber*10; i++ {
					if i < len(kaAPi.devicedata.sortedScores) {
						if value, ok := kaAPi.devicedata.infos[kaAPi.devicedata.sortedScores[i]]; ok {
							var info DeviceInfo = *value
							info.AndroidID = ""
							responses = append(responses, info)
						}
					}
				}
				if len(responses) > 0 {
					b, err := json.Marshal(responses)
					if err == nil {
						response = kaAPi.client.ResponseBody(string(b))
					}
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
	ID             string    `json:"id"`
	AndroidID      string    `json:"android_id,omitempty"`
	AndroidVersion string    `json:"android_version"`
	KernelVersion  string    `json:"kernel_version"`
	AppVersion     string    `json:"app_version"`
	Board          string    `json:"board"`
	Model          string    `json:"model"`
	Vendor         string    `json:"vendor"`
	CpuInfo        string    `json:"cpuinfo"`
	Commands       []string  `json:"commands"`
	Times          []float64 `json:"times"`
	Cpu            float64   `json:"cpu"`
	Date           string    `json:"date"`
	Score          float64   `json:"score"`
}

func NewDeviceInfo(data map[string]interface{}, post bool) *DeviceInfo {
	var j utils.Json = utils.Json{data}

	var dInfo *DeviceInfo = &DeviceInfo{
		ID:             j.GetString("id"),
		AndroidID:      j.GetString("android_id"),
		AndroidVersion: j.GetString("android_version"),
		KernelVersion:  j.GetString("kernel_version"),
		AppVersion:     j.GetString("app_version"),
		Board:          j.GetString("board"),
		Model:          j.GetString("model"),
		Vendor:         j.GetString("vendor"),
		Commands:       j.GetStringArray("commands"),
		Times:          j.GetFloatArray("times"),
		Cpu:            j.GetFloat("cpu"),
		Date:           j.GetString("date"),
		Score:          j.GetFloat("score"),
	}

	if post {
		if cpuinfoencoded := j.GetString("cpuinfo"); !utils.StringEmpty(cpuinfoencoded) {
			if cpuinfo, err := utils.Decode(cpuinfoencoded); err == nil {
				dInfo.CpuInfo = string(cpuinfo)
			}
		}
	} else {
		dInfo.CpuInfo = j.GetString("cpuinfo")
	}

	if dInfo.valid() {
		if utils.StringEmpty(dInfo.ID) {
			dInfo.ID = utils.Encode(dInfo.AndroidID)
		}
		if utils.StringEmpty(dInfo.Date) {
			dInfo.Date = time.Now().Format(time.RFC3339)
		}
		if dInfo.Score == 0 {
			dInfo.Score = utils.GetAverage(dInfo.Times)*1e9 - dInfo.Cpu
		}
	}

	return dInfo
}

func (dInfo DeviceInfo) valid() bool {
	return !utils.StringEmpty(dInfo.AndroidID) &&
		!utils.StringEmpty(dInfo.AndroidVersion) &&
		!utils.StringEmpty(dInfo.KernelVersion) &&
		!utils.StringEmpty(dInfo.AppVersion) &&
		!utils.StringEmpty(dInfo.Board) &&
		!utils.StringEmpty(dInfo.Model) && dInfo.Model != "unknown" &&
		!utils.StringEmpty(dInfo.Vendor) &&
		!utils.StringEmpty(dInfo.CpuInfo) &&
		dInfo.Commands != nil &&
		dInfo.Times != nil && len(dInfo.Times) >= 20 &&
		dInfo.Cpu != 0
}

func (dInfo DeviceInfo) Json() ([]byte, error) {
	return json.Marshal(dInfo)
}

func (kaApi KernelAdiutorApi) putDatabase(dInfo *DeviceInfo) bool {
	return kaApi.devicedata.Update(dInfo)
}

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

func (kaApi KernelAdiutorApi) deviceCreate() *miniserver.Response {

        var data map[string]interface{}
        err := json.Unmarshal(kaApi.client.Request, &data)
        if err != nil {
                return nil
        }

        var dInfo *DeviceInfo = NewDeviceInfo(data, true)
        if dInfo.valid() {

                var updated bool = kaApi.devicedata.Update(dInfo)
                if updated {
                        utils.LogI(KA_TAG, fmt.Sprintf("Updating device %s", dInfo.Model))
                } else {
                        utils.LogI(KA_TAG, fmt.Sprintf("Inserting device %s", dInfo.Model))
                }

                if b, err := kaApi.createStatus(true); err == nil {
                        return kaApi.client.ResponseBody(string(b))
                }
        }

        return nil
}

func (kaApi KernelAdiutorApi) deviceGet() *miniserver.Response {
        var pageNumber int = 1
        if pageQuery, pageQueryok := kaApi.client.Queries["page"]; pageQueryok {
                if num, err := strconv.Atoi(pageQuery[0]); err == nil {
                        if num > 0 {
                                pageNumber = num
                        }
                }
        }

        if id, idOk := kaApi.client.Queries["id"]; idOk {
                // Specific id
                realId, err := utils.Decode(id[0])
                if err == nil {
                        if value, valueOk := kaApi.devicedata.infos[string(realId)]; valueOk {
                                var info DeviceInfo = *value
                                info.AndroidID = ""
                                jsonBuf, err := json.Marshal(info)
                                if err == nil {
                                        return kaApi.client.ResponseBody(string(jsonBuf))
                                }
                        }
                }
        } else {
                // No specific id
                // Respond with list based on page
                responses := make([]DeviceInfo, 0)
                for i := (pageNumber - 1) * 10; i < pageNumber*10; i++ {
                        if i < len(kaApi.devicedata.sortedScores) {
                                if value, valueOk := kaApi.devicedata.infos[kaApi.devicedata.sortedScores[i]]; valueOk {
                                        var info DeviceInfo = *value
                                        info.AndroidID = ""
                                        responses = append(responses, info)
                                }
                        }
                }
                if len(responses) > 0 {
                        jsonBuf, err := json.Marshal(responses)
                        if err == nil {
                                return kaApi.client.ResponseBody(string(jsonBuf))
                        }
                }
        }

        return nil
}

func (kaApi KernelAdiutorApi) boardGet() *miniserver.Response {
        if id, idOk := kaApi.client.Queries["id"]; idOk {
                if devices, boardOk := kaApi.devicedata.board[id[0]]; boardOk {
                        var deviceIds []string = make([]string, 0)

                        for _, device := range devices {
                                deviceIds = append(deviceIds, kaApi.devicedata.infos[device].ID)
                        }

                        if len(deviceIds) > 0 {
                                buf, err := json.Marshal(deviceIds)
                                if err != nil {
                                        return nil
                                }

                                return kaApi.client.ResponseBody(string(buf))
                        }
                }
        } else {
                var boards []string = make([]string, 0)
                for board, boardList := range kaApi.devicedata.board {
                        if len(boardList) > 0 {
                                boards = append(boards, board)
                        }
                }

                if len(boards) > 0 {
                        buf, err := json.Marshal(boards)
                        if err != nil {
                                return nil
                        }

                        return kaApi.client.ResponseBody(string(buf))
                }
        }

        return nil
}

func (kaApi KernelAdiutorApi) kernelAdiutorApiv1() *miniserver.Response {
        var response *miniserver.Response
        var silentStatusCode bool = false

        if silentcodeQuery, silentcodeQueryok := kaApi.client.Queries["silent"]; silentcodeQueryok {
                silentStatusCode = silentcodeQuery[0] == "true"
        }

        switch kaApi.path {
        case "device/create":
                if kaApi.client.Method == http.MethodPost && len(kaApi.client.Request) > 0 {
                        response = kaApi.deviceCreate()
                }
        case "device/get":
                if kaApi.client.Method == http.MethodGet {
                        response = kaApi.deviceGet()
                }
        case "board/get":
                if kaApi.client.Method == http.MethodGet {
                        response = kaApi.boardGet()
                }
        }

        if response == nil {
                if b, err := kaApi.createStatus(false); err == nil {
                        response = kaApi.client.ResponseBody(string(b))
                        if !silentStatusCode {
                                response.SetStatusCode(http.StatusNotFound)
                        }
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
        Fingerprint    string    `json:"fingerprint"`
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
                Fingerprint:    j.GetString("fingerprint"),
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
                len(dInfo.AndroidID) >= 12 &&
                !utils.StringEmpty(dInfo.AndroidVersion) &&
                !utils.StringEmpty(dInfo.KernelVersion) &&
                !utils.StringEmpty(dInfo.AppVersion) &&
                !utils.StringEmpty(dInfo.Board) &&
                !utils.StringEmpty(dInfo.Model) && dInfo.Model != "unknown" &&
                !utils.StringEmpty(dInfo.Vendor) &&
                !utils.StringEmpty(dInfo.CpuInfo) &&
                !utils.StringEmpty(dInfo.Fingerprint) &&
                dInfo.Commands != nil &&
                dInfo.Times != nil && len(dInfo.Times) >= 20 &&
                dInfo.Cpu != 0
}

func (dInfo DeviceInfo) Json() ([]byte, error) {
        return json.Marshal(dInfo)
}

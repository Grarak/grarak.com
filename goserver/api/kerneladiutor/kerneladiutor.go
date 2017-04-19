package api_kerneladiutor

import (
        "encoding/json"
        "net/http"

        "strconv"

        "../../kerneladiutor"
        "../../miniserver"
        "../../utils"
)

type KernelAdiutorApi struct {
        client     *miniserver.Client
        path       string
        version    string
        devicedata *kerneladiutor.DeviceData
}

func (kaAPi KernelAdiutorApi) GetResponse() *miniserver.Response {
        switch kaAPi.version {
        case "v1":
                return kaAPi.kernelAdiutorApiv1()
        default:
                return nil
        }
}

func NewKernelAdiutorApi(
        client *miniserver.Client,
        path,
        version string,
        dData *kerneladiutor.DeviceData,
) KernelAdiutorApi {
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

        var dInfo *kerneladiutor.DeviceInfo = kerneladiutor.NewDeviceInfo(data, true)
        if dInfo.Valid() {

                kaApi.devicedata.UpdateDevice(dInfo)
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
                        if value, valueOk := kaApi.devicedata.Infos[string(realId)]; valueOk {
                                var info kerneladiutor.DeviceInfo = *value
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
                responses := make([]kerneladiutor.DeviceInfo, 0)
                for i := (pageNumber - 1) * 10; i < pageNumber*10; i++ {
                        if i < len(kaApi.devicedata.SortedScores) {
                                if value, valueOk := kaApi.devicedata.Infos[kaApi.devicedata.SortedScores[i]]; valueOk {
                                        var info kerneladiutor.DeviceInfo = *value
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
                if devices, boardOk := kaApi.devicedata.Board[id[0]]; boardOk {
                        var deviceIds []string = make([]string, 0)

                        for _, device := range devices {
                                deviceIds = append(deviceIds, kaApi.devicedata.Infos[device].ID)
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
                for board, boardList := range kaApi.devicedata.Board {
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

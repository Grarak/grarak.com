package kerneladiutor

import (
        "time"
        "encoding/json"
        "../utils"
)

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

        if dInfo.Valid() {
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

func (dInfo DeviceInfo) Valid() bool {
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

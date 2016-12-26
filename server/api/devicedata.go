package api

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"

	"sync"

	"time"

	"../utils"
)

type DeviceData struct {
	db           *sql.DB
	infos        map[string]*DeviceInfo
	sortedScores []string

	newdevices []*DeviceInfo
	mutex      sync.Mutex
}

func NewDeviceData() *DeviceData {
	db, err := sql.Open("sqlite3", "./serverdata/device.db")
	if err != nil {
		return nil
	}

	_, err = db.Exec("create table if not exists devices(id, json)")
	if err != nil {
		return nil
	}

	var dData *DeviceData = &DeviceData{
		db:         db,
		newdevices: make([]*DeviceInfo, 0),
	}
	dData.infos = dData.getDevices()

	dData.sortedScores = make([]string, 0)
	for k := range dData.infos {
		dData.sortedScores = append(dData.sortedScores, k)
	}
	dData.sortScores()

	go func() {
		for {
			time.Sleep(time.Second / 3)

			if len(dData.newdevices) > 0 {
				var id string = dData.newdevices[0].AndroidID
				if _, ok := dData.infos[id]; !ok {
					dData.sortedScores = append(dData.sortedScores, id)
					dData.infos[id] = dData.newdevices[0]
				}
				dData.sortScores()
				dData.newdevices = dData.newdevices[1:]
			}
		}
	}()

	return dData
}

func (dData *DeviceData) updateSortedScores(dInfo *DeviceInfo, insert bool) {
	dData.newdevices = append(dData.newdevices, dInfo)
}

func (dData *DeviceData) Update(dInfo *DeviceInfo) bool {
	var updated bool

	j, err := dInfo.Json()
	utils.Panic(err)

	trans, err := dData.db.Begin()
	utils.Panic(err)

	var stmt *sql.Stmt
	if _, ok := dData.infos[dInfo.AndroidID]; ok {
		updated = true

		stmt, err = trans.Prepare("update devices set json = ? where id = ?")
		utils.Panic(err)

		_, err = stmt.Exec(string(j), dInfo.AndroidID)
		utils.Panic(err)
	} else {
		updated = false

		stmt, err = trans.Prepare("insert into devices(id, json) values(?,?)")
		utils.Panic(err)

		_, err = stmt.Exec(dInfo.AndroidID, string(j))
		utils.Panic(err)
	}
	defer stmt.Close()

	err = trans.Commit()
	utils.Panic(err)

	dData.updateSortedScores(dInfo, !updated)
	return updated
}

func (dData *DeviceData) getDevices() map[string]*DeviceInfo {
	deviceInfos := make(map[string]*DeviceInfo)

	query, err := dData.db.Query("select json from devices")
	utils.Panic(err)
	defer query.Close()

	for query.Next() {
		var j string
		err := query.Scan(&j)
		utils.Panic(err)

		var data map[string]interface{}
		err = json.Unmarshal([]byte(j), &data)
		utils.Panic(err)

		if dInfo := NewDeviceInfo(data); dInfo.valid() {
			deviceInfos[dInfo.AndroidID] = dInfo
		}
	}
	err = query.Err()
	utils.Panic(err)

	return deviceInfos
}

func (dData *DeviceData) Close() error {
	return dData.db.Close()
}

func (dData *DeviceData) _sortScores(low, high int) {
	if low >= high {
		return
	}

	var list []string = dData.sortedScores
	var data map[string]*DeviceInfo = dData.infos

	var middle int = low + (high-low)/2
	var pivot float64 = data[list[middle]].Score

	var i, j int = low, high
	for i <= j {
		for data[list[i]].Score < pivot {
			i++
		}
		for data[list[j]].Score > pivot {
			j--
		}

		if i <= j {
			list[i], list[j] = list[j], list[i]
			i++
			j--
		}

		dData._sortScores(low, j)
		dData._sortScores(i, high)
	}
}

// Use Quicksort algorithm to sort device scores
func (dData *DeviceData) sortScores() {
	dData._sortScores(0, len(dData.sortedScores)-1)

	utils.ReverseStringArray(dData.sortedScores)
}

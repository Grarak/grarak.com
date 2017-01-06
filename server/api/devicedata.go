package api

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"

	"time"

	"../utils"
)

type DeviceData struct {
	db           *sql.DB
	infos        map[string]*DeviceInfo
	sortedScores []string

	newdevices []*DeviceInfo
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
	sortScores(dData.sortedScores, dData.infos)

	go func() {
		for {
			time.Sleep(time.Second / 3)

			if len(dData.newdevices) > 0 {
				var id string = dData.newdevices[0].AndroidID
				var tmpList []string = make([]string, len(dData.sortedScores))
				copy(tmpList, dData.sortedScores)

				if _, ok := dData.infos[id]; !ok {
					tmpList = append(dData.sortedScores, id)
					dData.infos[id] = dData.newdevices[0]
				}

				sortScores(tmpList, dData.infos)
				dData.sortedScores = tmpList

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

func _sortScores(list []string, data map[string]*DeviceInfo, low, high int) {
	if low >= high {
		return
	}

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

		_sortScores(list, data, low, j)
		_sortScores(list, data, i, high)
	}
}

// Use Quicksort algorithm to sort device scores
func sortScores(list []string, data map[string]*DeviceInfo) {
	_sortScores(list, data, 0, len(list)-1)

	utils.ReverseStringArray(list)
}

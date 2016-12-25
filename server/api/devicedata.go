package api

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"

	"sync"

	"../utils"
)

type DeviceData struct {
	db           *sql.DB
	infos        map[string]*DeviceInfo
	sortedScores []string
	mutex        sync.Mutex
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
		db: db,
	}
	dData.infos = dData.getDevices()

	dData.sortedScores = make([]string, 0)
	for k := range dData.infos {
		dData.sortedScores = append(dData.sortedScores, k)
	}
	dData.sortScores()

	return dData
}

func (dData *DeviceData) updateSortedScores(dInfo *DeviceInfo, insert bool) {

	dData.infos[dInfo.AndroidID] = dInfo

	dData.mutex.Lock()
	defer dData.mutex.Unlock()

	if insert {
		dData.sortedScores = append(dData.sortedScores, dInfo.AndroidID)
	}
	dData.sortScores()
}

func (dData *DeviceData) Update(dInfo *DeviceInfo) bool {
	var updated bool

	j, err := dInfo.Json()
	utils.Panic(err)

	trans, err := dData.db.Begin()
	utils.Panic(err)

	var stmt *sql.Stmt
	defer stmt.Close()

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

	err = trans.Commit()
	utils.Panic(err)

	dData.updateSortedScores(dInfo, updated)
	return updated
}

func (dData *DeviceData) getDevices() map[string]*DeviceInfo {
	deviceInfos := make(map[string]*DeviceInfo)

	query, err := dData.db.Query("select id, json from devices")
	utils.Panic(err)
	defer query.Close()

	for query.Next() {
		var id string
		var j string
		err := query.Scan(&id, &j)
		utils.Panic(err)

		var data map[string]interface{}
		err = json.Unmarshal([]byte(j), &data)
		utils.Panic(err)

		deviceInfos[id] = NewDeviceInfo(data)
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

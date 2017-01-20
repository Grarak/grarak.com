package api

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/mattn/go-sqlite3"

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

	for k := range dData.infos {
		dData.sortedScores = append(dData.sortedScores, k)
	}

	utils.SimpleSort(dData.sortedScores, dData.getMinMaxDeterminator(dData.sortedScores))

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

				utils.SimpleSort(tmpList, dData.getMinMaxDeterminator(tmpList))
				dData.sortedScores = tmpList

				dData.newdevices = dData.newdevices[1:]
			}
		}
	}()

	return dData
}

func (dData *DeviceData) getMinMaxDeterminator(array []string) func(i, j int) bool {
	return func(i, j int) bool {
		return dData.infos[array[i]].Score > dData.infos[array[j]].Score
	}
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

		if dInfo := NewDeviceInfo(data, false); dInfo.valid() {
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

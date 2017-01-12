package api

import (
	"database/sql"
	"encoding/json"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"../utils"
)

type scoreSorter struct {
	list  []string
	infos map[string]*DeviceInfo
}

func (sorter scoreSorter) Len() int {
	return len(sorter.list)
}

func (sorter scoreSorter) Swap(i, j int) {
	sorter.list[i], sorter.list[j] = sorter.list[j], sorter.list[i]
}

func (sorter scoreSorter) Less(i, j int) bool {
	return sorter.infos[sorter.list[i]].Score > sorter.infos[sorter.list[j]].Score
}

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
	sort.Sort(scoreSorter{dData.sortedScores, dData.infos})

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

				sort.Sort(scoreSorter{tmpList, dData.infos})
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

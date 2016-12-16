package api

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"

	"../utils"
)

type DeviceData struct {
	db *sql.DB
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

	return &DeviceData{db}
}

func (dData *DeviceData) Insert(dInfo DeviceInfo, dInfos map[string]DeviceInfo) {
	j, err := dInfo.Json()
	utils.Panic(err)

	trans, err := dData.db.Begin()
	utils.Panic(err)

	stmt, err := trans.Prepare("insert into devices(id, json) values(?,?)")
	utils.Panic(err)
	defer stmt.Close()

	_, err = stmt.Exec(dInfo.AndroidID, string(j))
	utils.Panic(err)

	err = trans.Commit()
	utils.Panic(err)

	dInfos[dInfo.AndroidID] = dInfo
}

func (dData *DeviceData) Update(dInfo DeviceInfo, dInfos map[string]DeviceInfo) {
	j, err := dInfo.Json()
	utils.Panic(err)

	trans, err := dData.db.Begin()
	utils.Panic(err)

	stmt, err := trans.Prepare("update devices set json = ? where id = ?")
	utils.Panic(err)
	defer stmt.Close()

	_, err = stmt.Exec(string(j), dInfo.AndroidID)
	utils.Panic(err)

	err = trans.Commit()
	utils.Panic(err)

	dInfos[dInfo.AndroidID] = dInfo
}

func (dData *DeviceData) GetDevices() map[string]DeviceInfo {
	deviceInfos := make(map[string]DeviceInfo)

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

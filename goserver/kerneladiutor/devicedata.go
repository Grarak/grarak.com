package kerneladiutor

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"

	"fmt"

	"../utils"
	"time"
)

type DeviceData struct {
	db *sql.DB

	Infos        map[string]*DeviceInfo
	SortedScores []string

	Board map[string][]string

	newDevices []*DeviceInfo
}

func NewDeviceData() *DeviceData {
	db, err := sql.Open("sqlite3", utils.KERNELADIUTOR+"/device.db")
	if err != nil {
		return nil
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS devices(id, json)")
	if err != nil {
		return nil
	}

	dData := &DeviceData{
		db,
		getDevices(db),
		make([]string, 0),
		make(map[string][]string, 0),
		make([]*DeviceInfo, 0),
	}

	for key := range dData.Infos {
		dData.SortedScores = append(dData.SortedScores, key)
	}

	utils.SimpleSort(dData.SortedScores, func(i, j int) bool {
		return dData.Infos[dData.SortedScores[i]].Score > dData.Infos[dData.SortedScores[j]].Score
	})

	// Extract all existing board names
	// Use already sorted list, so we don't have to sort it afterwards
	for _, id := range dData.SortedScores {
		var device *DeviceInfo = dData.Infos[id]

		var boardDevices []string = dData.Board[device.Board]
		if boardDevices == nil {
			boardDevices = make([]string, 0)
		}
		dData.Board[device.Board] = append(boardDevices, device.AndroidID)
	}

	// Start a go routine
	// So we sort new devices one after one
	go func() {
		for {
			if len(dData.newDevices) > 0 {
				newDevice := dData.newDevices[0]

				// Check if the board of the updated device got changed
				// If yes then remove it from the old map
				if oldDevice, ok := dData.Infos[newDevice.AndroidID]; ok &&
					newDevice.Board != oldDevice.Board {
					if boardDevices, boardDevicesOk := dData.Board[oldDevice.Board];
						boardDevicesOk && len(boardDevices) > 0 {
						if index, err := dData.findDevice(oldDevice, boardDevices); err == nil {
							dData.Board[oldDevice.Board] =
								utils.RemoveFromSlice(boardDevices, index)
						}
					}
				}

				// Insert to global sortedlist
				newSortedList := dData.insertDevice(newDevice, dData.SortedScores)

				// Insert to board sortedlist
				newBoardList := dData.insertDevice(newDevice, dData.Board[newDevice.Board])

				dData.Infos[newDevice.AndroidID] = newDevice
				dData.SortedScores = newSortedList
				dData.Board[newDevice.Board] = newBoardList
				dData.newDevices = dData.newDevices[1:]
			}

			time.Sleep(time.Second / 3)
		}
	}()

	return dData
}

func (dData DeviceData) insertDevice(newDevice *DeviceInfo, sortedList []string) []string {
	var index int = 0
	var err error

	// Remove the old position in sorted list
	if oldDevice, ok := dData.Infos[newDevice.AndroidID]; ok {
		index, err = dData.findDevice(oldDevice, sortedList)
		if err == nil {
			sortedList = utils.RemoveFromSlice(sortedList, index)
		}
	}

	index, err = dData.findDevice(newDevice, sortedList)
	if err == nil {
		panic(fmt.Sprintf("%s shouldn't be in the list", newDevice.AndroidID))
	}

	if len(sortedList) == 0 || newDevice.Score >= dData.Infos[sortedList[index]].Score {
		sortedList = utils.InsertToSlice(newDevice.AndroidID, sortedList, index)
	} else {
		sortedList = utils.InsertToSlice(newDevice.AndroidID, sortedList, index+1)
	}

	return sortedList
}

func (dData DeviceData) _findDevice(searchDevice *DeviceInfo, sortedList []string, min, max int) (int, error) {
	if len(sortedList) == 0 {
		return 0, utils.GenericError(fmt.Sprintf("Couldn't find %s", searchDevice.AndroidID))
	}

	length := max - min
	middle := length / 2
	index := middle + min

	middleDevice := dData.Infos[sortedList[index]]

	// Make sure if id actually exists
	// otherwise it will end in an endless loop
	if min >= max {
		if sortedList[min] == searchDevice.AndroidID {
			return min, nil
		}
		return min, utils.GenericError(fmt.Sprintf("Couldn't find %s", searchDevice.AndroidID))
	}

	if searchDevice.Score > middleDevice.Score {
		return dData._findDevice(searchDevice, sortedList, min, index-1)
	} else if searchDevice.Score < middleDevice.Score {
		return dData._findDevice(searchDevice, sortedList, index+1, max)
	}

	if searchDevice.AndroidID == sortedList[index] {
		return index, nil
	}
	return index, utils.GenericError(fmt.Sprintf("Couldn't find %s", searchDevice.AndroidID))
}

func (dData DeviceData) findDevice(newDevice *DeviceInfo, sortedList []string) (int, error) {
	return dData._findDevice(newDevice, sortedList, 0, len(sortedList)-1)
}

func (dData *DeviceData) UpdateDevice(dInfo *DeviceInfo) {

	j, err := dInfo.Json()
	utils.Panic(err)

	trans, err := dData.db.Begin()
	utils.Panic(err)

	var stmt *sql.Stmt
	if _, ok := dData.Infos[dInfo.AndroidID]; ok {
		// Device already registered
		// Update its informations

		stmt, err = trans.Prepare("update devices set json = ? where id = ?")
		utils.Panic(err)

		_, err = stmt.Exec(string(j), dInfo.AndroidID)
		utils.Panic(err)
	} else {
		// New device incoming
		// Insert its informations

		stmt, err = trans.Prepare("insert into devices(id, json) values(?,?)")
		utils.Panic(err)

		_, err = stmt.Exec(dInfo.AndroidID, string(j))
		utils.Panic(err)
	}

	err = trans.Commit()
	utils.Panic(err)
	defer stmt.Close()

	// Insert/Update sorted list
	// Go routine above will handle the sorting
	dData.newDevices = append(dData.newDevices, dInfo)
}

func getDevices(db *sql.DB) map[string]*DeviceInfo {

	deviceInfos := make(map[string]*DeviceInfo)
	var invalidDevices []string

	query, err := db.Query("SELECT json FROM devices")
	utils.Panic(err)
	defer query.Close()

	for query.Next() {
		var j string
		err := query.Scan(&j)
		utils.Panic(err)

		var data map[string]interface{}
		err = json.Unmarshal([]byte(j), &data)
		utils.Panic(err)

		if dInfo := NewDeviceInfo(data, false); dInfo.Valid() {
			deviceInfos[dInfo.AndroidID] = dInfo
		} else {
			// Collect invalid ids for later deletion
			invalidDevices = append(invalidDevices, dInfo.AndroidID)
		}
	}
	err = query.Err()
	utils.Panic(err)

	// Delete invalid ids from database
	deleteStmt, err := db.Prepare("DELETE FROM devices WHERE id = ?")
	utils.Panic(err)
	defer deleteStmt.Close()

	for _, invalidIds := range invalidDevices {
		utils.LogI(KA_TAG, fmt.Sprintf("%s invalid. Deleting", invalidIds))
		_, err = deleteStmt.Exec(invalidIds)
		utils.Panic(err)
	}

	return deviceInfos
}

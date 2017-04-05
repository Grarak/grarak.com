package api

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

        infos        map[string]*DeviceInfo
        sortedScores []string

        board map[string][]string

        newDevices []*DeviceInfo
}

func NewDeviceData() *DeviceData {
        db, err := sql.Open("sqlite3", "./serverdata/device.db")
        if err != nil {
                return nil
        }

        _, err = db.Exec("CREATE TABLE IF NOT EXISTS devices(id, json)")
        if err != nil {
                return nil
        }

        var dData *DeviceData = &DeviceData{
                db,
                getDevices(db),
                make([]string, 0),
                make(map[string][]string, 0),
                make([]*DeviceInfo, 0),
        }

        for key := range dData.infos {
                dData.sortedScores = append(dData.sortedScores, key)
        }

        utils.SimpleSort(dData.sortedScores, dData.getMinMaxDeterminator(dData.sortedScores))

        // Extract all existing board names
        // Use already sorted list, so we don't have to sort it afterwards
        for _, id := range dData.sortedScores {
                var device *DeviceInfo = dData.infos[id]

                var boardDevices []string = dData.board[device.Board]
                if boardDevices == nil {
                        boardDevices = make([]string, 0)
                }
                dData.board[device.Board] = append(boardDevices, device.AndroidID)
        }

        // Start a go routine
        // So we sort new devices one after one
        go func() {
                for {
                        if len(dData.newDevices) > 0 {
                                var newDevice *DeviceInfo = dData.newDevices[0]

                                // Check if the board of the updated device got changed
                                // If yes then remove it from the old map
                                if oldDevice, ok := dData.infos[newDevice.AndroidID]; ok &&
                                        newDevice.Board != oldDevice.Board {
                                        if boardDevices, boardDevicesOk := dData.board[oldDevice.Board];
                                                boardDevicesOk && len(boardDevices) > 0 {
                                                if index, err := dData.findDevice(oldDevice, boardDevices); err == nil {
                                                        utils.RemoveFromSlice(&boardDevices, index)
                                                        dData.board[oldDevice.Board] = boardDevices
                                                }
                                        }
                                }

                                // Insert to global sortedlist
                                var bufSlice []string = dData.sortedScores
                                dData.insertDevice(newDevice, &bufSlice)
                                dData.sortedScores = bufSlice

                                // Insert to board sortedlist
                                var boardSlice []string = dData.board[newDevice.Board]
                                bufSlice = make([]string, len(boardSlice))
                                copy(bufSlice, boardSlice)
                                dData.insertDevice(newDevice, &bufSlice)
                                dData.board[newDevice.Board] = bufSlice

                                dData.infos[newDevice.AndroidID] = newDevice
                                dData.newDevices = dData.newDevices[1:]
                        }

                        time.Sleep(time.Second / 3)
                }
        }()

        return dData
}

func (dData DeviceData) insertDevice(newDevice *DeviceInfo, sortedList *[]string) {
        var index int = 0
        var err error

        // Remove the old position in sorted list
        if oldDevice, ok := dData.infos[newDevice.AndroidID]; ok {
                index, err = dData.findDevice(oldDevice, *sortedList)
                if err == nil {
                        utils.RemoveFromSlice(sortedList, index)
                        index, err = dData.findDevice(newDevice, *sortedList)
                        if err == nil {
                                panic(fmt.Sprintf("%s is still there", newDevice.AndroidID))
                        }
                }
        }

        if len(*sortedList) == 0 || newDevice.Score >= dData.infos[(*sortedList)[index]].Score {
                utils.InsertToSlice(newDevice.AndroidID, sortedList, index)
        } else {
                utils.InsertToSlice(newDevice.AndroidID, sortedList, index+1)
        }
}

func (dData DeviceData) _findDevice(searchDevice *DeviceInfo, sortedList []string, min, max int) (int, error) {
        if len(sortedList) == 0 {
                return 0, utils.GenericError(fmt.Sprintf("Couldn't find %s", searchDevice.AndroidID))
        }

        var length int = max - min
        var middle int = length / 2
        var index int = middle + min

        var middleDevice *DeviceInfo = dData.infos[sortedList[index]]

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

func (dData *DeviceData) getMinMaxDeterminator(array []string) func(i, j int) bool {
        return func(i, j int) bool {
                return dData.infos[array[i]].Score > dData.infos[array[j]].Score
        }
}

func (dData *DeviceData) Update(dInfo *DeviceInfo) bool {
        var updated bool

        j, err := dInfo.Json()
        utils.Panic(err)

        trans, err := dData.db.Begin()
        utils.Panic(err)

        var stmt *sql.Stmt

        if _, ok := dData.infos[dInfo.AndroidID]; ok {
                // Device already registered
                // Update its informations
                updated = true

                stmt, err = trans.Prepare("update devices set json = ? where id = ?")
                utils.Panic(err)

                _, err = stmt.Exec(string(j), dInfo.AndroidID)
                utils.Panic(err)
        } else {
                // New device incoming
                // Insert its informations
                updated = false

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

        return updated
}

func getDevices(db *sql.DB) map[string]*DeviceInfo {

        deviceInfos := make(map[string]*DeviceInfo)
        invalidDdevices := make([]string, 0)

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

                if dInfo := NewDeviceInfo(data, false); dInfo.valid() {
                        deviceInfos[dInfo.AndroidID] = dInfo
                } else {
                        // Collect invalid ids for later deletion
                        invalidDdevices = append(invalidDdevices, dInfo.AndroidID)
                }
        }
        err = query.Err()
        utils.Panic(err)

        // Delete invalid ids from database
        deleteStmt, err := db.Prepare("DELETE FROM devices WHERE id = ?")
        utils.Panic(err)
        defer deleteStmt.Close()

        for _, invalidIds := range invalidDdevices {
                utils.LogI(KA_TAG, fmt.Sprintf("%s invalid. Deleting", invalidIds))
                _, err = deleteStmt.Exec(invalidIds)
                utils.Panic(err)
        }

        return deviceInfos
}

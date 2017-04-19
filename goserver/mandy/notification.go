package mandy

import (
        "net/http"
        "encoding/json"
        "bytes"
        "../utils"
        "../miniserver"
        "io/ioutil"
        "os"
        "time"
)

type MandyNotification int

const NOTIFICATION_VERIFIED MandyNotification = 0
const NOTIFICATION_NEW_TAG_FOUND MandyNotification = 1
const NOTIFICATION_MERGEABLE MandyNotification = 2
const NOTIFICATION_MERGING MandyNotification = 3
const NOTIFICATION_MERGED MandyNotification = 4
const NOTIFICATION_REVERTING MandyNotification = 5
const NOTIFICATION_REVERTED MandyNotification = 6
const NOTIFICATION_SUBMITTING MandyNotification = 7
const NOTIFICATION_SUBMITTED MandyNotification = 8

type Data struct {
        Code         int `json:"code"`
        BelongToUser string `json:"belongto,omitempty"`
        Data         interface{} `json:"data"`
        Date         string `json:"date"`
}

type Notification struct {
        firebaseApiKey string
        userdata       *UserData

        Activities []Data
}

func NewNotification(firebaseApiKey string, userdata *UserData) *Notification {
        data := make([]Data, 0)
        if _, err := os.Stat(utils.MANDY + "/notification.json"); err == nil {
                if buf, err := ioutil.ReadFile(utils.MANDY + "/notification.json"); err == nil {
                        err := json.Unmarshal(buf, &data)
                        utils.Panic(err)
                }
        }

        return &Notification{firebaseApiKey, userdata, data}
}

func (notification *Notification) saveData() {
        buf, err := json.Marshal(notification.Activities)
        utils.Panic(err)

        err = ioutil.WriteFile(utils.MANDY+"/notification.json", buf, 0644)
        utils.Panic(err)
}

func (notification *Notification) Notify(notificationCode MandyNotification, data interface{}, users ...*User) {
        noti := struct {
                Operation string `json:"operation"`
                To        string `json:"to"`
                Data      Data`json:"data"`
        }{
                Operation: "create",
                Data: Data{
                        Code: int(notificationCode),
                        Data: data,
                        Date: time.Now().Format("2006-01-02 15:04:05"),
                },
        }

        if notificationCode != NOTIFICATION_VERIFIED {
                if len(notification.Activities) >= 50 {
                        notification.Activities = notification.Activities[:50]
                }
                notification.Activities = append([]Data{noti.Data}, notification.Activities...)
                notification.saveData()
        }
        for _, user := range users {
                if user.FirebaseKey == nil || len(user.FirebaseKey) == 0 {
                        continue
                }

                noti.Data.BelongToUser = user.Name
                for index, key := range user.FirebaseKey {
                        noti.To = key
                        buf, err := json.Marshal(noti)
                        if err != nil {
                                utils.LogE(MANDY_TAG, "Can't send notification to "+user.Name)
                                continue
                        }

                        request, err := http.NewRequest(http.MethodPost,
                                "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(buf))
                        request.Header.Set("Content-Type", miniserver.ContentJson)
                        request.Header.Set("Authorization", "key="+notification.firebaseApiKey)
                        if err != nil {
                                utils.LogE(MANDY_TAG, "Can't send notification to "+user.Name)
                                continue
                        }

                        var client *http.Client = &http.Client{}
                        response, err := client.Do(request)
                        if err != nil {
                                utils.LogE(MANDY_TAG, "Can't send notification to "+user.Name)
                                continue
                        }

                        buf, err = ioutil.ReadAll(response.Body)
                        if err != nil {
                                utils.LogE(MANDY_TAG, "Can't send notification to "+user.Name)
                                continue
                        }
                        response.Body.Close()

                        type Return struct {
                                Success int `json:"success"`
                        }
                        var ret Return
                        err = json.Unmarshal(buf, &ret)
                        if err != nil || ret.Success == 0 {
                                user.FirebaseKey = utils.RemoveFromSlice(user.FirebaseKey, index)
                                notification.userdata.UpdateUser(user)
                        }
                }
        }
}

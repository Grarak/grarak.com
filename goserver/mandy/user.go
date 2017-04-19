package mandy

import (
        "encoding/json"
        "../utils"
)

type User struct {
        ApiToken    string `json:"apitoken,omitempty"`
        Name        string `json:"name"`
        Password    string `json:"password,omitempty"`
        Salt        string `json:"salt,omitempty"`
        Hash        string `json:"hash,omitempty"`
        Admin       bool   `json:"admin"`
        Moderator   bool   `json:"moderator"`
        Verified    bool   `json:"verified"`
        FirebaseKey []string `json:"firebasekey,omitempty"`
}

func (user User) ToJson() string {
        buf, err := json.Marshal(user)
        utils.Panic(err)

        return string(buf)
}

func (user *User) Strip() User {
        copyuser := *user
        copyuser.Salt = ""
        copyuser.Hash = ""
        copyuser.Password = ""
        copyuser.FirebaseKey = nil
        return copyuser
}

func (user *User) HardStrip() User {
        copyuser := user.Strip()
        copyuser.ApiToken = ""
        return copyuser
}

func NewUser(j []byte) (*User, error) {
        var user User
        err := json.Unmarshal(j, &user)
        return &user, err
}

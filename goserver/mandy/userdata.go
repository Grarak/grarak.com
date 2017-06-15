package mandy

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"../utils"
	"golang.org/x/crypto/pbkdf2"
	"crypto/sha256"
	"crypto/rand"
	"reflect"
	"strings"
)

type UserData struct {
	db *sql.DB

	apiTokens     map[string]*User
	users         map[string]string
	verifiedUsers map[string]string
}

type UserDataErr string

func (e UserDataErr) Error() string {
	return string(e)
}

func NewUserData() *UserData {
	db, err := sql.Open("sqlite3", utils.MANDY+"/user.db")
	if err != nil {
		return nil
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users(apiToken, json)")
	if err != nil {
		return nil
	}

	userData := &UserData{
		db,
		getApiTokens(db),
		make(map[string]string),
		make(map[string]string),
	}

	// Put usernames into a map
	// So you can quickly determinate if it's taken
	for apiToken, user := range userData.apiTokens {
		userData.users[strings.ToLower(user.Name)] = apiToken
		if user.Verified {
			userData.verifiedUsers[strings.ToLower(user.Name)] = apiToken
		}
	}

	return userData
}

func generateSalt() []byte {
	buf := make([]byte, 32)
	_, err := rand.Read(buf)
	utils.Panic(err)

	return buf
}

func (userdata *UserData) generateApiToken() string {
	token := utils.ToURLBase64(generateSalt())
	if _, ok := userdata.apiTokens[token]; ok {
		return userdata.generateApiToken()
	}

	return token
}

func hashPassword(password, salt []byte) []byte {
	return pbkdf2.Key(password, salt, 4096, sha256.Size, sha256.New)
}

func (userdata *UserData) InsertUser(user *User) (User, MandyErrorCode) {
	if len(user.Name) <= 3 {
		return user.Strip(), CODE_USERNAME_SHORT
	}
	if len(user.Password) <= 4 {
		return user.Strip(), CODE_PASSWORD_SHORT
	}
	if strings.Contains(user.Name, "\n") {
		return user.Strip(), cODE_USERNAME_INVALID
	}
	if _, ok := userdata.users[strings.ToLower(user.Name)]; ok {
		return user.Strip(), CODE_USERNAME_TAKEN
	}

	trans, err := userdata.db.Begin()
	utils.Panic(err)

	stmt, err := trans.Prepare("insert into users(apiToken, json) values(?,?)")
	utils.Panic(err)
	defer stmt.Close()

	// Hash password
	salt := generateSalt()
	hash := hashPassword([]byte(user.Password), salt)
	user.Password = ""
	user.Hash = utils.ToURLBase64(hash)
	user.Salt = utils.ToURLBase64(salt)

	// Generate api token
	newToken := userdata.generateApiToken()
	user.ApiToken = newToken

	// Automatically making the user admin if he is the first user
	// Also verify him
	if len(userdata.apiTokens) == 0 {
		user.Admin = true
		user.Verified = true

		userdata.verifiedUsers[strings.ToLower(user.Name)] = user.ApiToken
	}

	// Write to db
	_, err = stmt.Exec(newToken, user.ToJson())
	utils.Panic(err)

	utils.Panic(trans.Commit())

	userdata.apiTokens[newToken] = user
	userdata.users[strings.ToLower(user.Name)] = newToken

	return user.Strip(), CODE_NO_ERROR
}

func (userdata *UserData) UpdateUser(user *User) {
	trans, err := userdata.db.Begin()
	utils.Panic(err)

	stmt, err := trans.Prepare("update users set json = ? where apiToken = ?")
	utils.Panic(err)
	defer stmt.Close()

	_, err = stmt.Exec(user.ToJson(), user.ApiToken)
	utils.Panic(err)

	utils.Panic(trans.Commit())
}

func (userdata *UserData) GetUserWithPassword(user *User) (User, MandyErrorCode) {
	if apitoken, ok := userdata.users[strings.ToLower(user.Name)]; ok {
		actualuser := userdata.apiTokens[apitoken]
		salt, err := utils.FromURLBase64(actualuser.Salt)
		utils.Panic(err)
		hash := hashPassword([]byte(user.Password), salt)
		actualhash, err := utils.FromURLBase64(actualuser.Hash)
		utils.Panic(err)

		if reflect.DeepEqual(actualhash, hash) {
			user.Password = ""
			return actualuser.Strip(), CODE_NO_ERROR
		}
	}

	return *user, CODE_USERNAME_PASSWORD_INVALID
}

func (userdata *UserData) GetUsers() []*User {
	var users []*User
	for _, api := range userdata.users {
		users = append(users, userdata.apiTokens[api])
	}
	return users
}

func (userdata *UserData) GetVerifiedUsers() []*User {
	var users []*User
	for _, api := range userdata.verifiedUsers {
		users = append(users, userdata.apiTokens[api])
	}
	return users
}

func (userdata *UserData) Verify(requester *User, user string, verified bool) (*User, error) {
	if !requester.Admin {
		return nil, UserDataErr(requester.Name + " is not authorized to do this")
	}
	if api, ok := userdata.users[strings.ToLower(user)]; ok {
		user := userdata.apiTokens[api]
		if user.Admin {
			return nil, UserDataErr("You can't remove verification from yourself")
		}
		user.Verified = verified
		if verified {
			userdata.verifiedUsers[strings.ToLower(user.Name)] = user.ApiToken
		} else {
			delete(userdata.verifiedUsers, strings.ToLower(user.Name))
		}

		userdata.UpdateUser(user)
		return user, nil
	}

	return nil, UserDataErr("Can't find user " + user)
}

func (userdata *UserData) Moderator(requester *User, user string, moderator bool) (*User, error) {
	if !requester.Admin {
		return nil, UserDataErr(requester.Name + " is not authorized to do this")
	}
	if api, ok := userdata.users[strings.ToLower(user)]; ok {
		user := userdata.apiTokens[api]
		if user.Admin {
			return nil, UserDataErr("User " + user.Name + " is already admin")
		}
		if user.Verified {
			user.Moderator = moderator
			userdata.UpdateUser(user)
			return user, nil
		}
		return nil, UserDataErr("User " + user.Name + " is not verified yet")
	}

	return nil, UserDataErr("Can't find user " + user)
}

func (userdata *UserData) Remove(requester *User, user string) error {
	if !requester.Admin && !requester.Moderator {
		return UserDataErr(requester.Name + " is not authorized to do this")
	}
	if api, ok := userdata.users[strings.ToLower(user)]; ok {
		if userdata.apiTokens[api].Verified {
			return UserDataErr("Can't remove verified user " + user)
		}
		delete(userdata.apiTokens, api)
		delete(userdata.users, strings.ToLower(user))

		trans, err := userdata.db.Begin()
		utils.Panic(err)

		stmt, err := trans.Prepare("delete from users where apiToken = ?")
		utils.Panic(err)
		defer stmt.Close()

		_, err = stmt.Exec(api)
		utils.Panic(err)

		utils.Panic(trans.Commit())

		return nil
	}

	return UserDataErr("Can't find user " + user)
}

func (userdata *UserData) UpdateFirebaseKey(apiToken string, keys []string) error {
	user := userdata.FindUserByApi(apiToken)
	if user != nil {
		for _, key := range keys {
			if !utils.SliceContains(key, user.FirebaseKey) {
				user.FirebaseKey = append(user.FirebaseKey, key)
				userdata.UpdateUser(user)
			}
		}
		return nil
	}

	return UserDataErr("User does not exist")
}

func (userdata *UserData) FindUserByApi(api string) *User {
	if user, ok := userdata.apiTokens[api]; ok {
		return user
	}
	return nil
}

func getApiTokens(db *sql.DB) map[string]*User {
	apiTokens := make(map[string]*User)

	query, err := db.Query("SELECT json FROM users")
	utils.Panic(err)
	defer query.Close()

	for query.Next() {
		var j string
		err = query.Scan(&j)
		utils.Panic(err)

		if user, err := NewUser([]byte(j)); err == nil {
			apiTokens[user.ApiToken] = user
		} else {
			panic("User is not readable " + j)
		}
	}

	return apiTokens
}

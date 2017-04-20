package api_mandy

import (
        "../../miniserver"
        "../../mandy"
        "../../utils"
        "net/http"
        "encoding/json"
)

const MANDY_TAG = "Mandy"

type MandyApi struct {
        client      *miniserver.Client
        path        string
        version     string
        userdata    *mandy.UserData
        mandyStatus *mandy.MandyStatus
}

func (mandyApi MandyApi) GetResponse() *miniserver.Response {
        switch mandyApi.version {
        case "v1":
                return mandyApi.mandyApiv1()
        }
        return nil
}

func NewMandyApi(
        client *miniserver.Client,
        path,
        version string,
        userdata *mandy.UserData,
        mandyStatus *mandy.MandyStatus,
) MandyApi {
        return MandyApi{
                client,
                path,
                version,
                userdata,
                mandyStatus,
        }
}

func (mandyApi MandyApi) createResponse(code mandy.MandyErrorCode) *miniserver.Response {
        response := struct {
                Success    bool `json:"success"`
                StatusCode int `json:"statuscode"`
                Path       string `json:"path"`
        }{StatusCode: int(code), Path: mandyApi.path}
        response.Success = code == mandy.CODE_NO_ERROR

        buf, err := json.Marshal(response)
        utils.Panic(err)

        resBody := mandyApi.client.ResponseBody(string(buf))
        if !response.Success {
                resBody.SetStatusCode(http.StatusNotFound)
        }
        return resBody
}

func (mandyApi MandyApi) getApiFromQuery() string {
        if apiQuery, ok := mandyApi.client.Queries["key"]; ok && len(apiQuery) == 1 {
                return apiQuery[0]
        }
        return ""
}

func (mandyApi MandyApi) validateVerifiedApi() (*mandy.User, mandy.MandyErrorCode) {
        if apiToken := mandyApi.getApiFromQuery(); !utils.StringEmpty(apiToken) {
                user := mandyApi.userdata.FindUserByApi(apiToken)
                if user != nil && user.Verified {
                        return user, mandy.CODE_NO_ERROR
                }

                return nil, mandy.CODE_API_INVALID
        }

        return nil, mandy.CODE_UNKNOWN_ERROR
}

func (mandyApi MandyApi) accountSignup() *miniserver.Response {
        if user, err := mandy.NewUser(mandyApi.client.Request); err == nil {
                response, statuscode := mandyApi.userdata.InsertUser(user)
                if statuscode != mandy.CODE_NO_ERROR {
                        return mandyApi.createResponse(statuscode)
                }

                utils.LogI(MANDY_TAG, "New user "+user.Name)
                return mandyApi.client.ResponseBody(response.ToJson())
        }

        return mandyApi.createResponse(mandy.CODE_UNKNOWN_ERROR)
}

func (mandyApi MandyApi) accountSignin() *miniserver.Response {
        if user, err := mandy.NewUser(mandyApi.client.Request); err == nil {
                user, statuscode := mandyApi.userdata.GetUserWithPassword(user)
                if statuscode != mandy.CODE_NO_ERROR {
                        return mandyApi.createResponse(statuscode)
                }

                utils.LogI(MANDY_TAG, user.Name+" signed in")
                return mandyApi.client.ResponseBody(user.ToJson())
        }

        return mandyApi.createResponse(mandy.CODE_UNKNOWN_ERROR)
}

func (mandyApi MandyApi) accountFirebaseKey() *miniserver.Response {
        if apiToken := mandyApi.getApiFromQuery(); !utils.StringEmpty(apiToken) {
                if userKey, err := mandy.NewUser(mandyApi.client.Request); err == nil {
                        err := mandyApi.userdata.UpdateFirebaseKey(apiToken, userKey.FirebaseKey[0])
                        if err != nil {
                                return mandyApi.createResponse(mandy.CODE_API_INVALID)
                        }

                        return mandyApi.createResponse(mandy.CODE_NO_ERROR)
                }
        }

        return mandyApi.createResponse(mandy.CODE_UNKNOWN_ERROR)
}

func (mandyApi MandyApi) accountGet() *miniserver.Response {
        user, code := mandyApi.validateVerifiedApi()
        if code == mandy.CODE_NO_ERROR {
                return mandyApi.client.ResponseBody(user.Strip().ToJson())
        }
        return mandyApi.createResponse(code)
}

func (mandyApi MandyApi) userGet() *miniserver.Response {
        _, code := mandyApi.validateVerifiedApi()
        if code == mandy.CODE_NO_ERROR {
                var users []mandy.User
                for _, user := range mandyApi.userdata.GetUsers() {
                        users = append(users, user.HardStrip())
                }

                buf, err := json.Marshal(users)
                if err != nil {
                        return mandyApi.createResponse(mandy.CODE_GET_USERS_FAILED)
                }
                return mandyApi.client.ResponseBody(string(buf))
        }
        return mandyApi.createResponse(code)
}

func (mandyApi MandyApi) userGetVerified() *miniserver.Response {
        _, code := mandyApi.validateVerifiedApi()
        if code == mandy.CODE_NO_ERROR {
                var users []mandy.User
                for _, user := range mandyApi.userdata.GetVerifiedUsers() {
                        users = append(users, user.HardStrip())
                }

                buf, err := json.Marshal(users)
                if err != nil {
                        return mandyApi.createResponse(mandy.CODE_GET_USERS_FAILED)
                }
                return mandyApi.client.ResponseBody(string(buf))
        }
        return mandyApi.createResponse(code)
}

func (mandyApi MandyApi) userVerify() *miniserver.Response {
        requester, code := mandyApi.validateVerifiedApi()
        if code == mandy.CODE_NO_ERROR {
                tmpUser, err := mandy.NewUser(mandyApi.client.Request)
                if err != nil {
                        return mandyApi.createResponse(mandy.CODE_SET_VERIFICATION_FAILED)
                }

                user, err := mandyApi.userdata.Verify(requester, tmpUser.Name, tmpUser.Verified)
                if err != nil {
                        return mandyApi.createResponse(mandy.CODE_SET_VERIFICATION_FAILED)
                }

                mandyApi.mandyStatus.Notification.Notify(mandy.NOTIFICATION_VERIFIED, tmpUser, user)
                return mandyApi.userGet()
        }
        return mandyApi.createResponse(code)
}

func (mandyApi MandyApi) userModerator() *miniserver.Response {
        requester, code := mandyApi.validateVerifiedApi()
        if code == mandy.CODE_NO_ERROR {
                user, err := mandy.NewUser(mandyApi.client.Request)
                if err != nil {
                        return mandyApi.createResponse(mandy.CODE_SET_MODERATION_FAILED)
                }

                _, err = mandyApi.userdata.Moderator(requester, user.Name, user.Moderator)
                if err != nil {
                        return mandyApi.createResponse(mandy.CODE_SET_MODERATION_FAILED)
                }

                return mandyApi.userGetVerified()
        }
        return mandyApi.createResponse(code)
}

func (mandyApi MandyApi) userRemove() *miniserver.Response {
        requester, code := mandyApi.validateVerifiedApi()
        if code == mandy.CODE_NO_ERROR {
                user, err := mandy.NewUser(mandyApi.client.Request)
                if err != nil {
                        return mandyApi.createResponse(mandy.CODE_REMOVE_USER_FAILED)
                }

                err = mandyApi.userdata.Remove(requester, user.Name)
                if err != nil {
                        return mandyApi.createResponse(mandy.CODE_REMOVE_USER_FAILED)
                }

                return mandyApi.userGet()
        }
        return mandyApi.createResponse(code)
}

func (mandyApi MandyApi) statusGet() *miniserver.Response {
        _, code := mandyApi.validateVerifiedApi()
        if code == mandy.CODE_NO_ERROR {
                buf, err := json.Marshal(mandyApi.mandyStatus)
                utils.Panic(err)

                return mandyApi.client.ResponseBody(string(buf))
        }
        return mandyApi.createResponse(code)
}

func (mandyApi MandyApi) repoConflicted() *miniserver.Response {
        _, code := mandyApi.validateVerifiedApi()
        if code == mandy.CODE_NO_ERROR {
                var project mandy.AospaProject
                err := json.Unmarshal(mandyApi.client.Request, &project)
                if err != nil {
                        return mandyApi.createResponse(mandy.CODE_SET_CONFLICTION_FAILED)
                }

                mandyApi.mandyStatus.SetConflicted(project.Name, project.Conflicted)
                return mandyApi.statusGet()
        }
        return mandyApi.createResponse(code)
}

func (mandyApi MandyApi) repoMerge() *miniserver.Response {
        user, code := mandyApi.validateVerifiedApi()
        if code == mandy.CODE_NO_ERROR {
                err := mandyApi.mandyStatus.StartMerging(user)
                if err != nil {
                        return mandyApi.createResponse(mandy.CODE_MERGE_FAILED)
                }

                return mandyApi.statusGet()
        }
        return mandyApi.createResponse(code)
}

func (mandyApi MandyApi) repoRevert() *miniserver.Response {
        user, code := mandyApi.validateVerifiedApi()
        if code == mandy.CODE_NO_ERROR {
                err := mandyApi.mandyStatus.Revert(user)
                if err != nil {
                        return mandyApi.createResponse(mandy.CODE_REVERT_FAILED)
                }

                return mandyApi.statusGet()
        }
        return mandyApi.createResponse(code)
}

func (mandyApi MandyApi) repoSubmit() *miniserver.Response {
        user, code := mandyApi.validateVerifiedApi()
        if code == mandy.CODE_NO_ERROR {
                err := mandyApi.mandyStatus.Submit(user)
                if err != nil {
                        return mandyApi.createResponse(mandy.CODE_SUBMIT_FAILED)
                }

                return mandyApi.statusGet()
        }
        return mandyApi.createResponse(code)
}

func (mandyApi MandyApi) notificationActivities() *miniserver.Response {
        _, code := mandyApi.validateVerifiedApi()
        if code == mandy.CODE_NO_ERROR {
                buf, err := json.Marshal(mandyApi.mandyStatus.Notification.Activities)
                if err != nil {
                        return mandyApi.createResponse(mandy.CODE_GET_NOTIFICATION_ACTIVITIES_FAILED)
                }

                return mandyApi.client.ResponseBody(string(buf))
        }
        return mandyApi.createResponse(code)
}

func (mandyApi MandyApi) scriptMerge() *miniserver.Response {
        script := "#!/bin/bash\n\nset -e\n\nBRANCH=" + mandy.MANDY_BRANCH + "\n"

        for _, project := range mandyApi.mandyStatus.AospaProjects {
                script += "\n"
                script += "git -C " + project.Path + " remote add github " + mandy.GITHUB_HTTP + "/" + project.Name + "\n"
                script += "git -C " + project.Path + " fetch github\n"
                script += "git -C " + project.Path + " reset --hard github/$BRANCH\n"
                script += "echo \"" + project.Name + " checked out\"\n"
        }

        response := mandyApi.client.ResponseBody(script)
        response.SetContentType(miniserver.ContentText)
        return response
}

func (mandyApi MandyApi) scriptDeleteBranch() *miniserver.Response {
        script := "#!/bin/bash\n\nset -e\n\nBRANCH=" + mandy.MANDY_BRANCH + "\n"

        for _, project := range mandyApi.mandyStatus.AospaProjects {
                script += "\n"
                script += "git -C " + project.Path + " push git@github.com:" + project.Name + ".git $BRANCH --delete\n"
                script += "echo \"" + project.Name + " deleted\"\n"
        }

        response := mandyApi.client.ResponseBody(script)
        response.SetContentType(miniserver.ContentText)
        return response
}

func (mandyApi MandyApi) mandyApiv1() *miniserver.Response {
        var response *miniserver.Response

        switch mandyApi.path {
        case "account/signup":
                if mandyApi.client.Method == http.MethodPost && len(mandyApi.client.Request) > 0 {
                        response = mandyApi.accountSignup()
                }
                break
        case "account/signin":
                if mandyApi.client.Method == http.MethodPost && len(mandyApi.client.Request) > 0 {
                        response = mandyApi.accountSignin()
                }
                break
        case "account/firebasekey":
                if mandyApi.client.Method == http.MethodPost && len(mandyApi.client.Request) > 0 {
                        response = mandyApi.accountSignin()
                }
                break
        case "account/get":
                if mandyApi.client.Method == http.MethodGet {
                        response = mandyApi.accountGet()
                }
                break

        case "user/get":
                if mandyApi.client.Method == http.MethodGet {
                        response = mandyApi.userGet()
                }
                break
        case "user/getverified":
                if mandyApi.client.Method == http.MethodGet {
                        response = mandyApi.userGetVerified()
                }
                break
        case "user/verify":
                if mandyApi.client.Method == http.MethodPost && len(mandyApi.client.Request) > 0 {
                        response = mandyApi.userVerify()
                }
                break
        case "user/moderator":
                if mandyApi.client.Method == http.MethodPost && len(mandyApi.client.Request) > 0 {
                        response = mandyApi.userModerator()
                }
                break
        case "user/remove":
                if mandyApi.client.Method == http.MethodPost && len(mandyApi.client.Request) > 0 {
                        response = mandyApi.userRemove()
                }
                break

        case "status/get":
                if mandyApi.client.Method == http.MethodGet {
                        response = mandyApi.statusGet()
                }
                break

        case "repo/conflicted":
                if mandyApi.client.Method == http.MethodPost && len(mandyApi.client.Request) > 0 {
                        response = mandyApi.repoConflicted()
                }
                break
        case "repo/merge":
                if mandyApi.client.Method == http.MethodPost {
                        response = mandyApi.repoMerge()
                }
                break
        case "repo/revert":
                if mandyApi.client.Method == http.MethodPost {
                        response = mandyApi.repoRevert()
                }
                break
        case "repo/submit":
                if mandyApi.client.Method == http.MethodPost {
                        response = mandyApi.repoSubmit()
                }
                break

        case "notification/activities":
                if mandyApi.client.Method == http.MethodGet {
                        response = mandyApi.notificationActivities()
                }
                break

        case "script/merge":
                if mandyApi.client.Method == http.MethodGet {
                        response = mandyApi.scriptMerge()
                }
                break
        case "script/deletebranch":
                if mandyApi.client.Method == http.MethodGet {
                        response = mandyApi.scriptDeleteBranch()
                }
                break
        }

        if response == nil {
                response = mandyApi.createResponse(mandy.CODE_UNKNOWN_ERROR)
        }

        return response
}

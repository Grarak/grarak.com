package mandy

import (
	"../utils"
	"../utils/git"
	"fmt"
	"io/ioutil"
	"time"
	"strings"
	"strconv"
	"encoding/json"
)

const MANDY_TAG = "Mandy"

const GITHUB_HTTP = "git://github.com"

const AOSPA_MANIFEST_NAME = "manifest"
const AOSPA_MANIFEST_URL = GITHUB_HTTP + "/AOSPA/manifest.git"
const MANIFEST_DEFAULT = "default.xml"
const MANIFEST_AOSPA = "manifests/aospa.xml"

const AOSPA_BRANCH = "nougat-mr2"
const MANDY_BRANCH = AOSPA_BRANCH + "-bot"

const MANDY_USERNAME = "PAMergebot"
const MANDY_NAME = "Mergebot"
const MANDY_EMAIL = "mergebot@aospa.co"

const GERRIT_URL = "ssh://" + MANDY_USERNAME + "@gerrit.aospa.co:29418"

// Blacklist for debugging
// We don't want to download all repos
var projectsBlackList = []string{
}

// Whitelist for debugging
// As long it's empty we will accept any repos
var projectsWhiteList = []string{
}

type AospaProject struct {
	Path       string `json:"-"`
	project    Project `json:"-"`
	revision   string `json:"-"`
	remote     Remote `json:"-"`
	cafProject Project `json:"-"`
	cafRemote  Remote `json:"-"`
	git        git.Git `json:"-"`

	Name       string `json:"name"`
	LatestTag  string `json:"latesttag"`
	Conflicted bool `json:"conflicted"`
}

type MandyStatus struct {
	ManifestTag         string `json:"manifesttag"`
	ManifestTagSplitted []string `json:"-"`

	LatestTag     string `json:"latesttag"`
	AospaProjects []*AospaProject `json:"projects,omitempty"`

	Mergeable bool `json:"mergeable"`
	Merging   bool `json:"merging"`
	Merged    bool `json:"merged"`
	Merger    User `json:"merger"`

	Submittable bool `json:"submittable"`
	Submitting  bool `json:"submitting"`
	Submitted   bool `json:"submitted"`
	Submitter   User `json:"submitter"`

	Reverting bool `json:"reverting"`
	Reverter  User `json:"reverter"`

	Notification   *Notification `json:"-"`
	firebaseApiKey string `json:"-"`
	killed         bool `json:"-"`
	manifestGit    *git.Git `json:"-"`
}

type MandyErr string

func (e MandyErr) Error() string {
	return string(e)
}

var mandyStatus *MandyStatus

func repoAccepted(repo string) bool {
	return !utils.SliceContains(repo, projectsBlackList) &&
		(len(projectsWhiteList) == 0 || utils.SliceContains(repo, projectsWhiteList))
}

func newMandyGit(path, url string) git.Git {
	return git.NewGit("mandy", path, url)
}

func buildGitUrl(project Project, remote Remote) string {
	var remoteURL string
	if remote.Fetch == ".." {
		remoteURL = GITHUB_HTTP
	} else {
		remoteURL = remote.Fetch
	}
	return remoteURL + "/" + project.Name
}

func saveMandyStatus() {
	buf, err := json.Marshal(mandyStatus)
	utils.Panic(err)

	err = ioutil.WriteFile(utils.MANDY+"/status.json", buf, 0644)
	utils.Panic(err)
}

func getMandyStatus() MandyStatus {
	buf, err := ioutil.ReadFile(utils.MANDY + "/status.json")
	if err != nil {
		return MandyStatus{}
	}

	var mandyStatus MandyStatus
	err = json.Unmarshal(buf, &mandyStatus)
	utils.Panic(err)

	return mandyStatus
}

func (mandyStatus *MandyStatus) Kill() {
	if mandyStatus != nil {
		mandyStatus.killed = true
		for _, aospaProject := range mandyStatus.AospaProjects {
			aospaProject.git.Kill()
		}
		mandyStatus.manifestGit.Kill()
	}
}

func (mandyStatus *MandyStatus) Strip() MandyStatus {
	copyMandyStatus := *mandyStatus
	copyMandyStatus.AospaProjects = nil
	return copyMandyStatus
}

func (mandyStatus *MandyStatus) Submit(user *User) error {
	if !mandyStatus.Merged || mandyStatus.Submitted || mandyStatus.Submitting || mandyStatus.Reverting {
		return MandyErr("Can't submit merge")
	}
	if !user.Moderator && !user.Admin {
		return MandyErr(user.Name + " is not authorized to do this")
	}

	mandyStatus.Submitting = true
	mandyStatus.Submitter = user.Strip()
	mandyStatus.Merged = false
	saveMandyStatus()

	mandyStatus.Notification.Notify(NOTIFICATION_SUBMITTING, mandyStatus.Strip(),
		mandyStatus.Notification.userdata.GetVerifiedUsers()...)

	return nil
}

func startSubmitting() {
	for _, aospaProject := range mandyStatus.AospaProjects {
		if aospaProject.Conflicted {
			mandyStatus.Submittable = false
			mandyStatus.Submitting = false
			mandyStatus.Submitter = User{}
			saveMandyStatus()
			return
		}
	}

	utils.LogI(MANDY_TAG, "Start Submitting, requested by "+mandyStatus.Submitter.Name)
	successful := true
	for _, aospaProject := range mandyStatus.AospaProjects {
		// Update the actual repo
		err := aospaProject.git.Pull("origin", MANDY_BRANCH)

		if err != nil {
			utils.LogE(MANDY_TAG, "Couldn't pull "+aospaProject.Name+", skipping submission")
			aospaProject.Conflicted = true
			successful = false
			continue
		}

		status, err := aospaProject.git.MergeBranch("origin", aospaProject.revision)
		if status != 0 || err != nil {
			utils.LogE(MANDY_TAG, "Couldn't merge with origin "+aospaProject.Name+", skipping submission")
			aospaProject.Conflicted = true
			successful = false
			continue
		}
	}

	if successful {
		for _, aospaProject := range mandyStatus.AospaProjects {
			aospaProject.git.Push("gerrit", "HEAD:refs/heads/"+aospaProject.revision, false)
		}
		mandyStatus.Submitted = true
	} else {
		mandyStatus.Merged = true
	}
	mandyStatus.Submittable = false
	mandyStatus.Submitting = false
	mandyStatus.Submitter = User{}
	saveMandyStatus()

	if mandyStatus.Submitted {
		utils.LogI(MANDY_TAG, "Submitting "+mandyStatus.LatestTag+" finished")

		mandyStatus.Notification.Notify(NOTIFICATION_SUBMITTED, mandyStatus.Strip(),
			mandyStatus.Notification.userdata.GetVerifiedUsers()...)
	}
}

func (mandyStatus *MandyStatus) Revert(user *User) error {
	if !mandyStatus.Merged || mandyStatus.Submitted || mandyStatus.Submitting || mandyStatus.Reverting {
		return MandyErr("Can't revert merge")
	}

	mandyStatus.Reverter = user.Strip()
	mandyStatus.Reverting = true
	mandyStatus.Mergeable = false
	mandyStatus.Merged = false
	saveMandyStatus()

	mandyStatus.Notification.Notify(NOTIFICATION_REVERTING, mandyStatus.Strip(),
		mandyStatus.Notification.userdata.GetVerifiedUsers()...)

	return nil
}

func startReverting() {
	utils.LogI(MANDY_TAG, "Start Reverting, requested by "+mandyStatus.Reverter.Name)

	for _, aospaProject := range mandyStatus.AospaProjects {
		err := aospaProject.git.Pull("origin", aospaProject.revision)
		if err != nil {
			utils.LogE(MANDY_TAG, "Couldn't pull "+aospaProject.Name+", skipping revert")
		} else {
			aospaProject.git.Push("gerrit", "HEAD:refs/heads/"+MANDY_BRANCH, true)
		}
	}

	mandyStatus.Reverting = false
	mandyStatus.Reverter = User{}
	saveMandyStatus()

	utils.LogI(MANDY_TAG, "Reverting "+mandyStatus.LatestTag+" finished")

	mandyStatus.Notification.Notify(NOTIFICATION_REVERTED, mandyStatus.Strip(),
		mandyStatus.Notification.userdata.GetVerifiedUsers()...)
}

func (mandyStatus *MandyStatus) StartMerging(user *User) error {
	if mandyStatus.Submitted || mandyStatus.Submitting || mandyStatus.Merged || mandyStatus.Merging || mandyStatus.Reverting {
		return MandyErr("Can't start merge")
	}

	mandyStatus.Merger = user.Strip()
	mandyStatus.Merging = true
	saveMandyStatus()

	mandyStatus.Notification.Notify(NOTIFICATION_MERGING, mandyStatus.Strip(),
		mandyStatus.Notification.userdata.GetVerifiedUsers()...)

	return nil
}

func startMerging() {
	for _, aospaProject := range mandyStatus.AospaProjects {
		if mandyStatus.LatestTag != aospaProject.LatestTag {
			mandyStatus.Merging = false
			mandyStatus.Mergeable = false
			saveMandyStatus()
			return
		}
	}

	utils.LogI(MANDY_TAG, "Start merging, requested by "+mandyStatus.Merger.Name)

	submittable := true
	for _, aospaProject := range mandyStatus.AospaProjects {
		// Update the actual repo
		err := aospaProject.git.Pull("origin", aospaProject.revision)

		if err != nil {
			utils.LogE(MANDY_TAG, "Couldn't pull "+aospaProject.Name+", skipping merge")
			aospaProject.Conflicted = true
		} else {
			// Start merging
			// The returning status code will tell if it was successful
			mergeStatus, err := aospaProject.git.MergeTag(mandyStatus.LatestTag)
			aospaProject.Conflicted = mergeStatus != 0 || err != nil
			if aospaProject.Conflicted {
				submittable = false
			}

			if aospaProject.Conflicted {
				utils.LogI(MANDY_TAG, "Failed to merge "+aospaProject.Name)
			} else {
				utils.LogI(MANDY_TAG, "Successfully merged "+aospaProject.Name)
			}
		}

		aospaProject.git.Push("gerrit", "HEAD:refs/heads/"+MANDY_BRANCH, true)

		err = aospaProject.git.Clean()
		if err != nil {
			utils.LogE(MANDY_TAG, err.Error())
		}
	}
	mandyStatus.Mergeable = false
	mandyStatus.Merging = false
	mandyStatus.Merger = User{}
	mandyStatus.Merged = true
	mandyStatus.Submittable = submittable
	saveMandyStatus()

	utils.LogI(MANDY_TAG, "Merging "+mandyStatus.LatestTag+" finished")
	mandyStatus.Notification.Notify(NOTIFICATION_MERGED, mandyStatus.Strip(),
		mandyStatus.Notification.userdata.GetVerifiedUsers()...)
}

func (mandyStatus *MandyStatus) SetConflicted(name string, conflicted bool) {
	submittable := true
	for _, aospaProject := range mandyStatus.AospaProjects {
		if aospaProject.Name == name {
			aospaProject.Conflicted = conflicted
		} else if !conflicted && aospaProject.Conflicted && submittable {
			submittable = false
		}
	}

	if conflicted {
		mandyStatus.Submittable = false
	} else if submittable {
		mandyStatus.Submittable = submittable
	}
	saveMandyStatus()
}

func trackCaf() {
	var tagEquivalents = []struct {
		alias1, alias2 string
	}{
		{"5.5", "5.6"},
	}

trackingLoop:
	for {

		if !mandyStatus.Merged && !mandyStatus.Submitted && !mandyStatus.Reverting {
			for _, aospaProject := range mandyStatus.AospaProjects {
				if !aospaProject.git.Valid() {
					continue
				}

				// Fetch caf
				utils.LogI(MANDY_TAG, "Fetching "+aospaProject.Name)
				err := aospaProject.git.Fetch("caf")
				if err != nil {
					break trackingLoop
				}

				var latestTag string = mandyStatus.ManifestTag
				var maxScore int = 0
				tags, err := aospaProject.git.GetTags()
				if err != nil {
					break trackingLoop
				}
				for _, tag := range tags {
					score := compareTag(mandyStatus.ManifestTagSplitted, splitTag(tag))
					if score >= 400 {

						// Some tags are equivalent
						// Since caf doesn't update them at the same time
						// This is need, since some people are impatient.
						// FUCK YOU ALEX
						for _, alias := range tagEquivalents {
							var aliasTag string
							if strings.Contains(tag, alias.alias1) {
								aliasTag = strings.Replace(tag, alias.alias1, alias.alias2, 1)
							} else if strings.Contains(tag, alias.alias2) {
								aliasTag = strings.Replace(tag, alias.alias2, alias.alias1, 1)
							}

							newScore := compareTag(mandyStatus.ManifestTagSplitted, splitTag(aliasTag))

							// Tags with the same score mean that all of them are old
							// or match the current tag
							// Tag with single highest score means we found a new usable tag
							insertScore := func(score int, tag string) {
								if score >= 525 && score > maxScore {
									maxScore = score
									latestTag = tag
								}
							}
							if newScore > score {
								insertScore(newScore, tag)
							} else {
								insertScore(score, tag)
							}
						}
					}
				}

				if latestTag != mandyStatus.ManifestTag {
					if !mandyStatus.Mergeable && !mandyStatus.Merged && !mandyStatus.Submitting {
						var mergable bool = true
						// Check if other repos have the tag as well
						for _, aospaProject := range mandyStatus.AospaProjects {
							if latestTag != aospaProject.LatestTag {
								mergable = false
								break
							}
						}

						if mergable {
							utils.LogI(MANDY_TAG, latestTag+" is now mergeable")
							mandyStatus.Mergeable = true
							mandyStatus.Notification.Notify(NOTIFICATION_MERGEABLE,
								mandyStatus.Strip(),
								mandyStatus.Notification.userdata.GetVerifiedUsers()...)
						}
					}
					aospaProject.LatestTag = latestTag

					if latestTag != mandyStatus.LatestTag {
						utils.LogI(MANDY_TAG, "New tag found "+latestTag)
						mandyStatus.LatestTag = latestTag
						mandyStatus.Notification.Notify(NOTIFICATION_NEW_TAG_FOUND,
							mandyStatus.Strip(),
							mandyStatus.Notification.userdata.GetVerifiedUsers()...)
					}

					saveMandyStatus()
				}

				err = aospaProject.git.Clean()
				if err != nil {
					utils.LogE(MANDY_TAG, err.Error())
				}
			}
		}

		if mandyStatus.Merging {
			startMerging()
		}

		if mandyStatus.Reverting {
			startReverting()
		}

		if mandyStatus.Submitting {
			startSubmitting()
		}

		MandyInit(false, mandyStatus.firebaseApiKey, mandyStatus.Notification.userdata)

		time.Sleep(time.Second * 30)
	}
}

// The higher the score, the better the new tag matches
func compareTag(tag1, tag2 []string) int {

	genericScore := func(score int) (func(tagFrag1, tagFrag2 string) int) {
		return func(tagFrag1, tagFrag2 string) int {
			if score == 0 {
				return 0
			}
			if tagFrag1 == tagFrag2 {
				return score
			}
			return 0
		}
	}

	scoreTable := []func(tagFrag1, tagFrag2 string) int{
		genericScore(100), // LA
		genericScore(100), // UM
		genericScore(50),  // 5
		genericScore(25),  // 6
		genericScore(150), // r1

		// Returning higher number when tag1 is higher
		// That means it's a new tag
		func(tagFrag1, tagFrag2 string) int {
			if tagFrag1 == tagFrag2 {
				return 0
			}
			num1, err := strconv.Atoi(tagFrag1)
			if err != nil {
				return -200
			}
			num2, err := strconv.Atoi(tagFrag2)

			if err != nil {
				return -200
			}
			if num2 > num1 {
				return 100
			}
			return -100
		},

		genericScore(10), // 89xx
		genericScore(0),  // 0
	}

	scoreTablelen := len(scoreTable)
	tag1len := len(tag1)
	tag2len := len(tag2)
	if tag1len != tag2len || tag1len != scoreTablelen || tag2len != scoreTablelen {
		return -1
	}

	var score = 0
	for i := 0; i < scoreTablelen; i++ {
		score += scoreTable[i](tag1[i], tag2[i])
	}

	return score
}

func splitTag(tag string) []string {
	var splitted []string
	tagFragment := strings.Split(tag, "-")
	for _, tagbuf := range tagFragment {
		for _, tagFrag := range strings.Split(tagbuf, ".") {
			splitted = append(splitted, tagFrag)
		}
	}
	return splitted
}

func MandyInit(initialize bool, firebaseApiKey string, userdata *UserData) *MandyStatus {
	if mandyStatus == nil {
		mandyStatus = &MandyStatus{}
		*mandyStatus = getMandyStatus()
		mandyStatus.Notification = NewNotification(firebaseApiKey, userdata)
	}

	if mandyStatus.manifestGit == nil {
		mandyStatus.manifestGit = &git.Git{}
		*mandyStatus.manifestGit = newMandyGit(AOSPA_MANIFEST_NAME, AOSPA_MANIFEST_URL)
	}
	mandyStatus.firebaseApiKey = firebaseApiKey

	if !mandyStatus.manifestGit.Valid() {
		utils.LogI(MANDY_TAG, "Cloning "+mandyStatus.manifestGit.String())
		buf, err := mandyStatus.manifestGit.Clone(AOSPA_BRANCH)
		if err != nil {
			return mandyStatus
		}
		fmt.Println(string(buf))

		utils.LogI(MANDY_TAG, "Successfully cloned "+mandyStatus.manifestGit.String())
	} else {
		err := mandyStatus.manifestGit.Pull("origin", AOSPA_BRANCH)
		if err != nil {
			return mandyStatus
		}
	}
	err := mandyStatus.manifestGit.Clean()
	if err != nil {
		utils.LogE(MANDY_TAG, err.Error())
	}

	manifestDefaultBuf, err := ioutil.ReadFile(mandyStatus.manifestGit.GetPath() + "/" + MANIFEST_DEFAULT)
	if err != nil {
		utils.LogE(MANDY_TAG, mandyStatus.manifestGit.String()+" damaged, cloning again")
		MandyInit(initialize, firebaseApiKey, userdata)
	}

	aospaManifestBuf, err := ioutil.ReadFile(mandyStatus.manifestGit.GetPath() + "/" + MANIFEST_AOSPA)
	if err != nil {
		utils.LogE(MANDY_TAG, mandyStatus.manifestGit.String()+" damaged, cloning again")
		MandyInit(initialize, firebaseApiKey, userdata)
	}

	manifest, err := NewManifest(manifestDefaultBuf)
	utils.Panic(err)

	aospaManifest, err := NewManifest(aospaManifestBuf)
	utils.Panic(err)

	newManifestTag := strings.Replace(
		manifest.Defaults.Revision, "refs/tags/", "", 1)
	if utils.StringEmpty(mandyStatus.LatestTag) {
		mandyStatus.LatestTag = newManifestTag
	}

	// Manifast tag got updated
	// Start tracking again
	if mandyStatus.Submitted && newManifestTag != mandyStatus.ManifestTag {
		mandyStatus.Submitted = false
		mandyStatus.Submitter = User{}
	}

	mandyStatus.ManifestTag = newManifestTag
	mandyStatus.ManifestTagSplitted = splitTag(mandyStatus.ManifestTag)

	// Determinate our forked projects
	// Basically if the path equals it means we forked it
	// Plus check if caf project got removed
	var forkedProjects []*AospaProject
cafLoop:
	for _, cafProject := range manifest.Projects {
		for _, aospaProject := range aospaManifest.Projects {
			if cafProject.Path == aospaProject.Path {
				for _, removedProject := range aospaManifest.RemoveProjects {
					if cafProject.Name == removedProject.Name && repoAccepted(aospaProject.Name) {

						// Track down the remote
						// So we can easily access its name and url
						remote, err := FindRemoteByName(aospaProject.Remote, manifest)
						utils.Panic(err)

						// Get the revision
						// So we get to know what branch the aospa repo uses
						revision, err := GetRevision(aospaProject, manifest)
						utils.Panic(err)

						// Get CAF remote for later merging and fetching tags
						cafRemote, err := FindRemoteByName(cafProject.Remote, manifest)
						utils.Panic(err)

						forkedProjects = append(forkedProjects,
							&AospaProject{
								aospaProject.Path,
								aospaProject,
								revision,
								remote,
								cafProject,
								cafRemote,
								newMandyGit(aospaProject.Path,
									buildGitUrl(aospaProject, remote)),
								aospaProject.Name,
								mandyStatus.ManifestTag,
								false,
							})
						continue cafLoop
					}
				}
			}
		}
	}

	if len(mandyStatus.AospaProjects) != 0 {
		for _, newProject := range forkedProjects {
			for _, aospaProject := range mandyStatus.AospaProjects {
				if newProject.Name == aospaProject.Name {
					// Make sure we don't lose information
					newProject.LatestTag = aospaProject.LatestTag
					newProject.Conflicted = aospaProject.Conflicted

					// Also exit shell
					aospaProject.git.Exit()
				}
			}
		}
	}
	mandyStatus.AospaProjects = forkedProjects

	saveMandyStatus()

	// Cloning forked projects
	for _, aospaProject := range mandyStatus.AospaProjects {
		// Clone the repo now
		// Update it once merging process has started
		if !aospaProject.git.Valid() {
			utils.LogI(MANDY_TAG, "Cloning "+aospaProject.git.String())

			_, err := aospaProject.git.Clone(aospaProject.revision)
			if err != nil {
				continue
			}
			utils.LogI(MANDY_TAG, "Successfully cloned "+aospaProject.git.String())
		}

		// Add caf remote
		aospaProject.git.ReplaceRemote("caf",
			buildGitUrl(aospaProject.cafProject, aospaProject.cafRemote))

		// Add gerrit remote
		aospaProject.git.ReplaceRemote("gerrit", GERRIT_URL+"/"+aospaProject.Name)

		aospaProject.git.SetName(MANDY_NAME)
		aospaProject.git.SetEmail(MANDY_EMAIL)
	}

	if initialize && !mandyStatus.killed {
		// Start goroutine to track new caf tag
		go trackCaf()
	}

	return mandyStatus
}

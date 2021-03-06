package mandy

import (
	"../utils"
	"../utils/git"
	"../utils/shell"
	"io/ioutil"
	"time"
	"strings"
	"strconv"
	"encoding/json"
	"sort"
)

const MANDY_TAG = "Mandy"

const GITHUB_HTTP = "git://github.com"

const AOSPA_MANIFEST_NAME = "manifest"
const AOSPA_MANIFEST_URL = GITHUB_HTTP + "/AOSPA/manifest.git"
const MANIFEST_DEFAULT = "default.xml"
const MANIFEST_AOSPA = "manifests/aospa.xml"

const CAF_BRANCH = "caf"

const AOSPA_BRANCH = "quartz"
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
	Path       string   `json:"-"`
	project    Project  `json:"-"`
	revision   string   `json:"-"`
	remote     Remote   `json:"-"`
	cafProject Project  `json:"-"`
	cafRemote  Remote   `json:"-"`
	git        *git.Git `json:"-"`

	Name       string `json:"name"`
	LatestTag  string `json:"latesttag"`
	Conflicted bool   `json:"conflicted"`
}

type MandyStatus struct {
	ManifestTag string `json:"manifesttag"`

	LatestTag     string          `json:"latesttag"`
	AospaProjects []*AospaProject `json:"projects"`

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
	firebaseApiKey string        `json:"-"`
	killed         bool          `json:"-"`
	manifestGit    *git.Git      `json:"-"`

	shell *shell.Shell `json:"-"`
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

func (mandyStatus *MandyStatus) newMandyGit(path, url string) *git.Git {
	return git.NewGit("mandy", path, url, mandyStatus.shell)
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

func getMandyStatus() *MandyStatus {
	buf, err := ioutil.ReadFile(utils.MANDY + "/status.json")
	if err != nil {
		return &MandyStatus{}
	}

	var mandyStatus *MandyStatus
	err = json.Unmarshal(buf, &mandyStatus)
	if err != nil {
		return &MandyStatus{}
	}

	return mandyStatus
}

func (mandyStatus *MandyStatus) Kill() {
	if mandyStatus != nil {
		mandyStatus.killed = true
		mandyStatus.shell.Exit()
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
			aospaProject.git.RemoveAll()
			continue
		}

		status, err := aospaProject.git.MergeBranch("origin", aospaProject.revision)
		if status != 0 || err != nil {
			utils.LogE(MANDY_TAG, "Couldn't merge with origin "+aospaProject.Name+", skipping submission")
			aospaProject.Conflicted = true
			successful = false
			aospaProject.git.RemoveAll()
			continue
		}
	}

	if successful {
		for _, aospaProject := range mandyStatus.AospaProjects {
			aospaProject.git.Push("gerrit", "HEAD:refs/heads/"+aospaProject.revision, false)
			aospaProject.git.RemoveAll()
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
		aospaProject.git.RemoveAll()
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
			aospaProject.git.Fetch(CAF_BRANCH)
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

		if err != nil {
			utils.LogE(MANDY_TAG, err.Error())
		}
		aospaProject.git.RemoveAll()
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

	type tagScore struct {
		name           string
		score, numDiff int
	}

	for {

		if !mandyStatus.Merged && !mandyStatus.Submitted && !mandyStatus.Reverting {
			for _, aospaProject := range mandyStatus.AospaProjects {
				if !aospaProject.git.Valid() {
					continue
				}

				// Fetch caf
				utils.LogI(MANDY_TAG, "Fetching "+aospaProject.Name)
				err := aospaProject.git.Fetch(CAF_BRANCH)
				if err != nil {
					utils.LogE(MANDY_TAG, "Failed to fetch "+aospaProject.Name)
					continue
				}
				utils.LogI(MANDY_TAG, "Fetched "+aospaProject.Name)

				latestTag := mandyStatus.ManifestTag
				tags, err := aospaProject.git.GetTags()
				utils.Panic(err)

				allTags := make([]string, len(tags))
				copy(allTags, tags)
				for _, tag := range tags {

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
						if !utils.StringEmpty(aliasTag) {
							allTags = append(allTags, aliasTag)
						}
					}
				}

				var topTags []tagScore
				for _, tag := range allTags {
					score, numDiff := compareTag(splitTag(mandyStatus.ManifestTag), splitTag(tag))
					if score >= 515 {
						topTags = append(topTags, tagScore{tag, score, numDiff})
					}
				}

				sort.Slice(topTags, func(i, j int) bool {
					return topTags[i].score > topTags[j].score
				})

				if len(topTags) > 0 {
					topScore := topTags[0].score
					maxScore := 0
					for _, tag := range topTags {
						if tag.score != topScore {
							break
						}
						if tag.score+tag.numDiff > maxScore {
							maxScore = tag.score + tag.numDiff
							latestTag = tag.name
						}
					}
				}

				if latestTag != mandyStatus.ManifestTag {
					if !mandyStatus.Mergeable && !mandyStatus.Merged && !mandyStatus.Submitting {
						mergable := true
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

					_, latestProjectTagNumDiff := compareTag(splitTag(mandyStatus.ManifestTag), splitTag(latestTag))
					_, latestTagNumDiff := compareTag(splitTag(mandyStatus.ManifestTag), splitTag(mandyStatus.LatestTag))
					if latestTag != mandyStatus.LatestTag && latestProjectTagNumDiff > latestTagNumDiff {
						utils.LogI(MANDY_TAG, "New tag found "+latestTag)
						mandyStatus.LatestTag = latestTag
						mandyStatus.Notification.Notify(NOTIFICATION_NEW_TAG_FOUND,
							mandyStatus.Strip(),
							mandyStatus.Notification.userdata.GetVerifiedUsers()...)
					}

					saveMandyStatus()
				}

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
func compareTag(tag1, tag2 []string) (int, int) {

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

	var numDiff int
	scoreTable := []func(tagFrag1, tagFrag2 string) int{
		genericScore(100), // LA
		genericScore(100), // UM
		genericScore(50),  // 5
		genericScore(25),  // 6
		genericScore(150), // r1

		// Returning higher number when tag1 is higher
		// That means it's a new tag
		func(tagFrag1, tagFrag2 string) int {
			num1, err := strconv.Atoi(tagFrag1)
			if err != nil {
				return -200
			}
			num2, err := strconv.Atoi(tagFrag2)
			if err != nil {
				return -200
			}
			numDiff = num2 - num1
			return numDiff
		},

		genericScore(10), // 89xx
		genericScore(0),  // 0
	}

	scoreTablelen := len(scoreTable)
	tag1len := len(tag1)
	tag2len := len(tag2)
	if tag1len != tag2len || tag1len != scoreTablelen ||
		tag2len != scoreTablelen {
		return -1, 0
	}

	score := 0
	for i := 0; i < scoreTablelen; i++ {
		score += scoreTable[i](tag1[i], tag2[i])
	}

	if numDiff > 0 {
		score += 100
	}
	return score - numDiff, numDiff
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
		mandyStatus = getMandyStatus()

		mandyStatus.Notification = NewNotification(firebaseApiKey, userdata)
		mandyStatus.firebaseApiKey = firebaseApiKey

		mandyStatus.shell = shell.NewShell()
		mandyStatus.manifestGit = mandyStatus.newMandyGit(AOSPA_MANIFEST_NAME, AOSPA_MANIFEST_URL)
	}

	if !mandyStatus.manifestGit.Valid() {
		utils.LogI(MANDY_TAG, "Cloning "+mandyStatus.manifestGit.String())
		_, err := mandyStatus.manifestGit.Clone(AOSPA_BRANCH)
		if err != nil {
			return mandyStatus
		}

		utils.LogI(MANDY_TAG, "Successfully cloned "+mandyStatus.manifestGit.String())
	} else {
		err := mandyStatus.manifestGit.Pull("origin", AOSPA_BRANCH)
		if err != nil {
			return mandyStatus
		}
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

	// Determinate our forked projects
	// Basically if the path equals it means we forked it
	// Plus check if caf project got removed
	var forkedProjects []*AospaProject
	for _, cafProject := range manifest.Projects {
		for _, aospaProject := range aospaManifest.Projects {
			if cafProject.Path == aospaProject.Path && aospaProject.Remote == "aospa" {
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
								mandyStatus.newMandyGit(aospaProject.Path,
									buildGitUrl(aospaProject, remote)),
								aospaProject.Name,
								mandyStatus.ManifestTag,
								false,
							})
						break
					}
				}
				break
			}
		}
	}

	for _, aospaProject := range mandyStatus.AospaProjects {
		for _, forkedProject := range forkedProjects {
			if forkedProject.Name == aospaProject.Name {
				// Make sure we don't lose information
				forkedProject.LatestTag = aospaProject.LatestTag
				forkedProject.Conflicted = aospaProject.Conflicted
				break
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
		aospaProject.git.ReplaceRemote(CAF_BRANCH,
			buildGitUrl(aospaProject.cafProject, aospaProject.cafRemote))

		// Add gerrit remote
		aospaProject.git.ReplaceRemote("gerrit", GERRIT_URL+"/"+aospaProject.Name)

		aospaProject.git.SetName(MANDY_NAME)
		aospaProject.git.SetEmail(MANDY_EMAIL)

		aospaProject.git.RemoveAll()
	}

	if initialize && !mandyStatus.killed {
		// Start goroutine to track new caf tag
		go trackCaf()
	}

	return mandyStatus
}

package git

import (
	"../shell"
	"../../utils"
	"fmt"
	"os"
	"strings"
	"io/ioutil"
)

type Git struct {
	parentPath, path, url string
	shell                 shell.Shell
}

type GitError string

func (e GitError) Error() string {
	return string(e)
}

func NewGit(parentPath, path, url string) Git {
	return Git{
		parentPath,
		path,
		url,
		shell.NewShell(),
	}
}

func (git Git) Clone(branch string) ([]byte, error) {
	git.Delete()
	buf, status, err := git.shell.Run([]byte(fmt.Sprintf(
		"git clone %s -b %s %s", git.url, branch, git.GetPath())))
	if status != 0 {
		return buf, GitError("Cloning " + git.String() + " failed")
	}
	return buf, err
}

func (git Git) GetRemotes() ([]string, error) {
	buf, status, err := git.run("remote")
	if err != nil || status != 0 {
		return nil, err
	}

	return strings.Split(string(buf), "\n"), nil
}

func (git Git) AddRemote(name, url string) error {
	remotes, err := git.GetRemotes()
	if err != nil {
		return err
	}
	if utils.SliceContains(name, remotes) {
		return GitError(name + " already exists")
	}

	_, _, err = git.run("remote add " + name + " " + url)
	return err
}

func (git Git) ReplaceRemote(name, url string) error {
	git.RemoveRemote(name)
	return git.AddRemote(name, url)
}

func (git Git) RemoveRemote(name string) error {
	_, _, err := git.run("remote rm " + name)
	return err
}

func (git Git) MergeBranch(remote, branch string) (int, error) {
	_, status, err := git.run("merge " + remote + "/" + branch + " --no-edit")
	return status, err
}

func (git Git) MergeTag(tag string) (int, error) {
	_, status, err := git.run("merge " + tag + " --no-edit")
	return status, err
}

func (git Git) Pull(remote, branch string) error {
	err := git.Fetch(remote)
	if err != nil {
		return err
	}

	_, status, err := git.run("reset --hard " + remote + "/" + branch)
	if status != 0 || err != nil {
		return GitError("Remote " + remote + " doesn't have the branch " + branch)
	}

	return nil
}

func (git Git) Fetch(remote string) error {
	_, status, err := git.run("fetch " + remote)
	if err != nil {
		return err
	}
	if status != 0 {
		return GitError("Can't fetch remote " + remote)
	}

	return nil
}

func (git Git) GetTags() ([]string, error) {
	buf, status, err := git.run("tag")
	if err != nil || status != 0 {
		return nil, err
	}

	return strings.Split(string(buf), "\n"), nil
}

func (git Git) SetName(name string) (int, error) {
	_, status, err := git.run("config user.name \"" + name + "\"")
	return status, err
}

func (git Git) SetEmail(email string) (int, error) {
	_, status, err := git.run("config user.email \"" + email + "\"")
	return status, err
}

func (git Git) Push(remote, branch string, force bool) (int, error) {
	cmd := "push " + remote + " " + branch
	if force {
		cmd += " -f"
	}
	_, status, err := git.run(cmd)
	return status, err
}

func (git Git) Clean() error {
	_, status, err := git.run("clean -d -x -f")
	if status != 0 || err != nil {
		return GitError("Couldn't clean " + git.path)
	}
	return nil
}

func (git Git) RemoveAll() error {
	path := git.GetPath()
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return GitError("Couldn't remove all " + git.path)
	}
	for _, file := range files {
		if hasGitSubdir(path + "/" + file.Name()) {
			continue
		}
		os.RemoveAll(path + "/" + file.Name())
	}
	return nil
}

func hasGitSubdir(filePath string) bool {
	if strings.HasSuffix(filePath, ".git") {
		return true
	}

	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return false
	}
	for _, file := range files {
		if hasGitSubdir(filePath + "/" + file.Name()) {
			return true
		}
	}
	return false
}

func (git Git) ResetHard() error {
	_, status, err := git.run("reset --hard")
	if status != 0 || err != nil {
		return GitError("Couldn't reset hard " + git.path)
	}
	return nil
}

func (git Git) run(cmd string) ([]byte, int, error) {
	return git.shell.Run([]byte("git -C " + git.GetPath() + " " + cmd))
}

func (git Git) Valid() bool {
	if git.Exists() {
		buf, status, err := git.run("rev-parse --is-inside-work-tree")
		return err == nil && status == 0 && string(buf) == "true"
	}
	return false
}

func (git Git) Exists() bool {
	return utils.DirExists(git.GetPath() + "/" + ".git")
}

func (git Git) GetPath() string {
	return fmt.Sprintf("serverdata/%s/%s", git.parentPath, git.path)
}

func (git Git) Delete() {
	path := git.GetPath()
	if utils.DirExists(path) {
		os.Remove(path)
	}
}

func (git Git) String() string {
	return git.url
}

func (git Git) Exit() {
	git.shell.Exit()
}

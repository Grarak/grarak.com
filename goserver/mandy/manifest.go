package mandy

import (
	"../utils"
	"encoding/xml"
)

type Remote struct {
	Name     string `xml:"name,attr"`
	Fetch    string `xml:"fetch,attr"`
	Revision string `xml:"revision,attr"`
}

type Defaults struct {
	Revision string `xml:"revision,attr"`
	Remote   string `xml:"remote,attr"`
}

type Project struct {
	Name     string `xml:"name,attr"`
	Path     string `xml:"path,attr"`
	Remote   string `xml:"remote,attr"`
	Revision string `xml:"revision,attr"`
}

type RemoveProject struct {
	Name string `xml:"name,attr"`
}

type Manifest struct {
	XMLName        xml.Name `xml:"manifest"`
	Remotes        []Remote `xml:"remote"`
	Defaults       Defaults `xml:"default"`
	Projects       []Project `xml:"project"`
	RemoveProjects []RemoveProject `xml:"remove-project"`
}

type ManifestErr string

func (e ManifestErr) Error() string {
	return string(e)
}

func NewManifest(data []byte) (Manifest, error) {
	var manifest Manifest
	err := xml.Unmarshal(data, &manifest)
	if err == nil {
		// In some projects path is empty
		// That means it's the same as name
		for index, project := range manifest.Projects {
			if utils.StringEmpty(project.Path) {
				// Don't use the project variable
				// It won't overwrite the memory location of path
				// If you want to workaround this just use a pointer
				manifest.Projects[index].Path = project.Name
			}
		}
	}
	return manifest, err
}

func GetRevision(project Project, manifest Manifest) (string, error) {
	if !utils.StringEmpty(project.Revision) {
		return project.Revision, nil
	}

	remote, err := FindRemoteByName(project.Remote, manifest)
	if err == nil {
		return remote.Revision, nil
	}

	return "", ManifestErr("Couldn't get revision of " + project.Name)
}

func FindRemoteByName(name string, manifest Manifest) (Remote, error) {
	if utils.StringEmpty(name) {
		return FindRemoteByName(manifest.Defaults.Remote, manifest)
	}
	for _, remote := range manifest.Remotes {
		if name == remote.Name {
			return remote, nil
		}
	}
	return Remote{"", "", ""},
		ManifestErr("Couldn't find remote " + name)
}

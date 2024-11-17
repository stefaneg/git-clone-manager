package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"tools/internal/color"
	"tools/internal/gitrepo"
	l "tools/internal/log"
)

type Config struct {
	GitLab []GitLabConfig `yaml:"gitlab"`
}

type GitLabConfig struct {
	EnvTokenVariableName string    `yaml:"tokenEnvVar"`    // The environment variable name for the GitLab token
	HostName             string    `yaml:"hostName"`       // Gitlab host name
	CloneDirectory       string    `yaml:"cloneDirectory"` // Where to clone projects in local directory structure
	Groups               []Group   `yaml:"groups"`
	Projects             []Project `yaml:"projects"`
}

type Group struct {
	ID            string `yaml:"id"`
	CloneArchived bool   `yaml:"cloneArchived"`
}

type Project struct {
	Name     string `yaml:"name"`
	FullPath string `yaml:"fullPath"`
}

type Subgroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CustomFormatter struct {
	logrus.TextFormatter
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if entry.Level == logrus.InfoLevel {
		entry.Message = fmt.Sprintf("%s\n", entry.Message)
		return []byte(entry.Message), nil
	}
	return f.TextFormatter.Format(entry)
}

func (gitlab GitLabConfig) GetCloneDirectory() string {
	return gitlab.CloneDirectory
}

func (gitlab GitLabConfig) getBaseUrl() string {
	return fmt.Sprintf("https://%s/api/v4", gitlab.HostName)
}

type NullableBool struct {
	Value *bool
}

func (nb *NullableBool) Set(s string) error {
	v := (s == "true")
	nb.Value = &v
	return nil
}

func (nb *NullableBool) String() string {
	if nb.Value == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", *nb.Value)
}

func (nb *NullableBool) Val(defaultValue bool) bool {
	if nb.Value == nil {
		return defaultValue
	}
	return *nb.Value
}

func (nb *NullableBool) IsBoolFlag() bool {
	return true
}

func main() {
	l.Log.SetFormatter(&CustomFormatter{logrus.TextFormatter{}})

	var verbose = NullableBool{}
	flag.Var(&verbose, "verbose", "Print verbose output")
	flag.Parse()
	if verbose.Val(false) {
		l.Log.SetLevel(logrus.DebugLevel)
		l.Log.Debugln("Verbose (debug) logging enabled")
	}

	config, err := loadConfig("workingCopies.yaml") // Replace with your config file name
	if err != nil {
		l.Log.Fatalf("Failed to load configuration: %v", err)
	}

	for _, gitlab := range config.GitLab {
		l.Log.Infof("Cloning groups/projects in %s into %s", color.FgCyan(gitlab.HostName), color.FgCyan(gitlab.CloneDirectory))
		err = os.MkdirAll(gitlab.GetCloneDirectory(), os.ModePerm)
		if err != nil {
			l.Log.Fatalf("Failed to create clone directory: %v", err)
		}

		token := os.Getenv(gitlab.EnvTokenVariableName)
		if token == "" {
			l.Log.Printf("Environment variable %s not set; skipping", gitlab.EnvTokenVariableName)
			continue
		}

		for _, group := range gitlab.Groups {
			l.Log.Infof("Cloning projects in group  %s\n", color.FgCyan(group.ID))
			err := gitlab.cloneGroupProjects(token, group)
			if err != nil {
				l.Log.Printf("Failed to clone projects for group %s: %v", group.ID, err)
			}
		}

		for _, prj := range gitlab.Projects {
			project := convertToGitlabProjectSpec(prj, gitlab)
			if err != nil {
				l.Log.Printf("Failed to fetch project %s: %v", prj.FullPath, err)
				continue
			}
			err = project.CloneProject(gitlab.CloneDirectory, true)
			if err != nil {
				l.Log.Printf("Failed to clone project %s: %v", project.Name, err)
			}
		}
	}
	l.Log.Infoln(color.FgGreen("Done!"))
}

func loadConfig(configFileName string) (*Config, error) {
	configFilePath := filepath.Join("./", configFileName)

	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not determine home directory: %v", err)
		}
		configFilePath = filepath.Join(homeDir, configFileName)
		if _, err := os.Stat(configFileName); os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found in current directory or home directory")
		}
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config file: %v", err)
	}

	return &config, nil
}

func (gitlab GitLabConfig) cloneGroupProjects(token string, group Group) error {
	// Fetch and clone all projects in the group
	projects, err := gitlab.fetchProjects(token, group.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch projects for group %s: %w", group.ID, err)
	}
	for _, project := range projects {
		err := project.CloneProject(gitlab.CloneDirectory, group.CloneArchived)
		if err != nil {
			l.Log.Printf("Failed to clone project %s: %v", project.Name, err)
		}
	}

	// Fetch subgroups and recursively clone their projects
	subgroups, err := gitlab.fetchSubgroups(token, group.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch subgroups for group %s: %w", group.ID, err)
	}
	for _, subgroup := range subgroups {
		err := gitlab.cloneGroupProjects(token, Group{ID: fmt.Sprintf("%d", subgroup.ID), CloneArchived: group.CloneArchived})
		if err != nil {
			l.Log.Printf("Failed to clone projects for subgroup %s: %v", subgroup.Name, err)
		}
	}

	return nil
}

func convertToGitlabProjectSpec(project Project, gitlab GitLabConfig) *gitrepo.GitRepoSpec {
	var gitlabProjectSpec = gitrepo.GitRepoSpec{
		Name:              project.Name,
		PathWithNamespace: project.FullPath,
		SSHURLToRepo:      fmt.Sprintf("git@%s:%s", gitlab.HostName, project.FullPath),
	}

	return &gitlabProjectSpec
}

func (gitlab GitLabConfig) fetchProjects(token, groupID string) ([]gitrepo.GitRepoSpec, error) {
	l.Log.Debugf("Fetching projects in group %s\n", color.FgCyan(groupID))
	url := fmt.Sprintf("%s/groups/%s/projects", gitlab.getBaseUrl(), groupID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			l.Log.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API request failed with status: %s", resp.Status)
	}

	var projects []gitrepo.GitRepoSpec
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func (gitlab GitLabConfig) fetchSubgroups(token, groupID string) ([]Subgroup, error) {
	l.Log.Debugf("Fetching subgroups in group %s", color.FgCyan(groupID))
	url := fmt.Sprintf("%s/groups/%s/subgroups", gitlab.getBaseUrl(), groupID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			l.Log.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API request failed with status: %s", resp.Status)
	}

	var subgroups []Subgroup
	if err := json.NewDecoder(resp.Body).Decode(&subgroups); err != nil {
		return nil, err
	}

	return subgroups, nil
}

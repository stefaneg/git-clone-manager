package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
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
	ID string `yaml:"id"`
}

type Project struct {
	Name     string `yaml:"name"`
	FullPath string `yaml:"fullPath"`
}

type ProjectGitlabSpec struct {
	Name              string `json:"name"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	PathWithNamespace string `json:"path_with_namespace"`
}

type Subgroup struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (gitlab GitLabConfig) getCloneDirectory() string {
	return gitlab.CloneDirectory
}

func (gitlab GitLabConfig) getBaseUrl() string {
	return fmt.Sprintf("https://%s/api/v4", gitlab.HostName)
}

func main() {

	config, err := loadConfig("config.yaml") // Replace with your config file name
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	for _, gitlab := range config.GitLab {
		err = os.MkdirAll(gitlab.getCloneDirectory(), os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create clone directory: %v", err)
		}

		token := os.Getenv(gitlab.EnvTokenVariableName)
		if token == "" {
			log.Printf("Environment variable %s not set; skipping", gitlab.EnvTokenVariableName)
			continue
		}

		for _, group := range gitlab.Groups {
			err := cloneGroupProjects(token, group.ID, gitlab)
			if err != nil {
				log.Printf("Failed to clone projects for group %s: %v", group.ID, err)
			}
		}

		// Process individual projects
		for _, prj := range gitlab.Projects {
			project := convertToGitlabProjectSpec(prj, gitlab)
			if err != nil {
				log.Printf("Failed to fetch project %s: %v", prj.FullPath, err)
				continue
			}
			err = cloneProject(project, gitlab)
			if err != nil {
				log.Printf("Failed to clone project %s: %v", project.Name, err)
			} else {
				fmt.Printf("Successfully cloned project %s\n", project.Name)
			}
		}

	}
}

func loadConfig(filePath string) (*Config, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func cloneGroupProjects(token, groupID string, gitlab GitLabConfig) error {
	// Fetch and clone all projects in the group
	projects, err := fetchProjects(token, groupID, gitlab)
	if err != nil {
		return fmt.Errorf("failed to fetch projects for group %s: %w", groupID, err)
	}
	for _, project := range projects {
		err := cloneProject(&project, gitlab)
		if err != nil {
			log.Printf("Failed to clone project %s: %v", project.Name, err)
		} else {
			fmt.Printf("Successfully cloned project %s\n", project.Name)
		}
	}

	// Fetch subgroups and recursively clone their projects
	subgroups, err := fetchSubgroups(token, groupID, gitlab)
	if err != nil {
		return fmt.Errorf("failed to fetch subgroups for group %s: %w", groupID, err)
	}
	for _, subgroup := range subgroups {
		err := cloneGroupProjects(token, fmt.Sprintf("%d", subgroup.ID), gitlab)
		if err != nil {
			log.Printf("Failed to clone projects for subgroup %s: %v", subgroup.Name, err)
		}
	}

	return nil
}

func convertToGitlabProjectSpec(project Project, gitlab GitLabConfig) *ProjectGitlabSpec {
	var gitlabProjectSpec = ProjectGitlabSpec{
		Name:              project.Name,
		PathWithNamespace: project.FullPath,
		SSHURLToRepo:      fmt.Sprintf("git@%s:%s", gitlab.HostName, project.FullPath),
	}

	return &gitlabProjectSpec
}

func fetchProjects(token, groupID string, gitlab GitLabConfig) ([]ProjectGitlabSpec, error) {
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API request failed with status: %s", resp.Status)
	}

	var projects []ProjectGitlabSpec
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func fetchSubgroups(token, groupID string, gitlab GitLabConfig) ([]Subgroup, error) {
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API request failed with status: %s", resp.Status)
	}

	var subgroups []Subgroup
	if err := json.NewDecoder(resp.Body).Decode(&subgroups); err != nil {
		return nil, err
	}

	return subgroups, nil
}

func cloneProject(project *ProjectGitlabSpec, gitlab GitLabConfig) error {
	projectPath := path.Join(gitlab.getCloneDirectory(), project.PathWithNamespace)

	// Check if the project directory already exists
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		fmt.Printf("ProjectGitlabSpec %s already exists at %s, skipping clone\n", project.Name, projectPath)
		return nil // Skip cloning if directory already exists
	}

	fmt.Printf("Cloning project to %s\n", projectPath)

	cmd := exec.Command("git", "clone", project.SSHURLToRepo, projectPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s", string(output))
	}

	return nil
}

//go:build integration

package integration_test

import (
	"fmt"
	"gcm/internal/gitlab"
	"log"
	"os"
	"testing"
)

func GetApiTokenFromEnv() string {
	token := os.Getenv("GITLAB_API_TOKEN")
	if token == "" {
		log.Fatal("required environment variable GITLAB_API_TOKEN is not set")
	}
	return token
}

func TestAPIClient_FetchProjects(t *testing.T) {
	// Replace with a valid token and hostname for your GitLab instance
	token := GetApiTokenFromEnv()
	hostName := "gitlab.example.com"

	apiClient := gitlab.NewAPIClient(token, hostName)

	// Replace with a valid group ID in your GitLab instance
	group := &gitlab.Group{ID: 1}

	projects, err := apiClient.FetchProjects(group)
	if err != nil {
		t.Fatalf("Failed to fetch projects: %v", err)
	}

	if len(projects) == 0 {
		t.Fatalf("Expected to fetch at least one project, got none")
	}

	t.Logf("Fetched %d projects successfully", len(projects))
}

func TestAPIClient_FetchSubgroups(t *testing.T) {
	// Replace with a valid token and hostname for your GitLab instance
	token := GetApiTokenFromEnv()
	hostName := "gitlab.example.com"

	apiClient := gitlab.NewAPIClient(token, hostName)

	// Replace with a valid group ID in your GitLab instance
	groupID := "1"

	subgroups, err := apiClient.FetchSubgroups(groupID)
	if err != nil {
		t.Fatalf("Failed to fetch subgroups: %v", err)
	}

	t.Logf("Fetched %d subgroups successfully", len(subgroups))
}

func TestAPIClient_FetchValidGroupInfo(t *testing.T) {
	// Replace with a valid token and hostname for your GitLab instance
	token := GetApiTokenFromEnv()
	hostName := "gitlab.controlant.com"

	apiClient := gitlab.NewAPIClient(token, hostName)

	// Replace with a valid group ID in your GitLab instance
	groupID := "catapult"

	groupInfo, err := apiClient.FetchGroupInfo(groupID)
	if err != nil {
		t.Fatalf("Failed to fetch group info: %v", err)
	}

	if groupInfo == nil {
		t.Fatalf("Expected to fetch group info, got nil")
	}

	t.Logf("Fetched group info successfully: %+v", groupInfo)
}

func TestAPIClient_FetchInvalidGroupInfo(t *testing.T) {
	// Replace with a valid token and hostname for your GitLab instance
	token := GetApiTokenFromEnv()
	hostName := "gitlab.controlant.com"

	apiClient := gitlab.NewAPIClient(token, hostName)

	// Replace with a valid group ID in your GitLab instance
	groupID := "random_non_existing_group"

	_, err := apiClient.FetchGroupInfo(groupID)

	if err == nil {
		t.Fatalf("Expected to get an error")
	}
	expectedError := "GitLab API request on https://gitlab.controlant.com/api/v4/groups/random_non_existing_group failed with status: 404 Not Found"
	if fmt.Sprintf(
		"%v",
		err,
	) != expectedError {
		t.Fatalf("Unexpected error, expected %v, got %v", expectedError, err)
	}
}

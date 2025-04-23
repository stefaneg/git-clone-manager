//go:build integration

package integration_test

import (
	"fmt"
	"gcm/internal/gitlab"
	"log"
	"os"
	"testing"
)

const apiTokenVariableName = "GITLAB_API_TOKEN"
const hostNameVariableName = "GITLAB_TEST_HOST_NAME"

func GetRequiredEnvVariable(variableName string) string {
	token := os.Getenv(variableName)
	if token == "" {
		log.Fatalf("required environment variable %v is not set", variableName)
	}
	return token
}

func newTestApiClient() gitlab.API {
	token := GetRequiredEnvVariable(apiTokenVariableName)
	hostName := GetRequiredEnvVariable(hostNameVariableName)
	apiClient := gitlab.NewAPIClient(token, hostName)
	return apiClient
}

func TestAPIClient_FetchAccessibleProjects(t *testing.T) {
	// Retrieve the API token from the environment
	apiClient := newTestApiClient()

	// Fetch accessible projects
	projects, err := apiClient.FetchAccessibleProjects()
	if err != nil {
		t.Fatalf("Failed to fetch accessible projects: %v", err)
	}

	// Ensure at least one project is returned
	if len(projects) == 0 {
		t.Fatalf("Expected to fetch at least one accessible project, got none")
	}

	t.Logf("Fetched %d accessible projects successfully", len(projects))
	for _, project := range projects {
		t.Logf("Fetched project: %v", project)
	}
}

func TestAPIClient_FetchProjects(t *testing.T) {
	apiClient := newTestApiClient()

	// Replace with a valid group ID in your GitLab instance
	group := &gitlab.Group{ID: 1}

	projects, err := apiClient.FetchGroupProjects(group)
	if err != nil {
		t.Fatalf("Failed to fetch projects: %v", err)
	}

	if len(projects) == 0 {
		t.Fatalf("Expected to fetch at least one project, got none")
	}

	t.Logf("Fetched %d projects successfully", len(projects))
}

func TestAPIClient_FetchSubgroups(t *testing.T) {
	apiClient := newTestApiClient()

	// Replace with a valid group ID in your GitLab instance
	groupID := "1"

	subgroups, err := apiClient.FetchSubgroups(groupID)
	if err != nil {
		t.Fatalf("Failed to fetch subgroups: %v", err)
	}

	t.Logf("Fetched %d subgroups successfully", len(subgroups))
}

func TestAPIClient_FetchValidGroupInfo(t *testing.T) {
	apiClient := newTestApiClient()

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
	apiClient := newTestApiClient()

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

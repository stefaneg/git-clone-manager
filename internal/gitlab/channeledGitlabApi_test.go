package gitlab

import (
	"errors"
	"fmt"
	"gcm/internal/counter"
	"gcm/internal/gitremote"
	"testing"
)

func expectGroupCount(t *testing.T, groupCounter *counter.Counter, expectedGroupCount int) {
	if groupCounter.Count() != expectedGroupCount {
		t.Fatalf("expected group counter to be %d, got %d", expectedGroupCount, groupCounter.Count())
	}
}

func expectProjectCount(t *testing.T, projects []Project, projectCounter *counter.Counter, exepectedCount int) {
	if len(projects) != exepectedCount {
		t.Fatalf("expected %d projects, got %d", exepectedCount, len(projects))
	}

	if projectCounter.Count() != exepectedCount {
		t.Fatalf("expected project counter to be %d, got %d", exepectedCount, projectCounter.Count())
	}
}

func panicOnError(errorChannel chan error) {
	go func() {
		for err := range errorChannel {
			// If using regular test failure, we get hard to read deadlock failure.
			panic(fmt.Sprintf("Test failed with error %v", err))
		}
	}()
}

func expectErrors(t *testing.T, errorChannel chan error, expectedErrors []string) {
	expectedSet := make(map[string]struct{})
	for _, err := range expectedErrors {
		expectedSet[err] = struct{}{}
	}

	for err := range errorChannel {
		if _, exists := expectedSet[err.Error()]; !exists {
			t.Fatalf("unexpected error: %v", err)
		}

		// Remove the matched error from the set
		delete(expectedSet, err.Error())

		// Stop reading if all expected errors are processed
		if len(expectedSet) == 0 {
			return
		}
	}

	if len(expectedSet) > 0 {
		t.Fatalf("expected errors not found: %v", expectedErrors)
	}
}

func TestChanneledApi_ScheduleFetchMultiRootGitlabGroupProjects(t *testing.T) {
	fakeAPI := &FakeAPI{
		GroupInfo: map[string]*Group{
			"root1": {ID: 1, Name: "root1"},
			"root2": {ID: 4, Name: "root2"},
		},
		Subgroups: map[string][]Group{
			"1": {
				{ID: 2, Name: "subgroup1"},
				{ID: 3, Name: "subgroup2"},
			},
			"2": {},
			"3": {},
			"4": {
				{ID: 5, Name: "subgroup3"},
				{ID: 6, Name: "subgroup4"},
			},
			"5": {},
			"6": {},
		},
		Projects: map[string][]Project{
			"1": {},
			"2": {
				{Name: "subgroup1, project1", SSHURLToRepo: "git@somewhere:project1.git"},
				{Name: "subgroup1, project2", SSHURLToRepo: "git@somewhere:project2.git"},
			},
			"3": {
				{Name: "subgroup2, project3", SSHURLToRepo: "git@somewhere:project3.git"},
				{Name: "subgroup2, project4", SSHURLToRepo: "git@somewhere:project4.git"},
			},
			"4": {},
			"5": {
				{Name: "subgroup3, project5", SSHURLToRepo: "git@somewhere:project5.git"},
				{Name: "subgroup3, project6", SSHURLToRepo: "git@somewhere:project6.git"},
			},
			"6": {
				{Name: "subgroup4, project7", SSHURLToRepo: "git@somewhere:project7.git"},
				{Name: "subgroup4, project8", SSHURLToRepo: "git@somewhere:project8.git"},
			},
		},
	}
	projectCounter := counter.NewCounter()
	groupCounter := counter.NewCounter()

	errorChannel := make(chan error, 10)
	panicOnError(errorChannel)

	config := &GitLabConfig{
		Projects: []gitremote.ProjectConfig{
			{Name: "project1"},
			{Name: "project2"},
		},
	}

	channeledApi := NewChanneledApi(fakeAPI, config, projectCounter, groupCounter, errorChannel)

	groupConfigs := []GroupConfig{
		{Name: "root1"},
		{Name: "root2"},
	}

	projectChannel := channeledApi.ScheduleFetchGitlabGroupProjects(groupConfigs)
	var projects []Project
	for project := range projectChannel {
		projects = append(projects, project)
	}

	expectProjectCount(t, projects, projectCounter, 8)
	expectGroupCount(t, groupCounter, 4)
}

func TestChanneledApi_ErrorHandling_FetchGroupInfo(t *testing.T) {
	fakeAPI := &FakeAPI{
		FetchError: errors.New("failed to fetch group info"),
	}
	errorChannel := make(chan error, 10)
	projectCounter := counter.NewCounter()
	groupCounter := counter.NewCounter()

	channeledApi := NewChanneledApi(fakeAPI, &GitLabConfig{}, projectCounter, groupCounter, errorChannel)

	groupConfigs := []GroupConfig{
		{Name: "nonexistent-group"},
	}
	channeledApi.ScheduleFetchGitlabGroupProjects(groupConfigs)

	expectErrors(
		t,
		errorChannel,
		[]string{"failed to fetch rootGroupConfig info for rootGroupConfig nonexistent-group: failed to fetch group info"},
	)
}

func TestChanneledApi_ErrorHandling_FetchSubgroups(t *testing.T) {
	// Simulate an error when fetching subgroups
	fakeAPI := &FakeAPI{
		GroupInfo: map[string]*Group{
			"root1": {ID: 1, Name: "root1"},
		},
		FetchError: errors.New("intentional fetch error"),
	}
	errorChannel := make(chan error, 10)
	projectCounter := counter.NewCounter()
	groupCounter := counter.NewCounter()

	channeledApi := NewChanneledApi(fakeAPI, &GitLabConfig{}, projectCounter, groupCounter, errorChannel)

	groupConfigs := []GroupConfig{
		{Name: "root1"},
	}

	channeledApi.ScheduleFetchGitlabGroupProjects(groupConfigs)

	expectErrors(
		t,
		errorChannel,
		[]string{"failed to fetch rootGroupConfig info for rootGroupConfig root1: intentional fetch error"},
	)
}

func TestChanneledApi_ErrorHandling_FetchProjects(t *testing.T) {
	// Simulate an error when fetching projects
	fakeAPI := &FakeAPI{
		GroupInfo: map[string]*Group{
			"root1": {ID: 1, Name: "root1"},
		},
		Subgroups: map[string][]Group{
			"1": {},
		},
		FetchError: errors.New("intentional projects fetch error"),
	}
	errorChannel := make(chan error, 10)
	projectCounter := counter.NewCounter()
	groupCounter := counter.NewCounter()

	channeledApi := NewChanneledApi(fakeAPI, &GitLabConfig{}, projectCounter, groupCounter, errorChannel)

	groupConfigs := []GroupConfig{
		{Name: "root1"},
	}

	channeledApi.ScheduleFetchGitlabGroupProjects(groupConfigs)

	expectErrors(
		t,
		errorChannel,
		[]string{"failed to fetch rootGroupConfig info for rootGroupConfig root1: intentional projects fetch error"},
	)
}

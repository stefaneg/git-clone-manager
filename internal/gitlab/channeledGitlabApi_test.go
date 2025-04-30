package gitlab

import (
	"errors"
	"fmt"
	"gcm/internal/counter"
	"gcm/internal/ext"
	"gcm/internal/fs"
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

	fakeAPI := FakeSetupTwoRootsFourSubgroupsEightProjects()
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

	filesystem := fs.RealFs{}
	channeledApi := NewChanneledApi(fakeAPI, &filesystem, config, projectCounter, groupCounter, errorChannel)

	groupConfigs := []GroupConfig{
		{Name: "root1"},
		{Name: "root2"},
	}

	projectChannel := channeledApi.ChannelGroupProjects(groupConfigs)
	expectOpenChannel(t, projectChannel)
	var projects []Project
	for project := range projectChannel {
		projects = append(projects, project)
	}

	expectProjectCount(t, projects, projectCounter, 8)
	expectGroupCount(t, groupCounter, 6) // Roots plus subgroups
}

func expectOpenChannel(t *testing.T, channel <-chan Project) {
	if ext.IsChannelClosed(channel) {
		t.Errorf("expected channel to be open")
	}
}

func expectClosedChannel(t *testing.T, channel <-chan Project) {
	if !ext.IsChannelClosed(channel) {
		t.Errorf("expected channel to be closed")
	}
}

func TestChanneledApi_ErrorHandling_FetchGroupInfo(t *testing.T) {
	fakeAPI := &FakeAPI{
		FetchError: errors.New("failed to fetch group info"),
	}
	errorChannel := make(chan error, 10)
	projectCounter := counter.NewCounter()
	groupCounter := counter.NewCounter()

	fileSystem := &fs.FakeFileSystem{}

	channeledApi := NewChanneledApi(fakeAPI, fileSystem, &GitLabConfig{}, projectCounter, groupCounter, errorChannel)

	groupConfigs := []GroupConfig{
		{Name: "nonexistent-group"},
	}
	projectChannel := channeledApi.ChannelGroupProjects(groupConfigs)

	for project := range projectChannel {
		t.Errorf("Not expecting any projects but got %v", project)
	}

	expectErrors(
		t,
		errorChannel,
		[]string{"failed to fetch group info for root group nonexistent-group: failed to fetch group info"},
	)
}

func TestChanneledApi_ErrorHandling_FetchSubgroups(t *testing.T) {
	// Simulate an error when fetching subgroups
	fakeAPI := &FakeAPI{
		GroupInfo: map[string]*Group{
			"root1": {ID: 1, Name: "root1"},
		},
		Projects: map[string][]Project{
			"1": {},
		},
	}
	errorChannel := make(chan error, 10)
	projectCounter := counter.NewCounter()
	groupCounter := counter.NewCounter()

	fakeFs := &fs.FakeFileSystem{}
	channeledApi := NewChanneledApi(fakeAPI, fakeFs, &GitLabConfig{}, projectCounter, groupCounter, errorChannel)

	groupConfigs := []GroupConfig{
		{Name: "root1"},
	}

	projectChan := channeledApi.ChannelGroupProjects(groupConfigs)
	for project := range projectChan {
		t.Errorf("Not expecting any projects but got %v", project)
	}

	expectErrors(
		t,
		errorChannel,
		[]string{"failed to fetch subgroups for group 1: Fake API request on /groups/1/subgroups failed with status: 404 Not Found"},
	)
	// FLAKY TEST
	expectClosedChannel(t, projectChan)
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
	}
	errorChannel := make(chan error, 10)
	projectCounter := counter.NewCounter()
	groupCounter := counter.NewCounter()

	fakeFs := &fs.FakeFileSystem{}
	channeledApi := NewChanneledApi(fakeAPI, fakeFs, &GitLabConfig{}, projectCounter, groupCounter, errorChannel)

	groupConfigs := []GroupConfig{
		{Name: "root1"},
	}

	projectChannel := channeledApi.ChannelGroupProjects(groupConfigs)

	for project := range projectChannel {
		t.Errorf("Not expecting any projects but got %v", project)
	}
	expectErrors(
		t,
		errorChannel,
		[]string{"failed to fetch projects for group root1: Fake API request on /groups/1/projects failed with status: 404 Not Found"},
	)

	if !ext.IsChannelClosed(projectChannel) {
		t.Errorf("Expected project channel to be closed")
	}

}

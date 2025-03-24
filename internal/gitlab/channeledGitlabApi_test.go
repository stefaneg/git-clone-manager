package gitlab

import (
	"gcm/internal/counter"
	"gcm/internal/gitremote"
	"testing"
)

func TestChanneledApi_FetchAndChannelGroupProjects(t *testing.T) {
	fakeAPI := &FakeAPI{
		GroupInfo: map[string]*Group{
			"root": {ID: 1, Name: "root"},
		},
		Subgroups: map[string][]Group{
			"1": {
				{ID: 2, Name: "subgroup1"},
				{ID: 3, Name: "subgroup2"},
			},
		},
		Projects: map[string][]Project{
			"2": {
				{Name: "subgroup1, project1", SSHURLToRepo: "git@somewhere:project1.git"},
				{Name: "subgroup1, project2", SSHURLToRepo: "git@somewhere:project2.git"},
			},
			"3": {
				{Name: "subgroup2, project3", SSHURLToRepo: "git@somewhere:project3.git"},
				{Name: "subgroup2, project4", SSHURLToRepo: "git@somewhere:project4.git"},
			},
		},
	}
	projectCounter := counter.NewCounter()
	groupCounter := counter.NewCounter()
	errorChannel := make(chan error, 10)
	config := &GitLabConfig{}

	channeledApi := NewChanneledApi(fakeAPI, config, projectCounter, groupCounter, errorChannel)

	rootGroupConfig := &GroupConfig{Name: "root"}
	projectChannel := channeledApi.FetchAndChannelGroupProjects(rootGroupConfig)

	var projects []Project
	for project := range projectChannel {
		projects = append(projects, project)
	}

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}

	if projectCounter.Count() != 2 {
		t.Fatalf("expected project counter to be 2, got %d", projectCounter.Count())
	}

	if groupCounter.Count() != 3 {
		t.Fatalf("expected group counter to be 3, got %d", groupCounter.Count())
	}

	if len(errorChannel) != 0 {
		t.Fatalf("expected no errors, got %d", len(errorChannel))
	}
}

func TestChanneledApi_ScheduleFetchGitlabGroupProjects(t *testing.T) {
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
			"4": {
				{ID: 5, Name: "subgroup3"},
				{ID: 6, Name: "subgroup4"},
			},
		},
		Projects: map[string][]Project{
			"2": {
				{Name: "subgroup1, project1", SSHURLToRepo: "git@somewhere:project1.git"},
				{Name: "subgroup1, project2", SSHURLToRepo: "git@somewhere:project2.git"},
			},
			"3": {
				{Name: "subgroup2, project3", SSHURLToRepo: "git@somewhere:project3.git"},
				{Name: "subgroup2, project4", SSHURLToRepo: "git@somewhere:project4.git"},
			},
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
	config := &GitLabConfig{
		Projects: []gitremote.ProjectConfig{
			{Name: "project1"},
			{Name: "project2"},
		},
	}

	channeledApi := NewChanneledApi(fakeAPI, config, projectCounter, groupCounter, errorChannel)

	groupConfigs := []GroupConfig{
		{Name: "root"},
	}

	projectChannel := channeledApi.ScheduleFetchGitlabGroupProjects(groupConfigs)

	var projects []Project
	for project := range projectChannel {
		projects = append(projects, project)
	}

	if len(projects) != 4 {
		t.Fatalf("expected 4 projects, got %d", len(projects))
	}

	if projectCounter.Count() != 4 {
		t.Fatalf("expected project counter to be 2, got %d", projectCounter.Count())
	}

	if groupCounter.Count() != 2 {
		t.Fatalf("expected group counter to be 2, got %d", groupCounter.Count())
	}

	if len(errorChannel) != 0 {
		t.Fatalf("expected no errors, got %d", len(errorChannel))
	}
}

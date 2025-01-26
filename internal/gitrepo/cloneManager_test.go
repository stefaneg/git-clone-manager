package gitrepo

import (
	"fmt"
	"gcm/internal/counter"
	"testing"
)

type MockGitRepo struct {
	name                 string
	archived             bool
	cloneOptions         CloneOptions
	needsCloning         bool
	isCloned             bool
	cloneError           error
	isClonedError        error
	checkCloneErr        error
	markerHasBeenWritten bool
}

func (m *MockGitRepo) WriteArchivedMarker(_ string) error {
	m.markerHasBeenWritten = true
	return nil
}

func (m *MockGitRepo) GetName() string {
	return m.name
}

func (m *MockGitRepo) IsArchived() bool {
	return m.archived
}

func (m *MockGitRepo) GetCloneOptions() CloneOptions {
	return m.cloneOptions
}

func (m *MockGitRepo) CheckNeedsCloning() (bool, error) {
	return m.needsCloning, m.checkCloneErr
}

func (m *MockGitRepo) IsCloned() (bool, error) {
	return m.isCloned, m.isClonedError
}

func (m *MockGitRepo) Clone() error {
	return m.cloneError
}

type MockCloneOptions struct {
	cloneArchived bool
}

func (m MockCloneOptions) CloneArchived() bool {
	return m.cloneArchived
}

func (m MockCloneOptions) CloneRootDirectory() string {
	return "faking/it/somewhere"
}

func TestFilterCloneNeeded(t *testing.T) {
	archivedCounter := counter.NewCounter()
	clonedCounter := counter.NewCounter()
	errorChannel := make(chan error, 10)
	defer close(errorChannel)

	repos := []GitRepo{
		&MockGitRepo{
			name:         "repo1",
			archived:     false,
			needsCloning: true,
			isCloned:     false,
			cloneOptions: MockCloneOptions{},
		},
		&MockGitRepo{
			name:         "repo2",
			archived:     true,
			needsCloning: false,
			isCloned:     true,
			cloneOptions: MockCloneOptions{},
		},
		&MockGitRepo{
			name:         "cloneArchivedRepo",
			archived:     true,
			needsCloning: false,
			isCloned:     true,
			cloneOptions: MockCloneOptions{
				cloneArchived: true,
			},
		},
		&MockGitRepo{
			name:         "repo3",
			archived:     false,
			needsCloning: false,
			isCloned:     false,
			cloneOptions: MockCloneOptions{},
		},
	}

	repoChannel := make(chan GitRepo, len(repos))
	for _, repo := range repos {
		repoChannel <- repo
	}
	close(repoChannel)

	filteredChannel := FilterCloneNeeded(repoChannel, archivedCounter, clonedCounter, errorChannel)

	var filteredRepos []GitRepo
	for repo := range filteredChannel {
		filteredRepos = append(filteredRepos, repo)
	}

	if len(filteredRepos) != 1 {
		t.Errorf("expected 1 repo to be cloned, got %d", len(filteredRepos))
	}

	if filteredRepos[0].GetName() != "repo1" {
		t.Errorf("expected repo1 to be cloned, got %s", filteredRepos[0].GetName())
	}

	if archivedCounter.Count() != 1 {
		t.Errorf("expected 1 archived repo, got %d", archivedCounter.Count())
	}

	if clonedCounter.Count() != 3 {
		t.Errorf("expected 3 cloned repos, got %d", clonedCounter.Count())
	}
}

func TestFilterCloneNeeded_ErrorHandling(t *testing.T) {
	archivedCounter := counter.NewCounter()
	clonedCounter := counter.NewCounter()
	errorChannel := make(chan error, 10)
	defer close(errorChannel)

	repos := []GitRepo{
		&MockGitRepo{
			name:          "errorCheckingCloning",
			archived:      false,
			needsCloning:  false,
			isCloned:      false,
			isClonedError: fmt.Errorf("Fake error checking for cloned status"),
			cloneOptions:  MockCloneOptions{},
		},
		&MockGitRepo{
			name:          "errorCheckingNeedsCloning",
			archived:      false,
			needsCloning:  false,
			isCloned:      false,
			checkCloneErr: fmt.Errorf("Fake error checking if needs cloning"),
			cloneOptions:  MockCloneOptions{},
		},
	}

	repoChannel := make(chan GitRepo, len(repos))
	for _, repo := range repos {
		repoChannel <- repo
	}
	close(repoChannel)

	filteredChannel := FilterCloneNeeded(repoChannel, archivedCounter, clonedCounter, errorChannel)

	var filteredRepos []GitRepo
	for repo := range filteredChannel {
		filteredRepos = append(filteredRepos, repo)
	}

	if len(errorChannel) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errorChannel))
	}

	expectedErrors := []string{
		"error checking clone status errorCheckingCloning: Fake error checking for cloned status",
		"error checking if project needs cloning errorCheckingNeedsCloning: Fake error checking if needs cloning",
	}

	for _, expectedError := range expectedErrors {
		select {
		case err := <-errorChannel:
			if err.Error() != expectedError {
				t.Errorf("expected error %s, got %s", expectedError, err.Error())
			}
		default:
			t.Errorf("expected error %s, but no error found", expectedError)
		}
	}
}

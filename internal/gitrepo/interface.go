package gitrepo

type GitRepo interface {
	GetName() string
	Clone() error
	CheckNeedsCloning() (bool, error)
	IsCloned() (bool, error)
	WriteArchivedMarker(projectPath string) error
	IsArchived() bool
	GetCloneOptions() CloneOptions
}

type CloneOptions interface {
	CloneArchived() bool
	CloneRootDirectory() string
	// ... add project metadata
}

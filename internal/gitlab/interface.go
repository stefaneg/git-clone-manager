package gitlab

type API interface {
	FetchGroupProjects(group *Group) ([]Project, error)
	FetchSubgroups(groupID string) ([]Group, error)
	FetchGroupInfo(groupName string) (*Group, error)
	FetchAccessibleProjects() ([]Project, error)
}

type APIFactory func(token, hostName string) API

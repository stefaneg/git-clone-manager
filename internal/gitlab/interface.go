package gitlab

type API interface {
	FetchProjects(group *Group) ([]Project, error)
	FetchSubgroups(groupID string) ([]Group, error)
	FetchGroupInfo(groupName string) (*Group, error)
}

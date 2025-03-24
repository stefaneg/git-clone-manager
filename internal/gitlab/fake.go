package gitlab

import "fmt"

type FakeAPI struct {
	Projects   map[string][]Project // GroupID -> Projects
	Subgroups  map[string][]Group   // GroupID -> Subgroups
	GroupInfo  map[string]*Group    // GroupName -> GroupInfo
	FetchError error
}

func (api *FakeAPI) FetchProjects(group *Group) ([]Project, error) {
	if api.FetchError != nil {
		return nil, api.FetchError
	}
	return api.Projects[fmt.Sprintf("%d", group.ID)], nil
}

func (api *FakeAPI) FetchSubgroups(groupID string) ([]Group, error) {
	if api.FetchError != nil {
		return nil, api.FetchError
	}
	return api.Subgroups[groupID], nil
}

func (api *FakeAPI) FetchGroupInfo(groupName string) (*Group, error) {
	if api.FetchError != nil {
		return nil, api.FetchError
	}
	return api.GroupInfo[groupName], nil
}

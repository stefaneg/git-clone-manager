package gitlab

import "fmt"

type FakeAPI struct {
	Token      string
	Hostname   string
	Projects   map[string][]Project // GroupID -> Projects
	Subgroups  map[string][]Group   // GroupID -> Subgroups
	GroupInfo  map[string]*Group    // GroupName -> GroupInfo
	FetchError error
}

func (api *FakeAPI) FetchAccessibleProjects() ([]Project, error) {
	//TODO implement me
	panic("implement me")
}

func (api *FakeAPI) FetchGroupProjects(group *Group) ([]Project, error) {
	if api.FetchError != nil {
		return nil, api.FetchError
	}
	projects, exists := api.Projects[fmt.Sprintf("%d", group.ID)]
	if !exists {
		return nil, fmt.Errorf("Fake API request on /groups/%d/projects failed with status: 404 Not Found", group.ID)
	}
	return projects, nil
}

func (api *FakeAPI) FetchSubgroups(groupID string) ([]Group, error) {
	if api.FetchError != nil {
		return nil, api.FetchError
	}
	groups, exists := api.Subgroups[groupID]
	if !exists {
		return nil, fmt.Errorf("Fake API request on /groups/%s/subgroups failed with status: 404 Not Found", groupID)
	}
	return groups, nil
}

func (api *FakeAPI) FetchGroupInfo(groupName string) (*Group, error) {
	if api.FetchError != nil {
		return nil, api.FetchError
	}
	group, exists := api.GroupInfo[groupName]
	if !exists {
		return nil, fmt.Errorf("Fake API request on /groups/%s failed with status: 404 Not Found", groupName)
	}
	return group, nil
}

func FakeApiFactory(token, hostName string) API {
	return &FakeAPI{
		Token:    token,
		Hostname: hostName,
	}
}

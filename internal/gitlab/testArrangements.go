package gitlab

func FakeSetupTwoRootsFourSubgroupsEightProjects() *FakeAPI {
	return &FakeAPI{
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
}

func FakeSetupNoSubgroupsOneProject() *FakeAPI {
	return &FakeAPI{
		GroupInfo: map[string]*Group{
			"root1": {ID: 1, Name: "root1"},
		},
		Subgroups: map[string][]Group{
			"1": {},
		},
		Projects: map[string][]Project{
			"1": {
				{Name: "project1", SSHURLToRepo: "git@somewhere:project1.git"},
			},
		},
	}
}

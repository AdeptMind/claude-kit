package main

type RoleInfo struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type RoleService struct{}

func (s *RoleService) GetCurrent() RoleInfo {
	return RoleInfo{
		Name:  "Architect",
		Color: "gold",
	}
}

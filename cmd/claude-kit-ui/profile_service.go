package main

type ProfileInfo struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type ProfileService struct{}

func (s *ProfileService) List() []ProfileInfo {
	return []ProfileInfo{
		{Name: "default", Active: true},
	}
}

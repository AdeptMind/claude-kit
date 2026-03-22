package main

type FileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Children []FileNode `json:"children,omitempty"`
	Expanded bool       `json:"expanded,omitempty"`
}

type FileService struct{}

func (s *FileService) Tree() []FileNode {
	return []FileNode{
		{
			Name: ".claude", Path: ".claude", Expanded: true,
			Children: []FileNode{
				{Name: "CLAUDE.md", Path: ".claude/CLAUDE.md"},
			},
		},
	}
}

func (s *FileService) ReadPreview(path string) string {
	return "// Preview not yet implemented for " + path
}

func (s *FileService) OpenExternal(path string) error {
	// TODO: open file in external editor
	return nil
}

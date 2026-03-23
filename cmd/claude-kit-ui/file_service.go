package main

import (
	"os"
	"path/filepath"
	"sort"
)

type FileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Children []FileNode `json:"children,omitempty"`
	Expanded bool       `json:"expanded,omitempty"`
}

type FileService struct{}

func (s *FileService) Tree(projectPath string) []FileNode {
	if projectPath == "" {
		return nil
	}
	root := filepath.Clean(projectPath)
	if _, err := os.Stat(root); err != nil {
		return nil
	}
	return s.buildTree(root, root, 0)
}

func (s *FileService) buildTree(dir, root string, depth int) []FileNode {
	if depth > 4 {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var nodes []FileNode
	for _, e := range entries {
		name := e.Name()
		if name == ".git" || name == "node_modules" || name == ".DS_Store" {
			continue
		}
		rel, _ := filepath.Rel(root, filepath.Join(dir, name))
		node := FileNode{Name: name, Path: rel}
		if e.IsDir() {
			node.Children = s.buildTree(filepath.Join(dir, name), root, depth+1)
			node.Expanded = depth == 0
		}
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool {
		di := len(nodes[i].Children) > 0
		dj := len(nodes[j].Children) > 0
		if di != dj {
			return di
		}
		return nodes[i].Name < nodes[j].Name
	})
	return nodes
}

func (s *FileService) ReadPreview(projectPath, relPath string) string {
	if projectPath == "" || relPath == "" {
		return ""
	}
	full := filepath.Join(projectPath, relPath)
	data, err := os.ReadFile(full)
	if err != nil {
		return ""
	}
	content := string(data)
	if len(content) > 10000 {
		content = content[:10000] + "\n\n… (truncated)"
	}
	return content
}

func (s *FileService) RecentFiles(projectPath string) []map[string]string {
	if projectPath == "" {
		return nil
	}
	outputDir := filepath.Join(projectPath, ".claude", "output")
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil
	}

	type fileWithTime struct {
		name    string
		path    string
		modTime int64
	}
	var files []fileWithTime
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileWithTime{
			name:    e.Name(),
			path:    filepath.Join(".claude", "output", e.Name()),
			modTime: info.ModTime().Unix(),
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime > files[j].modTime
	})
	if len(files) > 10 {
		files = files[:10]
	}

	var result []map[string]string
	for _, f := range files {
		ext := filepath.Ext(f.name)
		if len(ext) > 0 {
			ext = ext[1:]
		}
		result = append(result, map[string]string{
			"name": f.name,
			"path": f.path,
			"type": ext,
		})
	}
	return result
}

func (s *FileService) OpenExternal(path string) error {
	// TODO: open file in external editor
	return nil
}

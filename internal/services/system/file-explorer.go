package system

import (
	"os"
	"path/filepath"
	systemServiceInterfaces "sylve/internal/interfaces/services/system"
)

func (s *Service) Traverse(path string) ([]systemServiceInterfaces.FileNode, error) {
	if path == "" {
		path = "/"
	}

	if !filepath.IsAbs(path) {
		path = "/" + path
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var nodes []systemServiceInterfaces.FileNode
	for _, e := range entries {
		full := filepath.Join(path, e.Name())
		info, err := e.Info()
		if err != nil {
			continue
		}

		node := systemServiceInterfaces.FileNode{
			ID:   full,
			Date: info.ModTime(),
		}
		if info.IsDir() {
			node.Type = "folder"
			node.Lazy = true
		} else {
			node.Type = "file"
			node.Size = info.Size()
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func (s *Service) AddFileOrFolder(path string, name string, isFolder bool) error {
	if path == "" {
		path = "/"
	}

	if !filepath.IsAbs(path) {
		path = "/" + path
	}

	if strings.Contains(name, "/") || name == "." || name == ".." {
		return fmt.Errorf("invalid name: %s", name)
	}

	fullPath := filepath.Join(path, name)

	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("file or folder already exists: %s", fullPath)
	} else if !os.IsNotExist(err) {
		return err
	}

	if isFolder {
		return os.Mkdir(fullPath, 0755)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}

	defer file.Close()

	return nil
}

func (s *Service) DeleteFileOrFolder(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if !filepath.IsAbs(path) {
		path = "/" + path
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file or folder does not exist: %s", path)
	}

	return os.RemoveAll(path)
}

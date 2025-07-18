package system

import (
	"fmt"
	"io/fs"
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

func (s *Service) RenameFileOrFolder(oldPath string, newName string) error {
	if oldPath == "" || newName == "" {
		return fmt.Errorf("old path and new name cannot be empty")
	}

	if !filepath.IsAbs(oldPath) {
		oldPath = filepath.Clean("/" + oldPath)
	}

	if strings.Contains(newName, "/") || newName == "." || newName == ".." {
		return fmt.Errorf("invalid new name: %s", newName)
	}

	newPath := filepath.Join(filepath.Dir(oldPath), newName)

	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("file or folder already exists: %s", newPath)
	} else if !os.IsNotExist(err) {
		return err
	}

	return os.Rename(oldPath, newPath)
}

func (s *Service) DownloadFile(id string) (string, error) {
	cleanPath := filepath.Clean(id)

	if !filepath.IsAbs(cleanPath) {
		return "", fmt.Errorf("path must be absolute")
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("cannot download a directory")
	}

	return cleanPath, nil
}

func copyFile(source, destination string, perm fs.FileMode) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	if err := os.WriteFile(destination, data, perm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func copyDir(sourceDir, destDir string) error {
	return filepath.Walk(sourceDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath, info.Mode())
	})
}

func (s *Service) CopyOrMoveFileOrFolder(source, destination string, move bool) error {
	if source == "" || destination == "" {
		return fmt.Errorf("source and destination cannot be empty")
	}
	if !filepath.IsAbs(source) || !filepath.IsAbs(destination) {
		return fmt.Errorf("both source and destination must be absolute paths")
	}

	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("source does not exist: %w", err)
	}

	if destInfo, err := os.Stat(destination); err == nil && destInfo.IsDir() {
		destination = filepath.Join(destination, filepath.Base(source))
	}

	if move {
		if err := os.Rename(source, destination); err != nil {
			return fmt.Errorf("failed to move: %w", err)
		}
		return nil
	}

	if info.IsDir() {
		return copyDir(source, destination)
	}

	return copyFile(source, destination, info.Mode())
}

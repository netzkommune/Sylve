package zfs

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type ZDB struct {
	z        *zfs            `json:"-"`
	Name     string          `json:"name"`
	GUID     uint64          `json:"guid"`
	Version  string          `json:"version"`
	Children []ZdbZPoolChild `json:"children,omitempty"`
}

type ZdbZPoolChild struct {
	Type          string            `json:"type,omitempty"`
	ID            int               `json:"id,omitempty"`
	GUID          uint64            `json:"guid,omitempty"`
	Path          string            `json:"path,omitempty"`
	WholeDisk     int               `json:"whole_disk,omitempty"`
	MetaslabArray int               `json:"metaslab_array,omitempty"`
	MetaslabShift int               `json:"metaslab_shift,omitempty"`
	Ashift        int               `json:"ashift,omitempty"`
	Asize         uint64            `json:"asize,omitempty"`
	IsLog         int               `json:"is_log,omitempty"`
	CreateTXG     uint64            `json:"create_txg,omitempty"`
	Properties    map[string]string `json:"properties,omitempty"`
	Children      []ZdbZPoolChild   `json:"children,omitempty"`
}

func (z *zfs) zdbOutput(arg ...string) ([][]string, error) {
	return z.run(nil, nil, "zdb", arg...)
}

func (z *ZDB) parseLine(line []string) error {
	prop := line[0]
	val := line[1]

	switch prop {
	case "version":
		z.Version = val
	case "name":
		z.Name = val
	}
	return nil
}

func (c *ZdbZPoolChild) parseLine(indentation int, line []string) error {
	prop := strings.TrimSuffix(line[0], ":")
	val := strings.TrimSpace(line[1])

	if c.Properties == nil {
		c.Properties = make(map[string]string)
	}
	c.Properties[prop] = val

	switch prop {
	case "type":
		c.Type = val
	case "id":
		_, err := fmt.Sscan(val, &c.ID)
		if err != nil {
			return fmt.Errorf("failed to parse id: %w", err)
		}
	case "guid":
		_, err := fmt.Sscan(val, &c.GUID)
		if err != nil {
			return fmt.Errorf("failed to parse guid: %w", err)
		}
	case "path":
		c.Path = val
	case "whole_disk":
		_, err := fmt.Sscan(val, &c.WholeDisk)
		if err != nil {
			return fmt.Errorf("failed to parse whole_disk: %w", err)
		}
	case "metaslab_array":
		_, err := fmt.Sscan(val, &c.MetaslabArray)
		if err != nil {
			return fmt.Errorf("failed to parse metaslab_array: %w", err)
		}
	case "metaslab_shift":
		_, err := fmt.Sscan(val, &c.MetaslabShift)
		if err != nil {
			return fmt.Errorf("failed to parse metaslab_shift: %w", err)
		}
	case "ashift":
		_, err := fmt.Sscan(val, &c.Ashift)
		if err != nil {
			return fmt.Errorf("failed to parse ashift: %w", err)
		}
	case "asize":
		_, err := fmt.Sscan(val, &c.Asize)
		if err != nil {
			return fmt.Errorf("failed to parse asize: %w", err)
		}
	case "is_log":
		_, err := fmt.Sscan(val, &c.IsLog)
		if err != nil {
			return fmt.Errorf("failed to parse is_log: %w", err)
		}
	case "create_txg":
		_, err := fmt.Sscan(val, &c.CreateTXG)
		if err != nil {
			return fmt.Errorf("failed to parse create_txg: %w", err)
		}
	}
	return nil
}

type zdbCacheEntry struct {
	zdb    *ZDB
	guid   uint64
	expiry time.Time
}

var zdbCache = make(map[string]zdbCacheEntry)
var zdbCacheMutex sync.RWMutex
var cacheTTL = 5 * time.Minute

func (z *zfs) GetZdb(name string) (*ZDB, error) {
	currentGUID, err := z.GetZpoolGUID(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current zpool guid: %w", err)
	}

	zdbCacheMutex.RLock()
	if entry, ok := zdbCache[name]; ok && time.Now().Before(entry.expiry) && entry.guid == currentGUID {
		zdbCacheMutex.RUnlock()
		return entry.zdb, nil
	}
	zdbCacheMutex.RUnlock()

	args := zdbArgs
	args = append(args, name)

	out, err := z.zdbOutput(args...)
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no output from zdb")
	}

	zdb := &ZDB{z: z, Name: name, GUID: currentGUID, Children: make([]ZdbZPoolChild, 0)}
	var stack []*ZdbZPoolChild
	var currentChild *ZdbZPoolChild

	for _, rawLine := range out {
		if len(rawLine) < 2 {
			continue
		}

		line := make([]string, len(rawLine))
		copy(line, rawLine)

		indentation := 0
		for _, r := range line[0] {
			if r == '\t' {
				indentation++
			} else {
				break
			}
		}
		line[0] = strings.TrimLeft(line[0], "\t ")

		propParts := strings.SplitN(line[0], ":", 2)
		if len(propParts) < 2 {
			continue
		}
		prop := strings.TrimSpace(propParts[0])
		val := strings.TrimSpace(propParts[1])
		currentLine := []string{prop, val}

		if prop == "version" && len(stack) == 0 {
			zdb.parseLine(currentLine)
		} else if prop == "name" && len(stack) == 0 {
			zdb.parseLine(currentLine)
		} else if prop == "type" {
			newChild := ZdbZPoolChild{}
			newChild.Type = val
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, newChild)
				stack = append(stack, &parent.Children[len(parent.Children)-1])
				currentChild = &parent.Children[len(parent.Children)-1]
			} else {
				zdb.Children = append(zdb.Children, newChild)
				stack = []*ZdbZPoolChild{&zdb.Children[len(zdb.Children)-1]}
				currentChild = &zdb.Children[len(zdb.Children)-1]
			}
			currentChild.parseLine(indentation, currentLine)
		} else if len(stack) > 0 && currentChild != nil {
			// Check if the indentation indicates we're still within the current vdev
			currentIndentation := 0
			for _, r := range rawLine[0] {
				if r == '\t' {
					currentIndentation++
				} else {
					break
				}
			}
			if currentIndentation == indentation {
				currentChild.parseLine(indentation, currentLine)
			} else if currentIndentation < indentation && len(stack) > 1 {
				stack = stack[:len(stack)-1]
				currentChild = stack[len(stack)-1]
				currentChild.parseLine(currentIndentation, currentLine)
			} else if currentIndentation > indentation {
				// This should ideally be handled when the 'type' is encountered.
			}
		}
	}

	zdbCacheMutex.Lock()
	zdbCache[name] = zdbCacheEntry{zdb: zdb, guid: currentGUID, expiry: time.Now().Add(cacheTTL)}
	zdbCacheMutex.Unlock()

	return zdb, nil
}

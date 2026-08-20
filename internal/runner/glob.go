package runner

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// ErrNoMatches distinguishes an unchecked file set from a successful empty scan.
var ErrNoMatches = errors.New("runner: file glob matched no files")

// MatchFiles returns tree-relative files matching pattern without traversing .git or symlinks.
func MatchFiles(root, pattern string) ([]string, error) {
	canonicalRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return nil, fmt.Errorf("canonicalize tree root %s: %w", root, err)
	}

	matches := make([]string, 0)
	err = doublestar.GlobWalk(
		treeFS{root: canonicalRoot},
		filepath.ToSlash(pattern),
		func(path string, entry fs.DirEntry) error {
			if entry.Type()&fs.ModeSymlink == 0 {
				matches = append(matches, path)
			}
			return nil
		},
		doublestar.WithFilesOnly(),
		doublestar.WithNoFollow(),
		doublestar.WithFailOnIOErrors(),
	)
	if err != nil {
		return nil, fmt.Errorf("match file glob %q: %w", pattern, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("match file glob %q: %w", pattern, ErrNoMatches)
	}
	sort.Strings(matches)
	return matches, nil
}

type treeFS struct {
	root string
}

func (tree treeFS) Open(name string) (fs.File, error) {
	path, err := tree.checkedPath(name)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open tree path %s: %w", name, err)
	}
	return file, nil
}

func (tree treeFS) ReadDir(name string) ([]fs.DirEntry, error) {
	path, err := tree.checkedPath(name)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read tree directory %s: %w", name, err)
	}
	visible := make([]fs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Name() != ".git" {
			visible = append(visible, entry)
		}
	}
	return visible, nil
}

func (tree treeFS) checkedPath(name string) (string, error) {
	if !fs.ValidPath(name) {
		return "", &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	path := tree.root
	for _, component := range strings.Split(name, "/") {
		if component == "." {
			continue
		}
		if component == ".git" {
			return "", &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
		}
		path = filepath.Join(path, component)
		info, err := os.Lstat(path)
		if err != nil {
			return "", fmt.Errorf("inspect tree path %s: %w", name, err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return "", &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
		}
	}
	return path, nil
}

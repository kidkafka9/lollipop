package gitrepo

import (
	"io/fs"
	"path/filepath"
)

// Find returns an array of any folder in a given root that contains .git (returns all git folders in a specific folder)
func Find(root string) (repos []string, skipped []error, err error) {
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if path == root {
				return walkErr
			}
			skipped = append(skipped, walkErr)
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		if d.Name() == ".git" {
			repoDir := filepath.Dir(path)
			repos = append(repos, repoDir)
			// skip going into .git
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return repos, skipped, nil
}

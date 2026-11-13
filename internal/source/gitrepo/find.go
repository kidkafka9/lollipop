package gitrepo

import (
	"io/fs"
	"path/filepath"
)

// Find returns every folder under root, root included, that holds a git
// repository (a folder with a .git folder in it). It searches at any depth,
// finds repositories nested inside other repositories, but does not look
// inside .git folders.
//
// Folders that can't be read are skipped and their errors returned in
// skipped, so the caller can tell the user. err is set only when root itself
// can't be read; repos and skipped are nil then.
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
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return repos, skipped, nil
}

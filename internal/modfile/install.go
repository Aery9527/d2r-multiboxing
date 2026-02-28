package modfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// D2RModsDir returns the mods directory path relative to the D2R executable.
func D2RModsDir(d2rPath string) string {
	return filepath.Join(filepath.Dir(d2rPath), "mods")
}

// InstallMod copies a mod directory to the D2R mods directory.
// srcModDir is the full path to the source mod (e.g., ./mods/d2r-hyper-show).
// d2rPath is the full path to D2R.exe.
func InstallMod(srcModDir, d2rPath string) error {
	modName := filepath.Base(srcModDir)
	dstDir := filepath.Join(D2RModsDir(d2rPath), modName)

	if err := copyDir(srcModDir, dstDir); err != nil {
		return fmt.Errorf("failed to install mod %s: %w", modName, err)
	}

	return nil
}

// DiscoverInstalledMods scans the D2R mods directory for installed mods.
func DiscoverInstalledMods(d2rPath string) ([]string, error) {
	return DiscoverMods(D2RModsDir(d2rPath))
}

// copyDir recursively copies a directory tree.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

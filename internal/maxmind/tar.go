package maxmind

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/walle/targz"
)

func (s *Service) unTarV3(ctx context.Context, dbType string) error {
	//tmpDir := os.TempDir()

	err := targz.Extract(s.cfg.IPService.MaxMind.ArchiveFilePath(dbType), s.cfg.IPService.MaxMind.BaseFolder)
	if err != nil {
		s.Log.Error(err, "targz extract failed")
		return err
	}

	// Walk through the extracted files to find the .mmdb file, since maxmind names the folder with a version number
	target := fmt.Sprintf("GeoLite2-%s.mmdb", dbType)
	var srcPath string
	err = filepath.Walk(s.cfg.IPService.MaxMind.BaseFolder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() && info.Name() == "skip" {
			return filepath.SkipDir
		}
		// Skip symlinks to avoid TOCTOU traversal via the walked path.
		if info.Mode()&fs.ModeSymlink != 0 {
			return nil
		}
		if info.Mode().IsRegular() && info.Name() == target {
			srcPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		s.Log.Error(err, "walk tmp dir failed")
		return err
	}
	if srcPath == "" {
		return fmt.Errorf("extracted db file %q not found", target)
	}

	f, err := os.Open(filepath.Clean(srcPath))
	if err != nil {
		s.Log.Error(err, "open extracted db file failed")
		return err
	}
	defer f.Close()

	dest, err := os.Create(filepath.Clean(s.cfg.IPService.MaxMind.DBFilePath(dbType)))
	if err != nil {
		s.Log.Error(err, "create dest db file failed")
		return err
	}
	defer dest.Close()

	if _, err = io.Copy(dest, f); err != nil {
		s.Log.Error(err, "copy extracted db file failed")
		return err
	}

	return nil
}

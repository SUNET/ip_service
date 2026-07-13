package maxmind

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
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
	err = filepath.Walk(s.cfg.IPService.MaxMind.BaseFolder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.Name() == fmt.Sprintf("GeoLite2-%s.mmdb", dbType) {
			f, err := os.Open(filepath.Clean(path))
			if err != nil {
				s.Log.Error(err, "open extracted db file failed")
				return err
			}
			//dest, err := os.Create("/tmp/db/GeoLite2-" + dbType + ".mmdb")
			dest, err := os.Create(filepath.Clean(s.cfg.IPService.MaxMind.DBFilePath(dbType)))
			if err != nil {
				s.Log.Error(err, "create dest db file failed")
				return err
			}
			_, err = io.Copy(dest, f)
			if err != nil {
				s.Log.Error(err, "copy extracted db file failed")
				return err
			}
		}
		if info.IsDir() && info.Name() == "skip" {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		s.Log.Error(err, "walk tmp dir failed")
		return err
	}

	return nil
}

func (s *Service) unTAR(ctx context.Context, dbType string) error {
	s.Log.Debug("entering unTAR")

	_, span := s.TP.Start(ctx, "maxmind:unTAR")
	defer span.End()

	if !s.cfg.IPService.MaxMind.IsArchivePresent(dbType) {
		return errors.New("archive file missing")
	}

	s.Log.Info("Unpacking tar archive", "dbType", dbType)

	fileName := s.cfg.IPService.MaxMind.ArchiveFilePath(dbType)
	s.Log.Debug("unTAR", "fileName", fileName)

	archiveFile, err := os.Open(filepath.Clean(fileName))
	if err != nil {
		s.Log.Error(err, "open archive file")
		return err
	}

	gzr, err := gzip.NewReader(archiveFile)
	if err != nil {
		s.Log.Error(err, "gzip new reader")
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	s.Log.Debug("untarring", "dbType", dbType)
	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			s.Log.Debug("end of tar archive")
			return nil
		case err != nil:
			s.Log.Error(err, "tar error")
			return err
		case header == nil:
			s.Log.Debug("tar header is nil")
			continue
		case filepath.Base(header.Name) != fmt.Sprintf("GeoLite2-%s.mmdb", dbType):
			continue
		}

		target := s.cfg.IPService.MaxMind.DBFilePath(dbType)
		s.Log.Debug("untar", "target", target)

		switch header.Typeflag {
		case tar.TypeDir:
			s.Log.Error(errors.New("is a dir, needs to be a file"), "unTAR")
			if _, err := os.Stat(header.Name); err != nil {
				if err := os.MkdirAll(target, 0750); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			file, err := os.OpenFile(filepath.Clean(target), os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.CopyN(file, tr, 90000000); err != nil {
				return err
			}

			if err := file.Close(); err != nil {
				return err
			}
		}
	}
}

func (s *Service) cleanUpTarArchive(ctx context.Context, dbType string) error {
	_, span := s.TP.Start(ctx, "maxmind:cleanUpTarArchive")
	defer span.End()

	s.Log.Info("Cleaning up tar archive", "dbType", dbType)

	return os.Remove(s.cfg.IPService.MaxMind.ArchiveFilePath(dbType))
}

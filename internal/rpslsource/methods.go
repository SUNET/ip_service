package rpslsource

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"ip_service/pkg/rpsl"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
)

func (s *Service) getRemoteSerial(ctx context.Context) (string, error) {
	switch s.sourceCfg.Transport {
	case TransportFTP:
		return s.getRemoteSerialFTP(ctx)
	case TransportHTTP:
		return s.getRemoteSerialHTTP(ctx)
	default:
		return "", fmt.Errorf("unknown transport type: %d", s.sourceCfg.Transport)
	}
}

func (s *Service) getRemoteSerialFTP(ctx context.Context) (string, error) {
	c, err := ftp.Dial(s.sourceCfg.Host,
		ftp.DialWithTimeout(60*time.Second),
		ftp.DialWithContext(ctx),
	)
	if err != nil {
		return "", err
	}

	if err := c.Login("anonymous", "anonymous"); err != nil {
		return "", err
	}

	r, err := c.Retr(s.sourceCfg.SerialPath)
	if err != nil {
		s.log.Error(err, "ftp retr failed")
		return "", err
	}
	defer r.Close()

	body, err := io.ReadAll(r)
	if err != nil {
		s.log.Error(err, "read all failed")
		return "", err
	}
	s.log.Debug("Remote serial", "source", s.sourceCfg.Name, "serial", string(body))

	return string(body), nil
}

func (s *Service) getRemoteSerialHTTP(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.sourceCfg.Host+s.sourceCfg.SerialPath, nil)
	if err != nil {
		return "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == 429 {
			return "", errors.New("remote rate limit exceeded")
		}
		return "", fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (s *Service) downloadArchive(ctx context.Context, remoteFile RemoteFile) error {
	s.log.Info("Downloading", "source", s.sourceCfg.Name, "file", remoteFile.Name)
	if !s.rateLimit.Allow() {
		return fmt.Errorf("rate limit exceeded")
	}

	switch s.sourceCfg.Transport {
	case TransportFTP:
		return s.downloadArchiveFTP(ctx, remoteFile)
	case TransportHTTP:
		return s.downloadArchiveHTTP(ctx, remoteFile)
	default:
		return fmt.Errorf("unknown transport type: %d", s.sourceCfg.Transport)
	}
}

func (s *Service) downloadArchiveFTP(ctx context.Context, remoteFile RemoteFile) error {
	c, err := ftp.Dial(s.sourceCfg.Host,
		ftp.DialWithTimeout(60*time.Second),
		ftp.DialWithContext(ctx),
	)
	if err != nil {
		s.log.Error(err, "ftp dial failed")
		return err
	}

	if err := c.Login("anonymous", "anonymous"); err != nil {
		return err
	}

	r, err := c.Retr(remoteFile.Path)
	if err != nil {
		s.log.Error(err, "ftp retr failed")
		return err
	}
	defer r.Close()

	body, err := io.ReadAll(r)
	if err != nil {
		s.log.Error(err, "read all failed")
		return err
	}

	absPath := s.archivePath(remoteFile.Name)
	f, err := os.OpenFile(filepath.Clean(absPath), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		s.log.Error(err, "open file")
		return err
	}
	defer f.Close()

	if _, err = f.Write(body); err != nil {
		s.log.Error(err, "write file")
		return err
	}

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	s.log.Debug("Downloaded", "file", remoteFile.Name, "size", stat.Size())

	return nil
}

func (s *Service) downloadArchiveHTTP(ctx context.Context, remoteFile RemoteFile) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.sourceCfg.Host+remoteFile.Path, nil)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.log.Error(err, "download failed", "file", remoteFile.Name)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return errors.New("remote rate limit exceeded")
	}

	absPath := s.archivePath(remoteFile.Name)
	outFile, err := os.OpenFile(filepath.Clean(absPath), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		s.log.Error(err, "open file")
		return err
	}
	defer outFile.Close()

	if _, err = io.Copy(outFile, resp.Body); err != nil {
		s.log.Error(err, "copy failed")
		return err
	}

	stat, err := outFile.Stat()
	if err != nil {
		return err
	}
	s.log.Info("Downloaded", "file", remoteFile.Name, "size", stat.Size())

	return nil
}

func (s *Service) unzip(ctx context.Context, dbType string) error {
	s.log.Info("Unzip", "source", s.sourceCfg.Name, "dbType", dbType)

	absPath := s.archivePath(dbType)
	f, err := os.Open(filepath.Clean(absPath))
	if err != nil {
		s.log.Error(err, "open file")
		return err
	}
	defer f.Close()

	reader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer reader.Close()

	localPath := s.localFilePath(dbType)
	outFile, err := os.Create(filepath.Clean(localPath))
	if err != nil {
		return err
	}
	defer outFile.Close()

	if _, err = io.Copy(outFile, reader); err != nil {
		return err
	}
	s.log.Info("File uncompressed", "path", localPath)

	return nil
}

func (s *Service) cleanupArchive(dbType string) error {
	return os.Remove(s.archivePath(dbType))
}

func (s *Service) addEOFMarker(dbType string) error {
	localPath := s.localFilePath(dbType)

	f, err := os.OpenFile(filepath.Clean(localPath), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("\n# EOF\n")
	return err
}

// archivePath returns the temp path for the compressed archive
func (s *Service) archivePath(dbType string) string {
	return fmt.Sprintf("/tmp/%s_db_%s.gz", s.sourceCfg.Name, dbType)
}

// localFilePath returns the persistent file path for the given dbType
func (s *Service) localFilePath(dbType string) string {
	return fmt.Sprintf("%s_%s.txt", s.sourceCfg.FilePath, dbType)
}

// loadFromLocal attempts to parse from local cached files. Returns true if successful.
func (s *Service) loadFromLocal(ctx context.Context) bool {
	for _, archive := range s.sourceCfg.RemoteFiles {
		localPath := s.localFilePath(archive.Name)
		if _, err := os.Stat(localPath); err != nil {
			return false
		}
	}

	rpslClient, err := rpsl.New(ctx)
	if err != nil {
		s.log.Error(err, "failed to create rpsl client")
		return false
	}

	for _, archive := range s.sourceCfg.RemoteFiles {
		localPath := s.localFilePath(archive.Name)
		if err := rpslClient.Parse(ctx, localPath); err != nil {
			s.log.Error(err, "failed to parse local cache", "source", s.sourceCfg.Name, "path", localPath)
			return false
		}
	}
	s.RPSLRouterClass = rpslClient.RouterClass
	return true
}

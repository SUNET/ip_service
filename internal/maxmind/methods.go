package maxmind

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"ip_service/pkg/helpers"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/schollz/progressbar/v3"
	"go.opentelemetry.io/otel/codes"
)

func (s *Service) dbDownloader(ctx context.Context, dbType string) error {
	ctx, span := s.TP.Start(ctx, "maxmind:dbDownloader")
	defer span.End()

	s.log.Info("downloading database file", "dbType", dbType)

	httpClient := http.Client{
		Timeout: 60 * time.Second,
	}

	if !s.DBMeta[dbType].rateLimit.Allow() {
		return fmt.Errorf("rate limit exceeded")
	}

	resp, err := httpClient.Get(fmt.Sprintf(s.DBMeta[dbType].urlDB, s.cfg.IPService.MaxMind.LicenseKey))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(fmt.Sprintf("geoip_database_%s.tar.gz", dbType))
	if err != nil {
		return err
	}
	defer out.Close()

	bar := progressbar.DefaultBytes(
		-1,
		"downloading",
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		return err
	}

	s.log.Info("UnTar", "dbType", dbType)
	if err := s.unTAR(ctx, dbType); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	s.log.Info("Cleanup tar archive for database", "dbType", dbType)
	if err := s.cleanUpTarArchive(ctx, dbType); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (s *Service) cleanUpTarArchive(ctx context.Context, dbType string) error {
	_, span := s.TP.Start(ctx, "maxmind:cleanUpTarArchive")
	defer span.End()

	return os.Remove(fmt.Sprintf("geoip_database_%s.tar.gz", dbType))
}

func (s *Service) unTAR(ctx context.Context, dbType string) error {
	_, span := s.TP.Start(ctx, "maxmind:unTAR")
	defer span.End()

	file, err := os.Open(fmt.Sprintf("geoip_database_%s.tar.gz", dbType))
	if err != nil {
		return err
	}
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		case filepath.Base(header.Name) != fmt.Sprintf("GeoLite2-%s.mmdb", dbType):
			continue
		}

		target := filepath.Join("db", filepath.Base(header.Name))

		switch header.Typeflag {
		case tar.TypeDir:
			s.log.Error(errors.New("is a dir, needs to be a file"), "unTAR")
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

// getRemoteVersion retrieve the latest remote version.
func (s *Service) getRemoteVersion(ctx context.Context, dbType string) (string, error) {
	_, span := s.TP.Start(ctx, "maxmind:getRemoteVersion")
	defer span.End()

	resp, err := http.Head(fmt.Sprintf(s.DBMeta[dbType].urlDB, s.cfg.IPService.MaxMind.LicenseKey))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	if resp.StatusCode != 200 {
		err := fmt.Errorf("errors http status code: %d", resp.StatusCode)
		if resp.StatusCode == 429 {
			err = fmt.Errorf("rate limit exceeded")
		}
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	remoteLastMod, err := time.Parse(time.RFC1123, resp.Header.Get("last-modified"))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	return remoteLastMod.String(), nil
}

func (s *Service) checkNewDBVersion(ctx context.Context, dbType string) bool {
	ctx, span := s.TP.Start(ctx, "maxmind:checkNewDBVersion")
	defer span.End()

	ok, err := s.compareVersion(ctx, dbType)
	if err != nil {
		return false
	}

	return ok
}

func (s *Service) compareVersion(ctx context.Context, dbType string) (bool, error) {
	ctx, span := s.TP.Start(ctx, "maxmind:compareVersion")
	defer span.End()

	remote, err := s.getRemoteVersion(ctx, dbType)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}

	if err := s.kvStore.SetLastChecked(ctx, dbType); err != nil {
		return false, err
	}

	local := s.kvStore.GetRemoteVersion(ctx, dbType)

	if remote == local {
		s.log.Info("No new maxmind database version found", "local_version", local, "remote_version", remote, "dbType", dbType)
		return false, nil
	}

	if err := s.kvStore.SetRemoteVersion(ctx, dbType, remote); err != nil {
		return true, err
	}

	s.log.Info("New version of MaxMind database found", "version", remote)

	return true, nil
}

func (s *Service) initial(ctx context.Context, dbType string) error {
	ctx, span := s.TP.Start(ctx, "maxmind:initial")
	defer span.End()

	// try a optimistic load of database file, if not present notify the download channel
	if err := s.loadDB(ctx, dbType); err != nil {
		if errors.Is(err, helpers.ErrMissingDBFile) {
			s.log.Info("Database file not found", "dbType", dbType)
			s.downloadChan <- dbType
			return nil
		}
		return err
	}

	return nil
}
func (s *Service) dbFilesMissing(ctx context.Context, dbType string) bool {
	_, span := s.TP.Start(ctx, "maxmind:dbFilesMissing")
	defer span.End()

	if _, err := os.Stat(s.DBMeta[dbType].filePath); errors.Is(err, os.ErrNotExist) {
		return true
	}

	return false
}

// City return a city object from ip
func (s *Service) City(ctx context.Context, ip net.IP) (*geoip2.City, error) {
	_, span := s.TP.Start(ctx, "maxmind:City")
	defer span.End()

	s.DBMeta["city"].MU.RLock()
	defer s.DBMeta["city"].MU.RUnlock()

	return s.DBCity.City(ip)
}

// ASN return information about the ASN
func (s *Service) ASN(ctx context.Context, ip net.IP) (*geoip2.ASN, error) {
	_, span := s.TP.Start(ctx, "maxmind:City")
	defer span.End()

	s.DBMeta["asn"].MU.RLock()
	defer s.DBMeta["asn"].MU.RUnlock()

	return s.DBASN.ASN(ip)
}

// ISP return information about the ISP
func (s *Service) ISP(ctx context.Context, ip net.IP) (*geoip2.ISP, error) {
	_, span := s.TP.Start(ctx, "maxmind:ISP")
	defer span.End()

	s.DBMeta["city"].MU.RLock()
	defer s.DBMeta["city"].MU.RUnlock()

	return s.DBCity.ISP(ip)
}

// AnonymousIP return information about any anonymous services
func (s *Service) AnonymousIP(ctx context.Context, ip net.IP) (*geoip2.AnonymousIP, error) {
	_, span := s.TP.Start(ctx, "maxmind:AnonymousIP")
	defer span.End()

	s.DBMeta["city"].MU.RLock()
	defer s.DBMeta["city"].MU.RUnlock()

	return s.DBCity.AnonymousIP(ip)
}

package maxmind

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/oschwald/geoip2-golang"
)

func (s *Service) fetchDatabase(ctx context.Context, url, dbType string) error {
	s.log.Info("Fetching a new version of maxmind database")
	resp, err := http.Get(fmt.Sprintf(url, s.cfg.MaxMind.LicenseKey))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(fmt.Sprintf("geoip_database_%s.tar.gz", dbType))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	remoteTime, err := s.getRemoteVersion(ctx, dbType, url)
	if err != nil {
		return err
	}
	if err := s.kvStore.Set(ctx, dbType, remoteTime); err != nil {
		return err
	}

	return nil
}

func (s *Service) cleanUpTarArchive(dbType string) error {
	return os.Remove(fmt.Sprintf("geoip_database_%s.tar.gz", dbType))
}

func (s *Service) unTAR(dbType string) error {
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
			fmt.Println("is a dir!")
			if _, err := os.Stat(header.Name); err != nil {
				if err := os.MkdirAll(target, 0775); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(file, tr); err != nil {
				return err
			}

			file.Close()
		}
	}
}

func (s *Service) getRemoteVersion(ctx context.Context, dbType, url string) (string, error) {
	resp, err := http.Head(fmt.Sprintf(url, s.cfg.MaxMind.LicenseKey))
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Errors status code: %d", resp.StatusCode)
	}

	remoteLastMod, err := time.Parse(time.RFC1123, resp.Header.Get("last-modified"))
	if err != nil {
		return "", err
	}
	return remoteLastMod.String(), nil
}

func (s *Service) isNewVersion(ctx context.Context, dbType, url string) (bool, error) {
	remoteLastMod, err := s.getRemoteVersion(ctx, dbType, url)
	if err != nil {
		return false, err
	}

	savedLastMod := s.kvStore.Get(ctx, dbType)

	// if we can't find a saved timestamp, save it and return
	if savedLastMod == "" {
		if err := s.kvStore.Set(ctx, dbType, remoteLastMod); err != nil {
			return false, err
		}
		return true, nil
	}

	if remoteLastMod == savedLastMod {
		s.log.Info("No new maxMind database version found")
		return false, nil
	}

	s.log.Info("New version of MaxMind database found", "version", remoteLastMod)

	return true, nil
}

// City return a city object from ip
func (s *Service) City(ip net.IP) (*geoip2.City, error) {
	return s.dbCity.City(ip)
}

// ASN return information about the ASN
func (s *Service) ASN(ip net.IP) (*geoip2.ASN, error) {
	return s.dbASN.ASN(ip)
}

// ISP return information about the ISP
func (s *Service) ISP(ip net.IP) (*geoip2.ISP, error) {
	return s.dbCity.ISP(ip)
}

// AnonymousIP return information about any anonymous services
func (s *Service) AnonymousIP(ip net.IP) (*geoip2.AnonymousIP, error) {
	return s.dbCity.AnonymousIP(ip)
}

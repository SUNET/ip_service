package maxmind

import (
	"context"
	"errors"
	"fmt"
	"io"
	"ip_service/pkg/model"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/oschwald/geoip2-golang"
	"go.opentelemetry.io/otel/codes"
)

func (s *Service) downloadArchive(ctx context.Context, dbType string) error {
	s.Log.Info("downloadArchive", "dbType", dbType)

	ctx, span := s.TP.Start(ctx, "maxmind:dbDownloader")
	defer span.End()

	s.DBMeta.DownloadInProgress(dbType)
	defer s.DBMeta.DownloadingDone(dbType)

	s.Log.Info("downloading database file", "dbType", dbType)

	if !s.DBMeta[dbType].rateLimit.Allow() {
		return fmt.Errorf("rate limit exceeded")
	}

	remoteURL, err := s.cfg.IPService.MaxMind.URL(dbType)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	s.Log.Info("Downloading", "url", remoteURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, remoteURL, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(s.cfg.IPService.MaxMind.Username, s.cfg.IPService.MaxMind.Password)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		s.Log.Error(err, "remote rate limit exceeded", "dbType", dbType)
		return errors.New("remote rate limit exceeded")

	}

	s.Log.Debug("downloadArchive", "path", s.cfg.IPService.MaxMind.ArchiveFilePath(dbType))

	archiveFile, err := os.OpenFile(s.cfg.IPService.MaxMind.ArchiveFilePath(dbType), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		s.Log.Error(err, "open file")
		return err
	}
	defer archiveFile.Close()

	//bar := progressbar.DefaultBytes(resp.ContentLength, fmt.Sprintf("downloading %s", dbType))
	_, err = io.Copy(archiveFile, resp.Body)
	if err != nil {
		s.Log.Error(err, "copy to broke")
		return err
	}

	s.Log.Info("download finished", "dbType", dbType)
	stat, err := archiveFile.Stat()
	if err != nil {
		return err
	}
	fmt.Println("stat size!!!!!!", dbType, stat.Size())

	if !s.cfg.IPService.MaxMind.IsArchivePresent(dbType) {
		return fmt.Errorf("archive file missing dbType: %s", dbType)
	}

	s.Log.Info("UnTar", "dbType", dbType)
	//if err := s.unTAR(ctx, dbType); err != nil {
	if err := s.unTarV3(ctx, dbType); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	s.reloadChan <- dbType

	return nil
}

func (s *Service) parseHeader(ctx context.Context, resp *http.Response) (string, error) {
	remoteLastMod, err := time.Parse(time.RFC1123, resp.Header.Get("last-modified"))
	if err != nil {
		return "", err
	}

	return remoteLastMod.String(), nil
}

// getRemoteVersion retrieve the latest remote version.
func (s *Service) getRemoteVersion(ctx context.Context, dbType string) (string, error) {
	_, span := s.TP.Start(ctx, "maxmind:getRemoteVersion")
	defer span.End()

	remoteURL, err := s.cfg.IPService.MaxMind.URL(dbType)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, remoteURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		s.Log.Error(err, "http head request failed")
		return "", err
	}

	if resp.StatusCode != 200 {
		err := fmt.Errorf("errors http status code: %d", resp.StatusCode)
		if resp.StatusCode == 429 {
			return "", errors.New("remote rate limit exceeded")
		}
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	remoteLastMod, err := s.parseHeader(ctx, resp)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	return remoteLastMod, nil
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
		s.Log.Info("No new maxmind database version found", "local_version", local, "remote_version", remote, "dbType", dbType)
		return false, nil
	}

	if err := s.kvStore.SetRemoteVersion(ctx, dbType, remote); err != nil {
		return true, err
	}

	s.Log.Info("New version of MaxMind database found", "version", remote)

	return true, nil
}

func (s *Service) initial(ctx context.Context, dbType string) error {
	ctx, span := s.TP.Start(ctx, "maxmind:initial")
	defer span.End()

	s.Log.Info("Initial", "dbType", dbType)

	if s.cfg.IPService.MaxMind.IsArchivePresent(dbType) {
		s.Log.Info("Archive file already exists", "dbType", dbType)

		if err := s.unTarV3(ctx, dbType); err != nil {
			return err
		}

		s.Log.Info("LoadDB")
		if err := s.loadDB(ctx, dbType); err != nil {
			s.Log.Error(err, "loadDB")
			return err
		}

		if !s.cfg.IPService.MaxMind.IsDBPresent(dbType) {
			s.Log.Info("Database file not found", "dbType", dbType)
			s.downloadChan <- dbType
			return nil

		}

	} else if !s.cfg.IPService.MaxMind.IsDBPresent(dbType) {
		//	} else if s.isDBFileMissing(ctx, dbType) {
		s.Log.Info("Archive file not found", "dbType", dbType)
		s.Log.Info("Downloading database file", "dbType", dbType)
		s.downloadChan <- dbType
	}

	return nil

}

// City return a city object from ip
func (s *Service) City(ctx context.Context, ip net.IP) (*geoip2.City, error) {
	_, span := s.TP.Start(ctx, "maxmind:City")
	defer span.End()

	s.DBMeta["City"].MU.RLock()
	defer s.DBMeta["City"].MU.RUnlock()

	if s.DBCity == nil {
		return nil, errors.New("city database not available")
	}

	return s.DBCity.City(ip)
}

// ASN return information about the ASN
func (s *Service) ASN(ctx context.Context, ip net.IP) (*geoip2.ASN, error) {
	s.Log.Debug("maxmind:ASN")
	_, span := s.TP.Start(ctx, "maxmind:ASN")
	defer span.End()

	s.Log.Debug("maxmind:ASN before RLock", "ip", ip.String())

	s.DBMeta[model.MaxmindDBTypeASN].MU.RLock()
	defer s.DBMeta[model.MaxmindDBTypeASN].MU.RUnlock()

	s.Log.Debug("maxmind:ASN after RUnlock")

	if s.DBASN == nil {
		return nil, errors.New("ASN database not available")
	}

	asn, err := s.DBASN.ASN(ip)
	if err != nil {
		s.Log.Error(err, "failed to get ASN")
		return nil, err
	}

	s.Log.Debug("maxmind:ASN done")

	return asn, nil
}

// ISP return information about the ISP
func (s *Service) ISP(ctx context.Context, ip net.IP) (*geoip2.ISP, error) {
	_, span := s.TP.Start(ctx, "maxmind:ISP")
	defer span.End()

	s.DBMeta["City"].MU.RLock()
	defer s.DBMeta["City"].MU.RUnlock()

	isp, err := s.DBCity.ISP(ip)
	if err != nil {
		s.Log.Error(err, "failed to get ISP")
		return nil, err
	}

	return isp, nil
}

// AnonymousIP return information about any anonymous services
func (s *Service) AnonymousIP(ctx context.Context, ip net.IP) (*geoip2.AnonymousIP, error) {
	_, span := s.TP.Start(ctx, "maxmind:AnonymousIP")
	defer span.End()

	s.DBMeta[model.MaxmindDBTypeCity].MU.RLock()
	defer s.DBMeta[model.MaxmindDBTypeCity].MU.RUnlock()

	asnIP, err := s.DBASN.AnonymousIP(ip)
	if err != nil {
		s.Log.Error(err, "failed to get AnonymousIP")
		return nil, err
	}

	return asnIP, nil
}

package rpsl

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

func (s *Client) getKey(line string) string {
	for i, r := range line {
		if i < 1 {
			if unicode.IsSpace(r) {
				return *s.currentKey
			}
			break
		}
	}

	lineParts := strings.Split(line, ":")
	if len(lineParts) > 1 {
		key := strings.Clone(strings.TrimSpace(lineParts[0]))
		s.currentKey = &key
		return key
	}
	return ""
}

func (s *Client) getValue(line, key string) string {
	sepKey := key + ":"
	_, after, found := strings.Cut(line, sepKey)
	if found {
		return strings.Clone(strings.TrimSpace(after))
	}
	return strings.Clone(strings.TrimSpace(line))
}

var interCount int

func (s *Client) Parse(ctx context.Context, databaseFilePath string) error {
	file, err := os.Open(filepath.Clean(databaseFilePath))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var isRouteObject bool

	for scanner.Scan() {
		line := scanner.Text()

		if line == "# EOF" {
			break
		}

		// Comment line, skip it
		if strings.HasPrefix(line, "#") {
			continue
		}

		interCount++

		if line == "" {
			interCount = 0

			// Insert directly into RouterClass
			if isRouteObject && s.currentRouteObject.Network != "" {
				obj := s.currentRouteObject
				routerClass, ok := s.RouterClass[obj.Network]
				if !ok {
					s.RouterClass[obj.Network] = map[string]*Object{obj.Origin: obj}
				} else {
					routerClass[obj.Origin] = obj
				}
			}

			s.currentRouteObject = &Object{}
			isRouteObject = false
			continue
		}

		key := s.getKey(line)
		if interCount == 1 {
			if key == Route || key == Route6 {
				isRouteObject = true
			} else {
				isRouteObject = false
			}
		}

		if !isRouteObject {
			continue
		}

		value := s.getValue(line, key)
		if err := s.currentRouteObject.Add(key, value); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	s.RouterClass.removeBlankRecords(ctx)

	return nil
}

// RouterClassOpinionatedMerge merges r1 into r2 if r1 does not exist in r2
func RouterClassOpinionatedMerge(ctx context.Context, r1, r2 RouterClass) (RouterClass, error) {
	_, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for r1IP, r1AsnObject := range r1 {
		r2AsnObject, ok := r2[r1IP]
		if !ok {
			// New network, add it
			r2[r1IP] = r1AsnObject
			continue
		}
		for asn, obj := range r1AsnObject {
			_, ok := r2AsnObject[asn]
			if !ok {
				// New ASN for existing network, add it
				r2AsnObject[asn] = obj
			}
		}
	}

	return r2, nil
}

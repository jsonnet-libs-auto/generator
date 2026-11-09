package swagger

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Loader interface {
	Load([]byte) (Definitions, error)
}

// Retries apply to transient upstream failures (GitHub release assets 5xx often).
var (
	fetchAttempts = 4
	fetchBackoff  = 2 * time.Second
)

func Load(loader Loader, uri string) (Definitions, error) {
	var data []byte
	var err error
	if isURL(uri) {
		data, err = fetch(uri)
	} else {
		data, err = os.ReadFile(uri)
	}
	if err != nil {
		return nil, err
	}
	return loader.Load(data)
}

func fetch(uri string) ([]byte, error) {
	var lastErr error
	for attempt := 1; attempt <= fetchAttempts; attempt++ {
		if attempt > 1 {
			time.Sleep(fetchBackoff * time.Duration(1<<(attempt-2)))
		}

		r, err := http.Get(uri)
		if err != nil {
			lastErr = err
			continue
		}
		data, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		switch {
		case r.StatusCode == http.StatusOK:
			return data, nil
		case r.StatusCode == http.StatusTooManyRequests || r.StatusCode >= 500:
			lastErr = fmt.Errorf("fetching %s: %s", uri, r.Status)
			continue
		default:
			return nil, fmt.Errorf("fetching %s: %s", uri, r.Status)
		}
	}
	return nil, fmt.Errorf("giving up after %d attempts: %w", fetchAttempts, lastErr)
}

func isURL(uri string) bool {
	endpoint, err := url.Parse(uri)
	if err != nil {
		return false
	}

	return len(endpoint.Scheme) > 0
}

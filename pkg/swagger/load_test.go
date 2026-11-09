package swagger

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withFastRetries(t *testing.T) {
	attempts, backoff := fetchAttempts, fetchBackoff
	fetchAttempts, fetchBackoff = 3, time.Millisecond
	t.Cleanup(func() { fetchAttempts, fetchBackoff = attempts, backoff })
}

func TestFetchRetriesTransientErrors(t *testing.T) {
	withFastRetries(t)
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	data, err := fetch(srv.URL)
	require.NoError(t, err)
	assert.Equal(t, "ok", string(data))
	assert.Equal(t, 3, calls)
}

func TestFetchFailsOnPersistentServerError(t *testing.T) {
	withFastRetries(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
	}))
	defer srv.Close()

	_, err := fetch(srv.URL)
	assert.ErrorContains(t, err, "504")
}

func TestFetchDoesNotRetryNotFound(t *testing.T) {
	withFastRetries(t)
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := fetch(srv.URL)
	assert.ErrorContains(t, err, "404")
	assert.Equal(t, 1, calls)
}

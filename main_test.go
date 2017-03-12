package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckForUnreachableSitesReportsServiceUnavailable(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failed", http.StatusServiceUnavailable)
	}))
	defer ts.Close()
	config := bytes.NewBufferString("{}")
	app := newApp(config)
	app.Urls = append(app.Urls, ts.URL)
	app.checkForUnreachableSites()
	select {
	case s := <-app.unreachables:
		{
			if s.status != "503 Service Unavailable" {

				t.Errorf("site.isDown() returned error status: %v, expected %v", s.status, "503 Service Unavailable")
			}
		}
	}
}

func TestCheckForUnreachableSitesDoesntReportErrorOnStatusOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, "Status OK")

	}))
	defer ts.Close()
	config := bytes.NewBufferString("{}")
	app := newApp(config)
	app.Urls = append(app.Urls, ts.URL)
	app.checkForUnreachableSites()
	if len(app.unreachables) > 0 {
		t.Error("site should not have been added to the unreachables queue")
	}
}

func TestCheckForUnreachableSitesReportsErrorOnTimeout(t *testing.T) {
	clientTimeout := 100 * time.Millisecond
	serverTimeout := 200 * time.Millisecond
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(serverTimeout)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, "Status OK")

	}))
	defer ts.Close()
	config := bytes.NewBufferString("{}")
	app := newApp(config)
	app.client.Timeout = clientTimeout
	app.Urls = append(app.Urls, ts.URL)
	app.checkForUnreachableSites()
	if len(app.unreachables) == 0 {
		t.Error("client should have timed out on request")
	}
}

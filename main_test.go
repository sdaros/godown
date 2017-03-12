package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIsDownReturnsCorrectStatusServiceUnavailable(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failed", http.StatusServiceUnavailable)
	}))
	defer ts.Close()
	//t.Error("%v should not return error status")
	app := new(App)
	// Timeout after 10 seconds
	tr := &http.Transport{
		IdleConnTimeout: 10 * time.Second,
	}
	app.client = &http.Client{Transport: tr}
	app.downSites = make(chan site)
	site := site{url: ts.URL}
	app.wg.Add(1)
	site.isDown(app)
	select {
	case s := <-app.downSites:
		{
			if s.status != "503 Service Unavailable" {

				t.Errorf("site.isDown() returned error status: %v, expected %v", s.status, "503 Service Unavailable")
			}
		}
	}
}

func TestIsDownReturnsCorrectStatusOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, "Status OK")

	}))
	defer ts.Close()
	//t.Error("%v should not return error status")
	app := new(App)
	// Timeout after 10 seconds
	tr := &http.Transport{
		IdleConnTimeout: 10 * time.Second,
	}
	app.client = &http.Client{Transport: tr}
	app.downSites = make(chan site)
	site := site{url: ts.URL}
	app.wg.Add(1)
	site.isDown(app)
	if len(app.downSites) > 0 {
		t.Error("site should not have been added to the down queue")
	}
}

func TestSiteIsAddedToDownQueueOnIdleConnTimeout(t *testing.T) {

}

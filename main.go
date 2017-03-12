package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type App struct {
	Urls      []string
	client    *http.Client
	downSites chan site
	wg        sync.WaitGroup
}

type site struct {
	url    string
	status string
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Print("usage: godown <config.json>\n")
		os.Exit(-1)
	}
	configFile := os.Args[1]
	config, err := os.Open(configFile)
	die("error: unable to find configuration file: %v", err)
	app := newApp(config)
	app.downSites = make(chan site, len(app.Urls))
	for _, url := range app.Urls {
		app.wg.Add(1)
		site := site{url: url}
		site.isDown(app)
	}
	// Wait until all sites are checked
	app.wg.Wait()
	close(app.downSites)
	if len(app.downSites) > 0 {
		for v := range app.downSites {
			log.Printf("URL %v \n returned status: %v", v.url, v.status)
		}
		os.Exit(-1)
	}
}

func (a *App) checkForUnreachableSites() {

}

func (s site) isDown(app *App) {
	go func(url string) {
		defer app.wg.Done()
		resp, err := app.client.Head(url)
		if err != nil {
			s.status = err.Error()
			app.downSites <- s
			return
		}
		if resp.StatusCode != http.StatusOK {
			s.status = resp.Status
			app.downSites <- s
		}
	}(s.url)
}

func newApp(config io.ReadWriter) *App {
	app := new(App)
	decoder := json.NewDecoder(config)
	err := decoder.Decode(app)
	die("error: unable to parse configuration file: %v", err)
	// Timeout after 10 seconds
	tr := &http.Transport{
		IdleConnTimeout: 10 * time.Second,
	}
	app.client = &http.Client{Transport: tr}
	return app
}

func die(format string, err error) {
	if err != nil {
		log.Fatalf(format, err)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type App struct {
	Urls             []string
	client           *http.Client
	unreachableSites []site
	siteIsDown       chan site
	wg               sync.WaitGroup
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
	app := newApp()
	app.siteIsDown = make(chan site)
	for _, url := range app.Urls {
		app.wg.Add(1)
		site := site{url: url}
		site.isDown(app)
	}
	// wait until all urls have been checked before proceeding
	app.wg.Wait()
	close(app.siteIsDown)
	if len(app.unreachableSites) != 0 {
		for _, v := range app.unreachableSites {
			log.Printf("URL %v \n returned status: %v", v.url, v.status)
		}
		os.Exit(-1)
	}
}

func (s site) isDown(app *App) {
	app.waitThenAddDownSite()
	s.checkStatus(app)
}

func (s site) checkStatus(app *App) {
	go func(url string) {
		defer app.wg.Done()
		resp, err := app.client.Head(url)
		if err != nil {
			s.status = err.Error()
			app.siteIsDown <- s
			return
		}
		if resp.StatusCode != http.StatusOK {
			s.status = resp.Status
			app.siteIsDown <- s
		}
	}(s.url)
}

func (app *App) waitThenAddDownSite() {
	go func() {
		for {
			site := <-app.siteIsDown
			// TODO: Still seems to be a data race here
			app.unreachableSites = append(app.unreachableSites, site)
		}
	}()
}

func newApp() *App {
	app := new(App)
	configFile := os.Args[1]
	app.loadConfig(configFile)
	// Timeout after 10 seconds
	tr := &http.Transport{
		IdleConnTimeout: 10 * time.Second,
	}
	app.client = &http.Client{Transport: tr}
	return app
}

func (app *App) loadConfig(path string) {
	configFile, err := os.Open(path)
	die("error: unable to find configuration file: %v", err)
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(app)
	die("error: unable to parse configuration file: %v", err)
}

func die(format string, err error) {
	if err != nil {
		log.Fatalf(format, err)
	}
}

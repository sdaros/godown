package main

import (
	"encoding/json"
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
}

type site struct {
	url    string
	status string
}

var siteDown chan site

func main() {
	app := new(App)
	siteDown = make(chan site)
	configFile := os.Args[1]
	app.loadConfig(configFile)

	// Timeout after 10 seconds
	tr := &http.Transport{
		IdleConnTimeout: 10 * time.Second,
	}
	app.client = &http.Client{Transport: tr}
	app.waitForThenAddUnrechableSite()
	app.checkForUnrechableSite()
	if len(app.unreachableSites) != 0 {
		for _, v := range app.unreachableSites {
			log.Printf("URL %v \n returned status: %v", v.url, v.status)
		}
		os.Exit(-1)
	}
}

func (app *App) waitForThenAddUnrechableSite() {
	go func() {
		for {
			res := <-siteDown
			app.unreachableSites = append(app.unreachableSites, res)
		}
	}()
}

func (app *App) checkForUnrechableSite() {
	var wg sync.WaitGroup
	for _, url := range app.Urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			resp, err := app.client.Head(url)
			if err != nil {
				siteDown <- site{url: url, status: err.Error()}
				return
			}
			if resp.StatusCode != http.StatusOK {
				siteDown <- site{url: url, status: resp.Status}
			}
		}(url)
	}
	wg.Wait()
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

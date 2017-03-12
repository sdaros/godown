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
	Urls         []string
	client       *http.Client
	unreachables chan site
	wg           sync.WaitGroup
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
	app.checkForUnreachableSites()
	if len(app.unreachables) > 0 {
		for v := range app.unreachables {
			log.Printf("URL %v \n returned status: %v", v.url, v.status)
		}
		os.Exit(-1)
	}

}

func (app *App) checkForUnreachableSites() {
	app.unreachables = make(chan site, len(app.Urls))
	for _, url := range app.Urls {
		app.wg.Add(1)
		site := site{url: url}
		site.isDown(app)
	}
	app.wg.Wait()
	close(app.unreachables)
}

func (s site) isDown(app *App) {
	go func(url string) {
		defer app.wg.Done()
		resp, err := app.client.Head(url)
		if err != nil {
			s.status = err.Error()
			app.unreachables <- s
			return
		}
		if resp.StatusCode != http.StatusOK {
			s.status = resp.Status
			app.unreachables <- s
		}
	}(s.url)
}

func newApp(config io.ReadWriter) *App {
	app := new(App)
	decoder := json.NewDecoder(config)
	err := decoder.Decode(app)
	die("error: unable to parse configuration file: %v", err)
	// Timeout and assume error if request takes too long
	app.client = &http.Client{Timeout: 10 * time.Second}
	return app
}

func die(format string, err error) {
	if err != nil {
		log.Fatalf(format, err)
	}
}

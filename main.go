package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"sync"
	"time"
)

type App struct {
	Email        Email    `json:"email"`
	Urls         []string `json:"urls"`
	client       *http.Client
	unreachables chan site
	wg           sync.WaitGroup
}

type site struct {
	url    string
	status string
}

type Email struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Server    string `json:"server"`
	Port      string `json:"port"`
	Recipient string `json:"recipient"`
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Print("usage: godown <config.json>\n")
		os.Exit(-1)
	}
	configFile := os.Args[1]
	config, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("error: unable to find configuration file: %v", err)
	}
	app := newApp(config)
	app.checkForUnreachableSites()
	if len(app.unreachables) > 0 {
		app.sendMail()
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

func (app *App) sendMail() {
	var unreachables string
	for v := range app.unreachables {
		unreachables += fmt.Sprintf("---\nURL: %v\nError: %v\n", v.url, v.status)
	}
	auth := smtp.PlainAuth("", app.Email.Username, app.Email.Password, app.Email.Server)
	to := []string{app.Email.Recipient}
	msg := []byte("To: " + app.Email.Recipient + "\r\n" +
		"Subject: godown discovered unreachable sites!\r\n" +
		"\r\n" +
		unreachables + "\r\n")
	err := smtp.SendMail(app.Email.Server+":"+app.Email.Port, auth, app.Email.Username, to, msg)
	if err != nil {
		log.Fatalf("error: unable to send email to %v: %v", to, err)
	}
}

func newApp(config io.ReadWriter) *App {
	app := new(App)
	if err := json.NewDecoder(config).Decode(app); err != nil {
		log.Fatalf("error: unable to parse configuration file: %v", err)
	}
	// Timeout and assume error if request takes too long
	app.client = &http.Client{Timeout: 10 * time.Second}
	return app
}

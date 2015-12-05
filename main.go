package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sendgrid/sendgrid-go"
)

func checkSite(config *configFile, site Site) {
	response, err := http.Get(site.Url)
	if err != nil {
		log.Println(err)
		body := fmt.Sprintf("Monitoring alert for %s could not connect: %s", site.Url, err)
		sendEmail(config, site, body)
		return
	} else {
		defer response.Body.Close()
		start := time.Now()
		contents, err := ioutil.ReadAll(response.Body)

		// connectivity
		if err != nil {
			body := fmt.Sprintf("Monitoring alert for %s could not read: %s", site.Url, err)
			sendEmail(config, site, body)
			return
		}
		// response time
		if site.Threshold > 0 {
			secs := time.Since(start).Seconds()
			if secs*1000.00 > site.Threshold {
				body := fmt.Sprintf("Monitoring alert for %s time spent: %f threshold %f", site.Url, secs*1000, site.Threshold)
				sendEmail(config, site, body)

			}
		}
		// keyword
		if site.Keyword != "" {
			if strings.Contains(string(contents), site.Keyword) {
				return
			}
		}
	}
	body := fmt.Sprintf("Monitoring alert for %s keyword %s not found", site.Url, site.Keyword)
	sendEmail(config, site, body)
}

func sendEmail(config *configFile, site Site, body string) {
	sg := sendgrid.NewSendGridClient(config.Sendgrid.User, config.Sendgrid.Password)
	message := sendgrid.NewMail()
	for _, rcpt := range config.Rcpts {
		message.AddTo(rcpt.Email)
		message.AddToName(rcpt.Name)
	}
	message.SetSubject(config.Sendgrid.Subject)
	message.SetText(body)
	message.SetFrom(config.Sendgrid.From)
	if r := sg.Send(message); r == nil {
		fmt.Println("Email sent!")
	} else {
		fmt.Println(r)
	}

}

func main() {
	configFile := flag.String("c", "server.conf", "")
	flag.Usage = func() {
		fmt.Println("Usage: httpmonitor [-c server.conf]")
		os.Exit(1)
	}
	flag.Parse()

	var err error
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	for _, site := range config.Sites {
		checkSite(config, site)
	}
}

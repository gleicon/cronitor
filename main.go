package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

func checkSite(config *configFile, site Site) {
	response, err := http.Get(site.Url)
	if err != nil {
		log.Println(err)
		body := fmt.Sprintf("Monitoring alert for %s could not connect: %s", site.Url, err)
		sendEmail(config, body)
		sendSlack(config, body)
		sendKeenMetrics(config, site.Url, "connect", 0.0, []string{"down"})
		return
	} else {
		defer response.Body.Close()
		start := time.Now()
		contents, err := ioutil.ReadAll(response.Body)

		// connectivity
		if err != nil {
			body := fmt.Sprintf("Monitoring alert for %s could not read: %s", site.Url, err)
			sendEmail(config, body)
			sendSlack(config, body)
			sendKeenMetrics(config, site.Url, "read", 0.0, []string{"down", "read error"})
			return
		}
		// response time
		secs := time.Since(start).Seconds()
		lt := secs * 1000.00
		if site.Threshold > 0 {
			if lt > site.Threshold {
				body := fmt.Sprintf("Monitoring alert for %s time spent: %f threshold %f", site.Url, lt, site.Threshold)
				sendEmail(config, body)
				sendSlack(config, body)
				sendKeenMetrics(config, site.Url, "latency", lt, []string{"slow"})

			}
		}
		// keyword
		if site.Keyword != "" {
			if strings.Contains(string(contents), site.Keyword) {
				sendKeenMetrics(config, site.Url, "check", lt, []string{"up"})
				return
			}
		}
	}
	body := fmt.Sprintf("Monitoring alert for %s keyword %s not found", site.Url, site.Keyword)
	sendEmail(config, body)
	sendSlack(config, body)
	sendKeenMetrics(config, site.Url, "keyword", 0.0, []string{"keyword not found"})
}

func sendEmail(config *configFile, body string) {
	d := gomail.NewPlainDialer(config.SMTP.Hostname, config.SMTP.Port, config.SMTP.User, config.SMTP.Password)
	if config.SMTP.SkipTLSCheck {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	s, err := d.Dial()
	if err != nil {
		log.Printf("Could not connect to smtp server %q port %d: %v", config.SMTP.Hostname, config.SMTP.Port, err)
		return
	}
	if config.Endpoint != "" {
		body = body + "\n\nEndpoint: " + config.Endpoint
	}

	m := gomail.NewMessage()
	for _, rcpt := range config.Rcpts {
		m.SetHeader("From", config.SMTP.From)
		m.SetAddressHeader("To", rcpt.Email, rcpt.Name)
		m.SetHeader("Subject", config.SMTP.Subject)
		m.SetBody("text/plain", body)

		if err := gomail.Send(s, m); err != nil {
			log.Printf("Could not send email to %q: %v", rcpt.Email, err)
		}
		m.Reset()
	}
}

func sendSlack(config *configFile, body string) {
	if config.SLACK.URL == "" {
		return
	}
	payload := fmt.Sprintf("{\"channel\": \"%s\", \"username\": \"%s\", \"text\": \"%s\", \"icon_emoji\": \"%s\"}", config.SLACK.Channel, config.SLACK.Username, body, config.SLACK.IconEmoji)

	log.Println(payload)

	req, err := http.NewRequest("POST", config.SLACK.URL, bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Could not connect to slack %q: %v", config.SLACK.Channel, err)
	}
	defer resp.Body.Close()
	bb, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(bb))
	if string(bb) != "ok" {
		log.Printf("Could not send slack message to %q: %v", config.SLACK.Channel, err)
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

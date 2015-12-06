package main

import (
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
		return
	} else {
		defer response.Body.Close()
		start := time.Now()
		contents, err := ioutil.ReadAll(response.Body)

		// connectivity
		if err != nil {
			body := fmt.Sprintf("Monitoring alert for %s could not read: %s", site.Url, err)
			sendEmail(config, body)
			return
		}
		// response time
		if site.Threshold > 0 {
			secs := time.Since(start).Seconds()
			if secs*1000.00 > site.Threshold {
				body := fmt.Sprintf("Monitoring alert for %s time spent: %f threshold %f", site.Url, secs*1000, site.Threshold)
				sendEmail(config, body)

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
	sendEmail(config, body)
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

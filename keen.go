package main

import "gopkg.in/inconshreveable/go-keen.v0"

type KeenEvent struct {
	URL      string
	Loadtime float64
	Event    string
	Tags     []string
}

func sendKeenMetrics(config *configFile, url string, eventType string,
	loadtime float64, tags []string) {
	keenClient := &keen.Client{ApiKey: config.KEEN.APIKey,
		ProjectToken: config.KEEN.ProjectToken}

	keenClient.AddEvent(config.KEEN.CollectionName, &KeenEvent{
		URL:      url,
		Event:    eventType,
		Tags:     tags,
		Loadtime: loadtime,
	})
}

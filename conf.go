// Copyright 2015 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import "github.com/BurntSushi/toml"

type Site struct {
	Url       string  `toml:"url"`
	Keyword   string  `toml:"keyword"`
	Threshold float64 `toml:"threshold"`
}

type Rcpt struct {
	Email string `toml:"email"`
	Name  string `toml:"name"`
}

type configFile struct {
	Debug    bool   `toml:"debug"`
	Endpoint string `toml:"endpoint"`

	SMTP struct {
		Hostname     string `toml:"hostname"`
		Port         int    `toml:"port"`
		User         string `toml:"user"`
		Password     string `toml:"password"`
		Subject      string `toml:"subject"`
		From         string `toml:"from"`
		SkipTLSCheck bool   `toml:"skip_tls_check"`
	} `toml:"smtp"`

	SLACK struct {
		URL       string `toml:"url"`
		Channel   string `toml:"channel"`
		Username  string `toml:"username"`
		IconEmoji string `toml:"icon_emoji"`
	} `toml:"slack"`

	Rcpts []Rcpt `toml:"rcpt"`

	Sites []Site `toml:"site"`
}

// LoadConfig reads and parses the configuration file.
func loadConfig(filename string) (*configFile, error) {
	c := &configFile{}
	if _, err := toml.DecodeFile(filename, c); err != nil {
		return nil, err
	}

	// Make files' path relative to the config file's directory.
	return c, nil
}

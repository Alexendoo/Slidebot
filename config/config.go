package config

import (
	"encoding/json"
	"io/ioutil"
)

// Config is the top level structure of the config.json file.
type Config struct {
	Tokens *APITokens        `json:"tokens"`
	Repos  map[string]string `json:"Repos"`
}

// APITokens used to authenticate the bot with various services.
type APITokens struct {
	LastFM  string `json:"last.fm"`
	Discord string `json:"discord"`
}

// Tokens is the set of APITokens read from the config file.
var Tokens *APITokens

// Repos as read from the config file, a map of GitHub repositories full names
// (e.g. ccrama/Slide) to Discord channel IDs.
var Repos map[string]string

// Load the config file located at config.json into the variables exposed by
// this package.
func Load() error {
	body, err := ioutil.ReadFile("config.json")

	if err != nil {
		return err
	}

	var config Config
	err = json.Unmarshal(body, &config)

	Tokens = config.Tokens
	Repos = config.Repos

	return err
}

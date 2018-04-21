package config

import (
	"encoding/json"
	"io/ioutil"
)

type APITokens struct {
	LastFM  string `json:"last.fm"`
	Discord string `json:"discord"`
}

var Tokens APITokens

func Open() error {
	body, err := ioutil.ReadFile("tokens.json")

	if err != nil {
		return err
	}

	return json.Unmarshal(body, &Tokens)
}

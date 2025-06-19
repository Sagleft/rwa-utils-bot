package main

import (
	"fmt"

	swissknife "github.com/Sagleft/swiss-knife"
)

const configFilePath = "config.json"

type Config struct {
	CollectionAddress string  `json:"collectionAddress"`
	NftYield          float64 `json:"nftYield"`
	AdminID           int64   `json:"adminID"`
	BotToken          string  `json:"botToken"`
}

func GetConfig() (Config, error) {
	var cfg Config
	if err := swissknife.ParseStructFromJSONFile(configFilePath, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse: %w", err)
	}
	return cfg, nil
}

package config

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
)

type BaseConfig struct {
	LogLevel    string `json:"log_level"`
	ServiceName string `json:"service_name"`
	ServerID    string `json:"server_id"`
}

type NeoSettings struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Timeout  int    `json:"timeout"`
	PoolSize int    `json:"pool_size"`
}

type TelegramSettings struct {
	APIToken string `json:"api_token"`
	BotName  string `json:"bot_name"`
}

type TelegramPolling struct {
	PollTimeout int `json:"poll_timeout"`
	ChannelSize int `json:"channel_size"`
	RetryDelay  int `json:"retry_delay"`
}

type TelegramSender struct {
	HttpTimeout int `json:"http_timeout"`
	Retries     int `json:"retries"`
}

func FromJSONFile(path string, config interface{}) error {
	if path == "" {
		return errors.New("empty config path")
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "cannot read config file")
	}
	err = json.Unmarshal(data, config)
	if err != nil {
		return errors.Wrap(err, "cannot parse json file")
	}
	return nil
}
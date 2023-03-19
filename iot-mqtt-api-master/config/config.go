package config

import (
	"github.com/BurntSushi/toml"
)

// Configuration struct have all fields from configuration JSON file
type Configuration struct {
	Database Database
	MQTT     MQTT
	Server   Server
	Log      Log
	Router   Router
	OwApiURL OwApiURL
}

type OwApiURL struct {
	URL string
}

// Database defines DB configuration
type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// MQTT defines MQTT configuration
type MQTT struct {
	Host          string
	Port          string
	User          string
	Password      string
	ClientID      string
	Subscriptions []string
}

// Server defines http server configuration (address and port)
type Server struct {
	Address string
}

// Log defines logging configuration (log filename)
type Log struct {
	Filename string
}

// Router defines router (Gin) configuration
type Router struct {
	Release bool
}

// Load loads configuration from path
func Load(path string) (*Configuration, error) {
	c := new(Configuration)

	if _, err := toml.DecodeFile(path, c); err != nil {
		return c, err
	}

	return c, nil
}

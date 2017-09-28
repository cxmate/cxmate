package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

//Config represents the configuration of the entire application. All configuration
//should be centralized here.
type Config struct {
	//General options that configure this application
	General GeneralConfig `json:"general"`
	//Service configures how the proxy will connect to the external Service.
	Service ServiceConfig `json:"service"`
}

//GeneralConfig options that configure this application
type GeneralConfig struct {
	//Location is a URL:PORT that this proxy will listen on.
	Location string `json:"location"`
	//Domain is a DNS domain name, setting it will also turn on TLS.
	Domain string `json:"domain"`
	//Debug will turn on verbose logging and print out configuration parameters.
	Logger LogConfig `json:"logger"`
	//ReadTimeout configures the HTTP timeout for reading a request body
	ReadTimeout time.Duration
	//WriteTimeout configures the HTTP timeout for writing a response body
	WriteTimeout time.Duration
	//IdleTimeout configures the HTTP timeout for TCP keep-alive
	IdleTimeout time.Duration
}

func (c GeneralConfig) validate() error {
	if c.Location == "" {
		return errors.New("general config missing required location field")
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 5 * time.Second
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 10 * time.Second
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 120 * time.Second
	}
	return nil
}

//ServiceConfig configures how the proxy will connect to the external Service.
type ServiceConfig struct {
	//Location is a URL:PORT that the Proxy expects the service to be listening on.
	Location string `json:"location"`
	//Title gives a short descriptive name to the algorithm.
	Title string `json:"title"`
	//Version gives a version to the algorithm.
	Version string `json:"version"`
	//Author is the name(s) of the algorithm author.
	Author string `json:"author"`
	//Description gives a small description of the service, what it does, and any caveats to using it.
	Description string `json:"description"`
	//Keywords is a list of words and phrases that describes the domain of the algorithm.
	Keywords []string `json:"keywords"`
	//License should be set to the name under which this algorithm is licensed.
	License string `json:"license"`
	//Language should be set to the name of the programming langugage the algorithm is written in.
	Language string `json:"language"`
	//Parameters is a list of key/value pair objects that should augment the way an algorithm behaves.
	Parameters ParameterConfig `json:"parameters"`
	//Inputs is used to describe multiple networks as input to the algorithm.
	Input ParserConfig `json:"input"`
	//SingletonInput will cause input configs of size 1 to be treated as scalars
	SingletonInput bool `json:"singletonInput"`
	//Outputs is used to describe multiple networks as output to the algorithm.
	Output GeneratorConfig `json:"output"`
	//SingletonOutput will cause output configs of size 1 to be treated as scalars
	SingletonOutput bool `json:"singletonOutput"`
}

func (c ServiceConfig) validate() error {
	if c.Location == "" {
		return errors.New("service config missing required location field")
	}
	if c.Title == "" {
		return errors.New("service config missing required name field")
	}
	if c.Version == "" {
		return errors.New("service config missing required version field")
	}
	return nil
}

//Print prints the config to standard output in JSON form.
func (c *Config) Print() error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	if err := enc.Encode(c); err != nil {
		return errors.New(fmt.Sprint("error reporting configuration:", err))
	}
	return nil
}

const confLocation = "cxmate.json"

//loadConfig loads a cxmate.json config file from the current directory
func loadConfig() (*Config, error) {
	logDebug("Loading cxmate.json configuration file")
	file, err := os.Open(confLocation)
	if err != nil {
		return nil, err
	}
	return loadFrom(file)
}

//loadFrom loads a cxMate config struct from a reader
func loadFrom(r io.Reader) (*Config, error) {
	c := &Config{}
	dec := json.NewDecoder(r)
	if err := dec.Decode(c); err != nil {
		return nil, err
	}
	return c, nil
}

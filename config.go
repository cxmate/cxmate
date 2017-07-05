package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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
}

//ServiceConfig configures how the proxy will connect to the external Service.
type ServiceConfig struct {
	//Location is a URL:PORT that the Proxy expects the service to be listening on.
	Location string `json:"location"`
	//Name gives a short descriptive name to the algorithm.
	Name string `json:"name"`
	//Author is the name(s) of the algorithm author.
	Author string `json:"author"`
	//Summary gives a small description of the service, what it does, and any caveats to using it.
	Summary string `json:"summary"`
	//Keywords is a list of words and phrases that describes the domain of the algorithm.
	Keywords []string `json:"keywords"`
	//License should be set to the name under which this algorithm is licensed.
	License string `json:"license"`
	//Language should be set to the name of the programming langugage the algorithm is written in.
	Language string `json:"language"`
	//Parameters is a list of key/value pair objects that should augment the way an algorithm behaves.
	Parameters []ParameterConfig `json:"parameters"`
	//Inputs is used to describe multiple networks as input to the algorithm.
	Input ParserConfig `json:"input"`
	//Outputs is used to describe multiple networks as output to the algorithm.
	Output GeneratorConfig `json:"output"`
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

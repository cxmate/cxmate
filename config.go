package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
	Input []DecoderConfig `json:"input"`
	//Outputs is used to describe multiple networks as output to the algorithm.
	Output []EncoderConfig `json:"output"`
}

//EncoderConfig is a MOCK
type EncoderConfig struct{}

const configfile = "cxmate.json"

//loadConfig loads a cxMate config file from the current directory
func loadConfig() (*Config, error) {
	file, err := ioutil.ReadFile(configfile)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	if err := json.Unmarshal(file, config); err != nil {
		return nil, err
	}
	return config, nil

}

//PrintConfig prints the given configuration to standard output. This is useful for debugging.
func (c *Config) PrintConfig() error {
	fmt.Println("Currently loaded configuration:")
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	if err := enc.Encode(c); err != nil {
		return errors.New("Error reporting configuration... " + err.Error())
	}
	return nil
}

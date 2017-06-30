package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var completeConfig = `
{
  "general": {
    "location": "0.0.0.0:80",
    "domain": "cxmate.ericsage.io",
    "debug": false
  },
  "service": {
    "location": "localhost:8080"
  },
  "algorithm": {
    "name": "Echo",
    "author": "Eric Sage",
    "summary": "A test service that echos its input to its output.",
	"keywords": ["bioinformatics", "networks", "algorithms"],
    "license": "MIT",
    "language": "Golang",
    "parameters": [
      {
        "key": "test",
        "default": "value",
        "description": "A test parameter."
      }
    ],
    "input": [
      {
        "label": "Input",
        "description": "An input network to be echoed",
        "aspects": ["node", "edge", "nodeAttribute", "edgeAttribute", "networkAttribute"],
        "download": {
          "endpoint": "http://ndexbio.org/v2/network",
          "method": "GET",
          "authorization": "BASIC LKJ544KL534J5KL"
        }
      }
    ],
    "output": [
      {
        "label": "Output",
        "description": "An output network which is the same network as the input.",
        "aspects": ["node", "edge", "nodeAttribute", "edgeAttribute", "networkAttribute"],
        "upload": {
          "endpoint": "http://ndexbio.org/v2/network",
          "method": "POST",
          "authorization": "BASIC LKJ544KL534J5KL"
        }
      }
    ]
  }
}
`

func ensureConfigExists(t *testing.T) {
	if _, err := os.Stat("./config.json"); os.IsNotExist(err) {
		fmt.Println("Could not find config")
		if _, err := os.Create("./config.json"); err != nil {
			t.Errorf("Could not create test config file, error: %#v", err)
		}
	}
}

func backupConfig(t *testing.T) []byte {
	backup, err := ioutil.ReadFile("./config.json")
	if err != nil {
		t.Errorf("Could not read config file into backup, error: %#v", err)
	}
	return backup
}

func restoreConfig(t *testing.T, backup []byte) {
	if err := ioutil.WriteFile("./config.json", backup, 0644); err != nil {
		t.Errorf("Could not restore config from backup, error: %#v", err)
	}
}

func writeConfig(t *testing.T, config string) {
	if err := ioutil.WriteFile("./config.json", []byte(config), 0644); err != nil {
		t.Errorf("Could not write sample config, %#v", err)
	}
}

func TestReadConfig(t *testing.T) {
	ensureConfigExists(t)
	backup := backupConfig(t)
	defer restoreConfig(t, backup)

	t.Run("SetCompleteConfig", func(t *testing.T) {
		writeConfig(t, completeConfig)
		c, err := LoadConfig("./config.json")
		if err != nil {
			t.Errorf("Could not load config, error: %#v", err)
		}
		b, err := json.MarshalIndent(c, "", " ")
		if err != nil {
			t.Errorf("Could not marshal config to byte array: %#v", err)
		}
		if compress(string(b)) != compress(completeConfig) {
			t.Errorf("Fields were not set in the complete config")
		}
	})

}

func compress(s string) string {
	return strings.NewReplacer(" ", "", "\n", "", "\t", "").Replace(s)
}

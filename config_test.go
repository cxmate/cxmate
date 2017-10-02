package main

import (
	"strings"
	"testing"
)

var completeConfig = `
{
	"general": {
		"location": "0.0.0.0:80",
		"domain": "test.domain.com",
		"logger": {
			"debug": true,
			"file": "cxmate.log",
			"format": "json"
		},
		"ReadTimeout": 1,
		"WriteTimeout": 2,
		"IdleTimeout": 3
	},
	"service": {
		"location": "localhost:8080",
		"title": "test",
		"author": "John Doe",
		"description": "my service",
		"keywords": ["networks"],
		"license": "MIT",
		"language": "Python",
		"parameters": [
			{
				"name": "test_param",
				"default": "1",
				"description": "test param",
				"type": "integer",
				"format": "test"
			}
		],
		"input": [
			{
				"label": "Input",
				"description": "A test input network",
				"aspects": ["nodes"]
			}
		],
		"output": [
			{
				"label": "Output",
				"description": "A test output network",
				"aspects": ["nodes"]
			}
		]
	}
}
`

// TestLoadCompleteConfig verifies that all of the valid config fields have been set correctly.
func TestLoadConfig(t *testing.T) {
	r := strings.NewReader(completeConfig)
	config, err := loadFrom(r)
	if err != nil {
		t.Fatal(err)
	}
	if config.General.Location != "0.0.0.0:80" {
		t.Error("config.General.Location not set")
	}
	if config.General.Domain != "test.domain.com" {
		t.Error("config.General.Domain not set")
	}
	if config.General.Logger.Debug != true {
		t.Error("config.General.Logger.Debug not set")
	}
	if config.General.Logger.File != "cxmate.log" {
		t.Error("config.General.Logger.File not set")
	}
	if config.General.Logger.Format != "json" {
		t.Error("config.General.Logger.Format not set")
	}
	if config.General.ReadTimeout != 1 {
		t.Error("config.General.ReadTimeout not set")
	}
	if config.General.WriteTimeout != 2 {
		t.Error("config.General.WriteTimeout not set")
	}
	if config.General.IdleTimeout != 3 {
		t.Error("config.General.IdleTimeout not set")
	}
	if config.Service.Location != "localhost:8080" {
		t.Error("config.Service.Location not set")
	}
	if config.Service.Title != "test" {
		t.Error("config.Service.Title not set")
	}
	if config.Service.Author != "John Doe" {
		t.Error("config.Service.Author not set")
	}
	if config.Service.Description != "my service" {
		t.Error("config.Service.Description not set")
	}
	if len(config.Service.Keywords) != 1 {
		t.Fatal("config.Service.Keywords does not contain the correct amount of parameters")
	}
	if config.Service.Keywords[0] != "networks" {
		t.Error("config.Service.Keywords not set")
	}
	if config.Service.License != "MIT" {
		t.Error("config.Service.License not set")
	}
	if config.Service.Language != "Python" {
		t.Error("config.Service.Language not set")
	}
	if len(config.Service.Parameters) != 1 {
		t.Fatal("config.Service.Parameters does not contain the correct amount of parameters")
	}
	if config.Service.Parameters[0].Name != "test_param" {
		t.Error("config.Service.Parameters[0].Name not set")
	}
	if config.Service.Parameters[0].Default != "1" {
		t.Error("config.Service.Parameters[0].Default not set")
	}
	if config.Service.Parameters[0].Description != "test param" {
		t.Error("config.Service.Parameters[0].Description not set")
	}
	if config.Service.Parameters[0].Type != "integer" {
		t.Error("config.Service.Parameters[0].Type not set")
	}
	if config.Service.Parameters[0].Format != "test" {
		t.Error("config.Service.Parameters[0].Format not set")
	}

}

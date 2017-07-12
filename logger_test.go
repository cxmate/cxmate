package main

import "testing"

func TestNewLogger(t *testing.T) {
	c := LogConfig{
		Debug:  false,
		File:   "",
		Format: "",
	}
	_, err := c.NewLogger("my-service", "v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
}

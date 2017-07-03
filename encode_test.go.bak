package main

import (
	"os"
	"testing"
)

func NewTestEncoder(t *testing.T, aspects ...string) *Encoder {
	c := EncoderConfig{
		Label:       "my_network",
		Description: "my test network",
		Aspects:     aspects,
	}
	enc, err := NewEncoder(os.Stdout, c)
	if err != nil {
		t.Error(err)
	}
	return enc
}

func TestEncodeNumberVerfication(t *testing.T) {
	enc := NewTestEncoder(t, "nodes")
	enc.EncodeNumberVerfication()
}

func TestEncoderPreMetadata(t *testing.T) {
	enc := NewTestEncoder(t, "nodes", "edges")
	enc.EncodeMetadata()
}

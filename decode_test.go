package main

import (
	"io"
	"strings"
	"testing"
)

var sampleCXNetwork = `
[
  {
    "numberVerification": [{"longNumber":281474976710655}]
  },
  {
    "metaData": [
      {"name": "nodes", "elementCount": 2, "properties":[], "version":"1.0"},
      {"name": "edges", "elementCount": 1, "properties":[], "version":"1.0"},
      {"name": "nodeAttributes","properties":[], "version":"1.0"},
      {"name": "edgeAttributes", "properties":[], "version":"1.0"},
      {"name": "networkAttributes", "properties":[], "version":"1.0"}
    ]
  },
  {
    "nodes": [
      { "@id": 1, "n": "test_node_1", "r": "test_data" },
      { "@id": 2, "n": "test_node_2", "r": "test_data" }
    ]
  },
  {
    "edges": [
      { "@id": 1, "s": 1, "t": 2, "i": "test_connection" }
    ]
  },
  {
    "nodeAttributes": [
      { "po": 1, "n": "test_name", "v": "test_value", "d": "string" }
    ]
  },
  {
    "edgeAttributes": [
      { "po": 1, "n": "test_name", "v": "test_value", "d": "string" }
    ]
  },
  {
    "networkAttributes": [
	  { "n": "test_name", "v": "test_value", "d": "string" }
    ]
  }
]
`

var sampleCXMagicNumber = `
{
  "numberVerification": [
    {"longNumber":281474976710655}
  ]
}
`

var sampleCXMagicNumberIncorrect = `
{
  "numberVerification": [
    {"longNumber":281474976655}
  ]
}
`

var sampleCXMagicNumberMissing = `
{
  "numberVerification": [
  ]
}
`

var sampleCXMetadata = `
{
  "metaData":[
    {"name": "nodes", "elementCount": 2, "properties":[], "version":"1.0"},
    {"name": "edges", "elementCount": 1, "properties":[], "version":"1.0"},
    {"name": "nodeAttributes","properties":[], "version":"1.0"},
    {"name": "edgeAttributes", "properties":[], "version":"1.0"},
    {"name": "networkAttributes", "properties":[], "version":"1.0"}
  ]
}
`

var sampleCXNodeAspect = `
{
  "nodes": [
    { "@id": 1, "n": "test_node_1", "r": "test_data" },
    { "@id": 2, "n": "test_node_2", "r": "test_data" }
  ]
}
`

var sampleCXAnonymousAspectElement = `
{
  "provenanceHistory": [
	 {
		"entity": {
		   "uri": "http://public.ndexbio.org/v2/network/71d5c80b-5226-11e7-8f50-0ac135e8bacf/summary",
		   "creationEvent": {
			  "startedAtTime": 1497571119413,
			  "endedAtTime": 1497571119413,
			  "eventType": "CX network update",
			  "inputs": [
				 {}
			  ],
			  "properties" :[
				 {
					"name": "user",
					"value": "Brett Settle"
				 },
				 {
					"name": "user name",
					"value": "bsettle"
				 }
			  ]
		   },
		   "properties": [
			  {
				 "name": "edge count",
				 "value": 128
			  },
			  {
				 "name": "node count",
				 "value": 138
			  },
			  {
				 "name": "dc:title",
				 "value": "BIOGRID-ORGANISM-Escherichia_coli_K12_MG1655-3.4.129.mitab"
			  }
		   ]
		}
	 }
  ]
}
`

func NewTestDecoder(t *testing.T, json string, s chan<- *Message, aspects ...string) *Decoder {
	r := strings.NewReader(json)
	c := DecoderConfig{
		Label:       "my_network",
		Description: "my test network",
		Aspects:     aspects,
	}
	dec, err := NewDecoder(r, s, c)
	if err != nil {
		t.Error(err)
	}
	return dec
}

//TestReadValidNumberVerification should accept a valid longNumber
func TestReadValidNumberVerification(t *testing.T) {
	dec := NewTestDecoder(t, sampleCXMagicNumber, nil, "test")
	err := dec.DecodeNumberVerification()
	if err != nil {
		t.Error(err)
	}
}

//TestReadInvalidNumberVerifiacation should throw an error because of a missing longNumber
func TestReadInvalidNumberVerification(t *testing.T) {
	dec := NewTestDecoder(t, sampleCXMagicNumberMissing, nil, "test")
	err := dec.DecodeNumberVerification()
	if err == nil {
		t.Error("should not accept a numberverification without a longNumber")
	}
}

//TestReadIncorrectNumberVerifiacation should throw an error because of a longNumber that is not set to the correct value
func TestReadIncorrectNumberVerification(t *testing.T) {
	dec := NewTestDecoder(t, sampleCXMagicNumberIncorrect, nil, "test")
	err := dec.DecodeNumberVerification()
	if err == nil {
		t.Error("should not accept a numberverification with an incorrect longNumber")
	}
}

//TestValidConfig should accept a valid configuration
func TestValidConfig(t *testing.T) {
	s := strings.NewReader("")
	c := DecoderConfig{
		Label:   "my_network",
		Aspects: []string{"test"},
	}
	_, err := NewDecoder(s, nil, c)
	if err != nil {
		t.Error(err)
	}
}

//TestInvalidConfigNoLabel should report an error when the label is missing from a config
func TestInvalidConfigNoLabel(t *testing.T) {
	s := strings.NewReader("")
	c := DecoderConfig{
		Aspects: []string{"test"},
	}
	_, err := NewDecoder(s, nil, c)
	if err == nil {
		t.Error("should not accept DecoderConfig when label is missing")
	}
}

//TestInvalidConfigNoAspects should report an error when aspects is empty. A stream should always have at least one aspect.
func TestInvalidConfigNoAspects(t *testing.T) {
	s := strings.NewReader("")
	c := DecoderConfig{
		Label: "my_network",
	}
	_, err := NewDecoder(s, nil, c)
	if err == nil {
		t.Error("should not accept DecoderConfig when aspects is empty or missing")
	}
}

// TestReadMetdata should test reading a valid config and parsing a valid sampleCXMetdata and should not
// encounter any errors
func TestReadMetadata(t *testing.T) {
	dec := NewTestDecoder(t, sampleCXMetadata, nil, "nodes", "edges", "nodeAttributes", "edgeAttributes", "networkAttributes")
	err := dec.DecodeMetadata()
	if err != nil {
		t.Error(err)
	}
}

// TestReadIncompatibleMetdata should test reading a valid config with a required aspect name that does not
// exist in the test metadata. It should find an error while decoding.
func TestReadIncompatibleMetdata(t *testing.T) {
	dec := NewTestDecoder(t, sampleCXMetadata, nil, "test")
	err := dec.DecodeMetadata()
	if err == nil {
		t.Error("did not encounter an error while decoding metadata that did not contain a required aspect")
	}
}

func TestReadAnonymousAspectElement(t *testing.T) {
	dec := NewTestDecoder(t, sampleCXMetadata, nil, "test")
	err := dec.consumeElement("ndexStatus")
	if err != nil {
		t.Error(err)
	}
	_, err = dec.dec.Token()
	if err != io.EOF {
		t.Error("expected EOF after reading the anonymous aspect element, was the entire element consumed? error: " + err.Error())
	}
}

func TestReadPostMetdata(t *testing.T) {
	dec := NewTestDecoder(t, sampleCXMetadata, nil, "nodes")
	err := dec.DecodeAspect()
	if err != io.EOF {
		t.Error("should have found EOF for post-metadata")
	}
}

func consumeMessages(messages <-chan *Message) {
	for m := range messages {
		close(m.errChan)
	}
}

//TestReadNodeAspect should read a node aspet without errors
func TestReadNodeAspect(t *testing.T) {
	s := make(chan *Message)
	go consumeMessages(s)
	dec := NewTestDecoder(t, sampleCXNodeAspect, s, "nodes")
	err := dec.DecodeAspect()
	close(s)
	if err != nil {
		t.Error(err)
	}
}

func TestReadNetwork(t *testing.T) {
	s := make(chan *Message)
	go consumeMessages(s)
	dec := NewTestDecoder(t, sampleCXNetwork, s, "nodes", "edges")
	err := dec.DecodeNetwork()
	close(s)
	if err != nil {
		t.Error(err)
	}
}

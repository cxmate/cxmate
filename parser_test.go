package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/cxmate/cxmate/proto"
)

func TestParseBracket(t *testing.T) {
	p, _ := MockParser("[{}]")
	if err := p.bracket('[', "opening bracket"); err != nil {
		t.Error(err)
	}
	if err := p.bracket('{', "opening brace"); err != nil {
		t.Error(err)
	}
	if err := p.bracket('}', "closing brace"); err != nil {
		t.Error(err)
	}
	if err := p.bracket(']', "closing bracket"); err != nil {
		t.Error(err)
	}
	if token, err := p.dec.Token(); err != io.EOF {
		t.Errorf("Expected EOF, found %v", token)
	}
}

func TestParseValue(t *testing.T) {
	p, _ := MockParser(`{"key":["value"]}`)
	var example map[string][]string
	if err := p.value(&example, "an example object"); err != nil {
		t.Error(err)
	}
	values, ok := example["key"]
	if !ok {
		t.Errorf("Expected value for key in parsed value")
	}
	if len(values) != 1 {
		t.Fatalf("Expected values to have one value found %d", len(values))
	}
	if values[0] != "value" {
		t.Errorf("Expected value in values to be value found %s", values[0])
	}

}

func TestParseKey(t *testing.T) {
	p, _ := MockParser(`{"key":[]}`)
	if err := p.bracket('{', "open bracket"); err != nil {
		t.Error(err)
	}
	key, err := p.key("a test key")
	if err != nil {
		t.Error(err)
	}
	if key != "key" {
		t.Errorf("Expected key found %s", key)
	}
}

const opaqueElement = `
{
  "provenanceHistory": [
   {
    "entity": {
       "uri": "http://public.ndexbio.org/network/1/summary",
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

func TestParseOpaque(t *testing.T) {
	p, _ := MockParser(opaqueElement)
	if err := p.opaque("provenanceHistory"); err != nil {
		t.Error(err)
	}
	if token, err := p.dec.Token(); err != io.EOF {
		t.Errorf("Expected EOF, found %v", token)
	}
}

const numberVerificationAspect = `
{
  "numberVerification": [
    {"longNumber":281474976710655}
  ]
}
`

func TestParseVerificationNumber(t *testing.T) {
	p, _ := MockParser(numberVerificationAspect)
	if err := p.numberVerification("test network"); err != nil {
		t.Error(err)
	}
	if token, err := p.dec.Token(); err != io.EOF {
		t.Errorf("Expected EOF, found %v", token)
	}
}

const metadataAspect = `
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

func TestParseMetadata(t *testing.T) {
	p, _ := MockParser(metadataAspect)
	if err := p.preMetadata("test network", []string{"nodes"}); err != nil {
		t.Error(err)
	}
	if token, err := p.dec.Token(); err != io.EOF {
		t.Errorf("Expected EOF, found %v", token)
	}
}

func TestParseInvalidMetadata(t *testing.T) {
	p, _ := MockParser(metadataAspect)
	if err := p.preMetadata("test network", []string{"cartesianCoordinates"}); err == nil {
		t.Error("Should not accept preMetadata that does not contain required aspect")
	}
}

func TestParseElement(t *testing.T) {
	p, s := MockParser(`{"@id":3,"n":"node_name"}`)
	go func(t *testing.T, s chan *Message) {
		testNode(t, s, 3, "node_name", "test network")
	}(t, s)
	if err := p.element("test network", "nodes"); err != nil {
		t.Error(err)
	}
}

const nodeAspect = `
{
  "nodes": [
    { "@id": 1, "n": "test_node_1", "r": "test_data" },
    { "@id": 2, "n": "test_node_2", "r": "test_data" }
  ]
}
`

func TestParseAspect(t *testing.T) {
	p, s := MockParser(nodeAspect)
	go func(t *testing.T, s chan *Message) {
		testNode(t, s, 1, "test_node_1", "test network")
		testNode(t, s, 2, "test_node_2", "test network")
	}(t, s)
	if err := p.aspect("test network", []string{"nodes"}); err != nil {
		t.Error(err)
	}
	if token, err := p.dec.Token(); err != io.EOF {
		t.Errorf("Expected EOF, found %v", token)
	}
}

const network = `
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
    "edges": [
      { "@id": 1, "s": 1, "t": 2, "i": "test_connection" }
    ]
  },
  {
    "nodes": [
      { "@id": 1, "n": "test_node_1", "r": "test_data" },
      { "@id": 2, "n": "test_node_2", "r": "test_data" }
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

func TestParseNetwork(t *testing.T) {
	p, s := MockParser(network)
	go func(t *testing.T, s chan *Message) {
		testNode(t, s, 1, "test_node_1", "test network")
		testNode(t, s, 2, "test_node_2", "test network")
	}(t, s)
	if err := p.network("test network", []string{"nodes"}); err != nil {
		t.Error(err)
	}
	if token, err := p.dec.Token(); err != io.EOF {
		t.Errorf("Expected EOF, found %v", token)
	}
}

const rawJSON = `{"key":"test"}`

func TestParseJSON(t *testing.T) {
	p, s := MockParser(rawJSON)
	go func(t *testing.T, s chan *Message) {
		testJSON(t, s, "test raw", rawJSON)
	}(t, s)
	if err := p.rawJSON("test raw"); err != nil {
		t.Errorf("Could not parse raw JSON, error: %v", err)
	}
}

const rawBadJSON = `{`

func TestParseBadJSON(t *testing.T) {
	p, _ := MockParser(rawBadJSON)
	if err := p.rawJSON("test raw"); err == nil {
		t.Fatal("Expected error from reading bad JSON, error was nil")
	}
}

func TestParseStream(t *testing.T) {
	p, s := MockParser(fmt.Sprintf("[%s,%s]", network, network))
	go func(t *testing.T, s chan *Message) {
		testNode(t, s, 1, "test_node_1", "test 1")
		testNode(t, s, 2, "test_node_2", "test 1")
		testNode(t, s, 1, "test_node_1", "test 2")
		testNode(t, s, 2, "test_node_2", "test 2")
	}(t, s)
	d := []NetworkDescription{
		NetworkDescription{
			Label:   "test 1",
			Aspects: []string{"nodes"},
		},
		NetworkDescription{
			Label:   "test 2",
			Aspects: []string{"nodes"},
		},
	}
	if err := p.stream(d); err != nil {
		t.Error(err)
	}
	if token, err := p.dec.Token(); err != io.EOF {
		t.Errorf("Expected EOF, found %v", token)
	}
}

func testNode(t *testing.T, s chan *Message, id int64, name string, network string) {
	m, ok := <-s
	if !ok {
		t.Fatal("Channel should not be closed")
	}
	if m.ele.Label != network {
		t.Errorf("Expected network label %s found %s", network, m.ele.Label)
	}
	node, ok := m.ele.GetElement().(*proto.NetworkElement_Node)
	if !ok {
		t.Fatal("Expected element to be a node")
	}
	if node.Node.Id != id {
		t.Errorf("Expected node id %d found %d", id, node.Node.Id)
	}
	if node.Node.Name != name {
		t.Errorf("Expected node name %s found %s", name, node.Node.Name)
	}
	close(m.errChan)
}

func testJSON(t *testing.T, s chan *Message, label string, expectedJSON string) {
	m, ok := <-s
	if !ok {
		t.Fatal("Channel shoud not be closed")
	}
	if m.ele.Label != label {
		t.Errorf("Expected label %s found %s", label, m.ele.Label)
	}
	rawJSON := m.ele.GetJson()
	if rawJSON != expectedJSON {
		t.Errorf("Expected %s as raw JSON, found %s", expectedJSON, rawJSON)
	}
	close(m.errChan)
}

func MockParser(j string) (*Parser, chan *Message) {
	r := strings.NewReader(j)
	s := make(chan *Message, 100)
	p := &Parser{
		dec:      json.NewDecoder(r),
		sendChan: s,
	}
	return p, s
}

package proto

import (
	"encoding/json"
	"errors"

	"github.com/golang/protobuf/jsonpb"
)

// NetworkElementFromJSON is an extension to handle decoding JSON into the oneOf in NetworkElement
func NetworkElementFromJSON(eleType string, dec *json.Decoder) (*NetworkElement, error) {
	n := &NetworkElement{}
	var err error
	switch eleType {
	case "nodes":
		ele := &Node{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_Node{ele}
	case "edges":
		ele := &Edge{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_Edge{ele}
	case "nodeAttributes":
		ele := &NodeAttribute{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_NodeAttribute{ele}
	case "edgeAttributes":
		ele := &EdgeAttribute{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_EdgeAttribute{ele}
	case "networkAttributes":
		ele := &NetworkAttribute{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_NetworkAttribute{ele}
	case "cartesianLayout":
		ele := &CartesianLayout{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_CartesianLayout{ele}
	default:
		return nil, errors.New("No converter for required aspect type " + eleType)
	}
	return n, err
}

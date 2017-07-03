package proto

import (
	"encoding/json"
	"errors"
	"io"

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

func NetworkElementToJSON(w io.Writer, e *NetworkElement) error {
	m := jsonpb.Marshaler{}
	ele := e.GetElement()
	switch ele.(type) {
	case *NetworkElement_Error:
		m.Marshal(w, e.GetError())
	case *NetworkElement_Node:
		m.Marshal(w, e.GetNode())
	case *NetworkElement_Edge:
		m.Marshal(w, e.GetEdge())
	case *NetworkElement_NodeAttribute:
		m.Marshal(w, e.GetNodeAttribute())
	case *NetworkElement_EdgeAttribute:
		m.Marshal(w, e.GetEdgeAttribute())
	case *NetworkElement_NetworkAttribute:
		m.Marshal(w, e.GetNetworkAttribute())
	default:
		return errors.New("No generator available")
	}
	return nil
}

func GetAspectName(e *NetworkElement) string {
	ele := e.GetElement()
	switch ele.(type) {
	case *NetworkElement_Error:
		return "error"
	case *NetworkElement_Node:
		return "nodes"
	case *NetworkElement_Edge:
		return "edges"
	case *NetworkElement_NodeAttribute:
		return "nodeAttributes"
	case *NetworkElement_EdgeAttribute:
		return "edgeAttributes"
	case *NetworkElement_NetworkAttribute:
		return "networkAttributes"
	default:
		return "unknown"
	}
}

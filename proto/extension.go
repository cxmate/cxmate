package proto

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/golang/protobuf/jsonpb"
)

// NetworkElementFromJSON is an extension to handle decoding JSON into the oneOf in NetworkElement
func NetworkElementFromJSON(label string, eleType string, dec *json.Decoder) (*NetworkElement, error) {
	n := &NetworkElement{Label: label}
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
		ele := &CartesianCoordinate{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_CartesianCoordinate{ele}
	case "cyGroups":
		ele := &CyGroup{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_CyGroup{ele}
	case "cyViews":
		ele := &CyView{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_CyView{ele}
	case "cyVisualProperties":
		ele := &CyVisualProperty{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_CyVisualProperty{ele}
	case "cyHiddenAttributes":
		ele := &CyHiddenAttribute{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_CyHiddenAttribute{ele}
	case "cyNetworkRelations":
		ele := &CyNetworkRelation{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_CyNetworkRelation{ele}
	case "cySubNetworks":
		ele := &CySubNetwork{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_CySubNetwork{ele}
	case "cyTableColumns":
		ele := &CyTableColumn{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_CyTableColumn{ele}
	case "ndexStatus":
		ele := &NdexStatus{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_NdexStatus{ele}
	case "citations":
		ele := &Citation{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_Citation{ele}
	case "nodeCitations":
		ele := &NodeCitations{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_NodeCitations{ele}
	case "edgeCitations":
		ele := &EdgeCitations{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_EdgeCitations{ele}
	case "supports":
		ele := &Support{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_Support{ele}
	case "nodeSupports":
		ele := &NodeSupportance{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_NodeSupportance{ele}
	case "edgeSupports":
		ele := &EdgeSupportance{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_EdgeSupportance{ele}
	case "functionTerms":
		ele := &FunctionTerm{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_FunctionTerm{ele}
	case "reifiedEdges":
		ele := &ReifiedEdge{}
		err = jsonpb.UnmarshalNext(dec, ele)
		n.Element = &NetworkElement_ReifiedEdge{ele}
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
	case *NetworkElement_CartesianCoordinate:
		m.Marshal(w, e.GetCartesianCoordinate())
	case *NetworkElement_CyGroup:
		m.Marshal(w, e.GetCyGroup())
	case *NetworkElement_CyView:
		m.Marshal(w, e.GetCyView())
	case *NetworkElement_CyVisualProperty:
		m.Marshal(w, e.GetCyVisualProperty())
	case *NetworkElement_CyHiddenAttribute:
		m.Marshal(w, e.GetCyHiddenAttribute())
	case *NetworkElement_CyNetworkRelation:
		m.Marshal(w, e.GetCyNetworkRelation())
	case *NetworkElement_CySubNetwork:
		m.Marshal(w, e.GetCySubNetwork())
	case *NetworkElement_CyTableColumn:
		m.Marshal(w, e.GetCyTableColumn())
	case *NetworkElement_NdexStatus:
		m.Marshal(w, e.GetNdexStatus())
	case *NetworkElement_Citation:
		m.Marshal(w, e.GetCitation())
	case *NetworkElement_NodeCitations:
		m.Marshal(w, e.GetNodeCitations())
	case *NetworkElement_EdgeCitations:
		m.Marshal(w, e.GetEdgeCitations())
	case *NetworkElement_Support:
		m.Marshal(w, e.GetSupport())
	case *NetworkElement_NodeSupportance:
		m.Marshal(w, e.GetNodeSupportance())
	case *NetworkElement_EdgeSupportance:
		m.Marshal(w, e.GetEdgeSupportance())
	case *NetworkElement_FunctionTerm:
		m.Marshal(w, e.GetFunctionTerm())
	case *NetworkElement_ReifiedEdge:
		m.Marshal(w, e.GetReifiedEdge())
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
	case *NetworkElement_CartesianCoordinate:
		return "cartesianLayout"
	case *NetworkElement_CyGroup:
		return "cyGroups"
	case *NetworkElement_CyView:
		return "cyViews"
	case *NetworkElement_CyVisualProperty:
		return "cyVisualProperties"
	case *NetworkElement_CyHiddenAttribute:
		return "cyHiddenAttributes"
	case *NetworkElement_CyNetworkRelation:
		return "cyNetworkRelations"
	case *NetworkElement_CySubNetwork:
		return "cySubNetworks"
	case *NetworkElement_CyTableColumn:
		return "cyTableColumns"
	case *NetworkElement_NdexStatus:
		return "ndexStatus"
	case *NetworkElement_Citation:
		return "citations"
	case *NetworkElement_NodeCitations:
		return "nodeCitations"
	case *NetworkElement_EdgeCitations:
		return "edgeCitations"
	case *NetworkElement_Support:
		return "supports"
	case *NetworkElement_NodeSupportance:
		return "nodeSupports"
	case *NetworkElement_EdgeSupportance:
		return "edgeSupports"
	case *NetworkElement_FunctionTerm:
		return "functionTerms"
	case *NetworkElement_ReifiedEdge:
		return "reifiedEdges"
	default:
		return "unknown"
	}
}

package main

import (
	"fmt"
	"strconv"

	"github.com/cxmate/cxmate/proto"
)

// ParameterConfig contains a series of Parameter definitions that cxMate will use to convert
// query string parameters into parameter messages that the service will consume. All parameters
// define a default value that will be sent to the service if the parameter does not appear in the
// query string. If a query string contains the same key twice that corrosponds to a parameter, multiple
// messages with the same parameter name will be sent with each of the provided values.
type ParameterConfig []Parameter

// validate validates the ParameterConfig by calling validate on each provided Parameter definition.
func (params ParameterConfig) validate() error {
	for _, p := range params {
		if err := p.validate(); err != nil {
			return err
		}
	}
	return nil
}

// send merges the query string parameters into the ParameterConfig, sending either the values
// in the query string or the provided default value to the service for each parameter.
func (params ParameterConfig) send(send chan *Message, query map[string][]string) error {
	for _, p := range params {
		if values, ok := query[p.Name]; ok {
			for _, value := range values {
				if err := p.send(send, value); err != nil {
					return err
				}
			}
		} else {
			if err := p.send(send, p.Default); err != nil {
				return err
			}
		}
	}
	return nil
}

// Parameter describes a parameter message that can be sent to the service. The type describes
// the JSON representation of the Parameter, while the format describes a more detailed representation
// for the service such as an unsigned integer, password, or complex object. The description will be used
// for the service specification.
type Parameter struct {
	Name        string `json:"name"`
	Default     string `json:"default"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Format      string `json:"format"`
}

// validate determines if a Parameter is valid by checking if:
// - The name field is not empty
// - The Default field is not empty
// - Description is not empty
// - Type is one of string, integer, boolean, number or is empty
// - The Default value must be castable to the provided type in Type
func (p Parameter) validate() error {
	if p.Name == "" {
		return fmt.Errorf("name is a required field")
	}
	if p.Default == "" {
		return fmt.Errorf("default is a required field")
	}
	if p.Description == "" {
		return fmt.Errorf("description is a required field")
	}
	if p.Type != "" {
		ok := false
		accepted := []string{"integer", "number", "boolean", "string"}
		for _, t := range accepted {
			if p.Type == t {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("expected type to be one of %v found %s", accepted, p.Type)
		}
	}
	if _, err := p.convert(p.Default); err != nil {
		return fmt.Errorf("default value must be convertable to specified type or string if no type is provided error: %v", err)
	}
	return nil
}

// send builds a message to send to the service from the parameter, using value as the parameter value
// that will be converted to the parameter type before being sent. send will block until the message has
// been received and the error value from the service has been returned.
func (p Parameter) send(send chan *Message, value string) error {
	param, err := p.convert(value)
	if err != nil {
		return err
	}
	errChan := make(chan error)
	message := &Message{
		ele: &proto.NetworkElement{
			Element: &proto.NetworkElement_Parameter{
				Parameter: param,
			},
		},
		errChan: errChan,
	}
	send <- message
	return <-errChan
}

// convert wraps the parameter value in the type specified by the Parameter definition. If no type
// was supplied, the val will be interpreted as a string.
func (p *Parameter) convert(val string) (*proto.Parameter, error) {
	pp := &proto.Parameter{Name: p.Name, Format: p.Format}
	switch p.Type {
	case "number":
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			break
		}
		pp.Value = &proto.Parameter_NumberValue{NumberValue: v}
		return pp, nil
	case "boolean":
		v, err := strconv.ParseBool(val)
		if err != nil {
			break
		}
		pp.Value = &proto.Parameter_BooleanValue{BooleanValue: v}
		return pp, nil
	case "integer":
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			break
		}
		pp.Value = &proto.Parameter_IntegerValue{IntegerValue: v}
		return pp, nil
	default:
		pp.Value = &proto.Parameter_StringValue{StringValue: val}
		return pp, nil
	}
	return nil, fmt.Errorf("cannot convert parameter %s with value %s to type %s", p.Name, val, p.Type)
}

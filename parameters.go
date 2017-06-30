package main

import (
	"github.com/ericsage/cxmate/proto"
)

//ParameterConfig configures an query string key that will be accepted as a parameter to an algorithm.
type ParameterConfig struct {
	//Key is the name of the parameter, which will be used to set and retrieve tha value of the parameter.
	Key string `json:"key"`
	//Default will be the value of the parameter if a user does not explicity set it.
	Default string `json:"default"`
	//Description gives a short description of the prameter, such as which its affect is on the algorithm.
	Description string `json:"description"`
}

// processParameters uses a parameter config to send Parameter NetworkElements to the service,
// looking for values in the query string parameters received by cxMate, or sending a default value
// supplied by the config
func (m *Mate) processParameters(send chan *Message, params map[string][]string) error {
	logDebugln("processing query string parameters")
	requestedParams := m.Config.Service.Parameters
	if requestedParams != nil {
		for _, param := range requestedParams {
			if values, ok := params[param.Key]; ok {
				for _, value := range values {
					logDebugln("Sending query string value", value, "for key", param.Key)
					pm := newParameterElement(param.Key, value)
					if err := sendParameter(send, pm); err != nil {
						return err
					}
				}
			} else {
				logDebugln("Sending default value", param.Default, "for key", param.Key)
				pm := newParameterElement(param.Key, param.Default)
				if err := sendParameter(send, pm); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// sendParameter sneds a Parameter NetworkElement to the service, and waits for an error
func sendParameter(send chan *Message, parameter *proto.NetworkElement) error {
	errChan := make(chan error)
	message := &Message{
		ele:     parameter,
		errChan: errChan,
	}
	send <- message
	err, ok := <-errChan
	if ok {
		return nil
	}
	return err
}

// newParameterElement creates and returns a pointer to new protobuf Parameter NetworkElement
func newParameterElement(key string, value string) *proto.NetworkElement {
	return &proto.NetworkElement{
		Element: &proto.NetworkElement_Parameter{
			Parameter: &proto.Parameter{
				Name:  key,
				Value: value,
			},
		},
	}
}

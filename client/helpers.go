package client

import (
	"github.com/mitchellh/mapstructure"
)

type StoppedPayload struct {
	Frame StoppedFrame
}
type StoppedFrame struct {
	Address      string `mapstructure:"addr"`
	Architecture string `mapstructure:"arch"`
}

func (gdb *GdbClient) ParseFrame(frame map[string]any) (*StoppedFrame, error) {
	var decodedPayload StoppedPayload

	err := mapstructure.Decode(frame["payload"], &decodedPayload)
	if err != nil {
		return nil, err
	}

	return &decodedPayload.Frame, err
}

func (gdb *GdbClient) Interrupt() error {
	return gdb.gdb.Interrupt()
}

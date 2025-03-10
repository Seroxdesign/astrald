package gw

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
)

func (drv *Driver) Unpack(network string, data []byte) (net.Endpoint, error) {
	if network != DriverName {
		return nil, errors.New("invalid network")
	}
	return Unpack(data)
}

// Unpack converts a binary representation of the address to a struct
func Unpack(data []byte) (addr Endpoint, err error) {
	return addr, cslq.Decode(bytes.NewReader(data), "vv", &addr.gate, &addr.target)
}

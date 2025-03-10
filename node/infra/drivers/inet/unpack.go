package inet

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
)

var _ infra.Unpacker = &Driver{}

func (drv *Driver) Unpack(network string, data []byte) (net.Endpoint, error) {
	if network != DriverName {
		return nil, errors.New("invalid network")
	}
	return Unpack(data)
}

func Unpack(buf []byte) (addr Endpoint, err error) {
	var r = bytes.NewReader(buf)

	if err = cslq.Decode(r, "c", &addr.ver); err != nil {
		return
	}

	switch addr.ver {
	case ipv4:
		return addr, cslq.Decode(r, "[4]c s", &addr.ip, &addr.port)
	case ipv6:
		return addr, cslq.Decode(r, "[16]c s", &addr.ip, &addr.port)
	}

	return addr, errors.New("invalid version")
}

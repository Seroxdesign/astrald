package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/query"
	"io"
)

type RelayRouter struct {
	log      *log.Logger
	target   string
	identity id.Identity
}

func (fwd *RelayRouter) RouteQuery(ctx context.Context, query query.Query, swc net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	c, err := proto.Dial(fwd.target)
	if err != nil {
		fwd.log.Errorv(2, "%s:%s forward to %s: %s", query.Target(), query.Query(), fwd.target, err)
		return nil, err
	}

	conn := proto.NewConn(c)

	err = conn.WriteMsg(proto.InQueryParams{
		Identity: query.Caller(),
		Query:    query.Query(),
	})
	if err != nil {
		return nil, err
	}

	if conn.ReadErr() != nil {
		return nil, err
	}

	go func() {
		io.Copy(swc, c)
		swc.Close()
	}()

	return net.NewSecureWriteCloser(c, query.Target()), err
}

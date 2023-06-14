package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/storage/rpc"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/query"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ tasks.Runner = &RegisterService{}

const RegisterServiceName = "storage.register"

type RegisterService struct {
	*Module
}

func (service *RegisterService) Run(ctx context.Context) error {
	s, err := service.node.Services().Register(ctx, service.node.Identity(), RegisterServiceName, service)
	if err != nil {
		service.log.Error("cannot register service %s: %s", RegisterServiceName, err)
		return err
	}

	<-s.Done()

	return nil
}

func (service *RegisterService) RouteQuery(ctx context.Context, q query.Query, remoteWriter net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	if !service.IsProvider(q.Caller()) {
		service.log.Errorv(2, "register_provider: %v is not a provider, rejecting...", q.Caller())
		return nil, link.ErrRejected
	}

	return query.Accept(q, remoteWriter, func(conn net.SecureConn) {
		service.handle(service.ctx, conn)
	})
}

func (service *RegisterService) handle(ctx context.Context, conn net.SecureConn) error {
	defer conn.Close()
	return cslq.Invoke(conn, func(msg rpc.MsgRegisterSource) error {
		var session = rpc.New(conn)

		source := &DataSource{
			Service:  msg.Service,
			Identity: conn.RemoteIdentity(),
		}

		service.AddDataSource(source)

		return session.EncodeErr(nil)
	})
}

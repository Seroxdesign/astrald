package reflectlink

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/reflectlink/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/modules"
	"reflect"
)

type Server struct {
	*Module
}

func (server *Server) Run(ctx context.Context) error {
	s, err := server.node.Services().Register(ctx, server.node.Identity(), serviceName, server)
	if err != nil {
		return err
	}

	disco, err := modules.Find[*discovery.Module](server.node.Modules())
	if err == nil {
		disco.AddSourceContext(ctx, server, server.node.Identity())
	}

	<-s.Done()

	return nil
}

func (server *Server) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	// Reject queries coming from sources without transport
	var output = net.FinalOutput(caller)
	var t, ok = output.(net.Transporter)
	if !ok {
		return net.Reject()
	}
	if t.Transport() == nil {
		return net.Reject()
	}

	return net.Accept(query, caller, server.reflect)
}

func (server *Server) reflect(conn net.SecureConn) {
	defer conn.Close()

	var output, _ = net.FinalOutput(conn).(net.Transporter)
	var endpoint = output.Transport().RemoteEndpoint()

	if endpoint == nil {
		server.log.Errorv(2, "link with %v has no remote endpoint (transport type %v)",
			conn.RemoteIdentity(),
			reflect.TypeOf(output.Transport()),
		)
		return
	}

	var reflection proto.Reflection

	if endpoint != nil {
		reflection.RemoteEndpoint = proto.Endpoint{
			Network: endpoint.Network(),
			Address: endpoint.String(),
		}
	}

	json.NewEncoder(conn).Encode(reflection)

	server.log.Infov(2, "reflected %v", conn.RemoteIdentity())

	return
}

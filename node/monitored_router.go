package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
)

const MonitoredConnHint = "monitored_conn"

type MonitoredRouter struct {
	net.Router
	conns  *ConnSet
	events events.Queue
}

func (router *MonitoredRouter) Conns() *ConnSet {
	return router.conns
}

func NewMonitoredRouter(router net.Router, eventParent *events.Queue) *MonitoredRouter {
	r := &MonitoredRouter{
		Router: router,
		conns:  NewConnSet(),
	}
	r.events.SetParent(eventParent)
	return r
}

func (router *MonitoredRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (target net.SecureWriteCloser, err error) {
	// check DontMonitor flag
	if hints.DontMonitor {
		return router.Router.RouteQuery(ctx, query, caller, hints)
	}

	// monitor the caller
	var callerMonitor = NewMonitoredWriter(caller)

	// prepare the en route connection
	var conn = NewMonitoredConn(callerMonitor, nil, query, hints)
	hints = hints.WithValue(MonitoredConnHint, conn)
	router.conns.Add(conn)

	// route to next hop
	target, err = router.Router.RouteQuery(ctx, query, callerMonitor, hints)
	if err != nil {
		router.conns.Remove(conn)
		return net.RouteNotFound(router, err)
	}

	// monitor the target
	var targetMonitor = NewMonitoredWriter(target)
	conn.SetTarget(targetMonitor)

	router.events.Emit(EventConnAdded{Conn: conn})

	// remove the conn after it's closed
	go func() {
		<-conn.Done()
		router.conns.Remove(conn)
		router.events.Emit(EventConnRemoved{Conn: conn})
	}()

	return targetMonitor, err
}

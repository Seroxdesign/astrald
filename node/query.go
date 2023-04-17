package node

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/streams"
	"time"
)

func (node *CoreNode) Query(ctx context.Context, remoteID id.Identity, query string) (link.BasicConn, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	if remoteID.IsZero() || remoteID.IsEqual(node.identity) {
		return node.Services().Query(ctx, query, nil)
	}

	link, err := node.Network().Link(ctx, remoteID)
	if err != nil {
		return nil, err
	}

	return link.Query(ctx, query)
}

func (node *CoreNode) onQuery(query *link.Query) error {
	select {
	case node.queryQueue <- query:
	default:
		log.Error("query dropped due to queue overflow: %s", query.Query())
		return errors.New("query queue overflow")
	}
	return nil
}

func (node *CoreNode) peerQueryWorker(ctx context.Context) error {
	for {
		select {
		case query := <-node.queryQueue:
			ctx, _ := context.WithTimeout(ctx, defaultQueryTimeout)
			var start = time.Now()
			var err = node.executeQuery(ctx, query)
			var elapsed = time.Since(start)

			log.Logv(2, "served query %s (time %s, err %s)",
				query.Query(),
				elapsed.Round(time.Microsecond),
				err,
			)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (node *CoreNode) executeQuery(ctx context.Context, query *link.Query) error {
	// Query a session with the service
	localConn, err := node.Services().Query(ctx, query.Query(), query.Link())
	if err != nil {
		query.Reject()
		return err
	}

	// Accept remote party's query
	remoteConn, err := query.Accept()
	if err != nil {
		localConn.Close()
		return err
	}

	go streams.Join(localConn, remoteConn)

	return nil
}

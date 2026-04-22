package driven

import (
	"context"
	"enfok_bd/src/domain/node"
)

type NodeRepository interface {
	Save(ctx context.Context, n *node.Node) error
	GetByID(ctx context.Context, id int) (*node.Node, error)
	GetByNodeID(ctx context.Context, nodeID string) (*node.Node, error)
	UpdateStatus(ctx context.Context, nodeID string, status string) error
	UpdateSignal(ctx context.Context, nodeID string) error
	ListAll(ctx context.Context) ([]node.Node, error)
}

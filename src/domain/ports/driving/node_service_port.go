package driving

import (
	"context"

	"enfok_bd/src/domain/node"
)

// NodeServicePort define las operaciones que el API REST puede invocar
// sobre el dominio de nodos. Es el contrato de entrada (driving).
type NodeServicePort interface {
	RegisterNode(ctx context.Context, n *node.Node) error
	Heartbeat(ctx context.Context, nodeID string) error
	GetNode(ctx context.Context, nodeID string) (*node.Node, error)
	ListNodes(ctx context.Context) ([]node.Node, error)
	UpdateStatus(ctx context.Context, nodeID string, status string) error
}

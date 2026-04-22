package service

import (
	"context"
	"time"

	"enfok_bd/src/domain/node"
	"enfok_bd/src/domain/ports/driven"
)

type NodeService struct {
	repo driven.NodeRepository
}

func NewNodeService(r driven.NodeRepository) *NodeService {
	return &NodeService{repo: r}
}

func (s *NodeService) RegisterNode(ctx context.Context, n *node.Node) error {
	if n.LastSignal.IsZero() {
		n.LastSignal = time.Now()
	}
	if n.Status == "" {
		n.Status = "ACTIVE"
	}
	return s.repo.Save(ctx, n)
}

func (s *NodeService) Heartbeat(ctx context.Context, nodeID string) error {
	return s.repo.UpdateSignal(ctx, nodeID)
}

func (s *NodeService) GetNode(ctx context.Context, nodeID string) (*node.Node, error) {
	return s.repo.GetByNodeID(ctx, nodeID)
}

func (s *NodeService) ListNodes(ctx context.Context) ([]node.Node, error) {
	return s.repo.ListAll(ctx)
}

func (s *NodeService) UpdateStatus(ctx context.Context, nodeID string, status string) error {
	return s.repo.UpdateStatus(ctx, nodeID, status)
}

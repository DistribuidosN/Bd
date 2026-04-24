package repository

import (
	"context"
	"time"

	"enfok_bd/src/domain/node"
	"enfok_bd/src/domain/ports/driven"
	"enfok_bd/src/infrastructure/dto"
	"enfok_bd/src/utils"

	"github.com/jmoiron/sqlx"
)

type postgresNodeRepository struct {
	db *sqlx.DB
}

func NewPostgresNodeRepository(db *sqlx.DB) driven.NodeRepository {
	return &postgresNodeRepository{db: db}
}

func (r *postgresNodeRepository) Save(ctx context.Context, n *node.Node) error {
	d := dto.NodeDTO{
		NodeID:     n.NodeID,
		Host:       n.Host,
		Port:       n.Port,
		StatusID:   utils.GetIDFromStatus(utils.NodeStatuses, n.Status),
		LastSignal: n.LastSignal,
	}
	query := `INSERT INTO nodes (node_id, host, port, status_id, last_signal) 
			  VALUES (:node_id, :host, :port, :status_id, :last_signal)
			  ON CONFLICT (node_id) DO UPDATE SET 
			  host = EXCLUDED.host, port = EXCLUDED.port, 
			  status_id = EXCLUDED.status_id, last_signal = EXCLUDED.last_signal`
	_, err := r.db.NamedExecContext(ctx, query, d)
	return err
}

func (r *postgresNodeRepository) GetByNodeID(ctx context.Context, nodeID string) (*node.Node, error) {
	var d dto.NodeDTO
	if err := r.db.GetContext(ctx, &d, `SELECT id, node_id, host, port, status_id, last_signal FROM nodes WHERE node_id = $1`, nodeID); err != nil {
		return nil, err
	}
	return toNodeModel(d), nil
}

func (r *postgresNodeRepository) GetByID(ctx context.Context, id int) (*node.Node, error) {
	var d dto.NodeDTO
	if err := r.db.GetContext(ctx, &d, `SELECT id, node_id, host, port, status_id, last_signal FROM nodes WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return toNodeModel(d), nil
}

func (r *postgresNodeRepository) UpdateStatus(ctx context.Context, nodeID string, status string) error {
	statusID := utils.GetIDFromStatus(utils.NodeStatuses, status)
	_, err := r.db.ExecContext(ctx, `UPDATE nodes SET status_id = $1 WHERE node_id = $2`, statusID, nodeID)
	return err
}

func (r *postgresNodeRepository) UpdateSignal(ctx context.Context, nodeID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE nodes SET last_signal = $1 WHERE node_id = $2`, time.Now(), nodeID)
	return err
}

func (r *postgresNodeRepository) ListAll(ctx context.Context) ([]node.Node, error) {
	var dtos []dto.NodeDTO
	if err := r.db.SelectContext(ctx, &dtos, `SELECT id, node_id, host, port, status_id, last_signal FROM nodes`); err != nil {
		return make([]node.Node, 0), nil
	}
	result := make([]node.Node, 0)
	for _, d := range dtos {
		result = append(result, *toNodeModel(d))
	}
	return result, nil
}

func toNodeModel(d dto.NodeDTO) *node.Node {
	return &node.Node{
		ID:         d.ID,
		NodeID:     d.NodeID,
		Host:       d.Host,
		Port:       d.Port,
		Status:     utils.GetStatusFromID(utils.NodeStatuses, d.StatusID),
		LastSignal: d.LastSignal,
	}
}

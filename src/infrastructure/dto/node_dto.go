package dto

import "time"

// Status IDs para nodos — deben coincidir con la tabla estatus_nodo en BD.
const (
	NodeStatusActive   = 1
	NodeStatusInactive = 2
	NodeStatusError    = 3
)

type NodeDTO struct {
	ID         int       `db:"id"`
	NodeID     string    `db:"node_id"`
	Host       string    `db:"host"`
	Port       int       `db:"port"`
	StatusID   int       `db:"status_id"`
	LastSignal time.Time `db:"last_signal"`
}

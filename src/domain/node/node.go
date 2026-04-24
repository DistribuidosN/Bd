package node

import "time"

type Node struct {
	ID         int       `json:"id"`
	NodeID     string    `json:"node_id"`
	Host       string    `json:"host"`
	Port       int       `json:"port"`
	Status     string    `json:"status"`
	LastSignal time.Time `json:"last_signal"`
}

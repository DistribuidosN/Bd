package logs

import "time"

type ProcessingLog struct {
	ID        int       `json:"id"`
	NodeID    string    `json:"node_id"`
	ImageUUID *string   `json:"image_uuid,omitempty"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	LogTime   *time.Time `json:"log_time" json:"logTime"`
}

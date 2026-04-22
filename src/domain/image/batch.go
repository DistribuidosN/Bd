package image

import "time"

type Batch struct {
	BatchUUID   string    `json:"batch_uuid"`
	UserUUID    string    `json:"user_uuid"`
	RequestTime time.Time `json:"request_time"`
	Status      string    `json:"status"`
}

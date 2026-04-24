package image

import "time"

type Batch struct {
	BatchUUID   string    `json:"batch_uuid" json:"batchUuid"`
	UserUUID    string    `json:"user_uuid" json:"userUuid"`
	RequestTime time.Time `json:"request_time" json:"requestTime"`
	Status      string    `json:"status"`
	StatusID    int       `json:"status_id" json:"statusId"`
}

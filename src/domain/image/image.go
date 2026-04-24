package image

import "time"

type Image struct {
	ImageUUID      string     `json:"image_uuid" json:"imageUuid"`
	BatchUUID      string     `json:"batch_uuid" json:"batchUuid"`
	OriginalName   string     `json:"original_name" json:"originalName"`
	ResultPath     *string    `json:"result_path,omitempty" json:"resultPath,omitempty"`
	Status         string     `json:"status"`
	StatusID       int        `json:"status_id" json:"statusId"`
	NodeID         string     `json:"node_id" json:"nodeId"`
	ReceptionTime  time.Time  `json:"reception_time" json:"receptionTime"`
	ConversionTime *time.Time `json:"conversion_time,omitempty" json:"conversionTime,omitempty"`
}

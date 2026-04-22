package image

import "time"

type Image struct {
	ImageUUID      string     `json:"image_uuid"`
	BatchUUID      string     `json:"batch_uuid"`
	OriginalName   string     `json:"original_name"`
	ResultPath     *string    `json:"result_path,omitempty"`
	Status         string     `json:"status"`
	NodeID         *int       `json:"node_id,omitempty"`
	ReceptionTime  time.Time  `json:"reception_time"`
	ConversionTime *time.Time `json:"conversion_time,omitempty"`
}

package image

// Transformation representa un tipo de transformación disponible en el sistema.
type Transformation struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

// ImageTransformation representa una transformación aplicada a una imagen específica.
type ImageTransformation struct {
	ID        int    `json:"id"`
	ImageUUID string `json:"image_uuid"`
	TypeID    int    `json:"type_id"`
	Params    string `json:"params"`
}

// BatchTransformation define el pipeline de transformaciones para un lote.
type BatchTransformation struct {
	BatchUUID      string `json:"batch_uuid"`
	TypeID         int    `json:"type_id"`
	Params         string `json:"params"`
	ExecutionOrder int    `json:"execution_order"`
}

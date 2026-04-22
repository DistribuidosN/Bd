package dto

type TransformationDTO struct {
	ID          int     `db:"id"`
	Name        string  `db:"name"`
	Price       float64 `db:"price"`
	Description string  `db:"description"`
}

type ImageTransformationDTO struct {
	ID        int    `db:"id"`
	ImageUUID string `db:"image_uuid"`
	TypeID    int    `db:"type_id"`
	Params    string `db:"params"`
}

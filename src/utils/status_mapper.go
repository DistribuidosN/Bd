package utils

import "strings"

// Catalog mappings based on bdPosgres.sql comments
var (
	NodeStatuses = map[string]int{
		"ACTIVE":   1,
		"INACTIVE": 2,
		"IDLE":     3,
		"BUSY":     4,
		"STEALING": 5,
		"ERROR":    6,
	}
	BatchStatuses = map[string]int{
		"PENDING":    1,
		"PROCESSING": 2,
		"COMPLETED":  3,
		"FAILED":     4,
	}
	ImageStatuses = map[string]int{
		"RECEIVED":   1,
		"PROCESSING": 2,
		"CONVERTED":  3,
		"FAILED":     4,
	}
	LogLevels = map[string]int{
		"INFO":    1,
		"WARNING": 2,
		"ERROR":   3,
		"FATAL":   4,
	}
	TransformationTypes = map[string]int{
		"GRAYSCALE": 1,
		"RESIZE":    2,
		"BLUR":      3,
		"CROP":      4,
	}
)

func GetIDFromStatus(catalog map[string]int, status string) int {
	if id, ok := catalog[strings.ToUpper(status)]; ok {
		return id
	}
	return 1 // Default to first status if not found
}

func GetStatusFromID(catalog map[string]int, id int) string {
	for name, catalogID := range catalog {
		if catalogID == id {
			return name
		}
	}
	return "UNKNOWN"
}

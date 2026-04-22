package utils

import "strings"

// Catalog mappings based on bdPosgres.sql comments
var (
	NodeStatuses = map[string]int{
		"ACTIVE":   1,
		"INACTIVE": 2,
		"ERROR":    3,
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

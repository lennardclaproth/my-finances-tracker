package api

// RowError represents a row-level import error.
type RowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

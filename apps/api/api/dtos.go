package api

type RowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

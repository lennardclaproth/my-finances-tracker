package api

import (
	"mime/multipart"
	"net/textproto"

	"github.com/google/uuid"
)

type ImportCsv struct {
	File     multipart.File       `multipart:"file"`
	Filename string               `multipart:"filename"`
	Size     int64                `multipart:"size"`
	Header   textproto.MIMEHeader `multipart:"header"`
	VendorID string               `form:"vendor_id"`
}

type GetUntaggedTransactionsRequest struct {
	Page     int `json:"page" query:"page"`
	PageSize int `json:"page_size" query:"page_size"`
}

type TagTransactionRequest struct {
	Id  uuid.UUID `json:"id"`
	Tag string    `json:"tag"`
}

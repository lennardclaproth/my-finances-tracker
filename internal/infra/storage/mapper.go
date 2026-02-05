package storage

import "github.com/lennardclaproth/my-finances-tracker/internal/domain"

func toDomainTransaction(tx TransactionRecord) domain.Transaction {
	return domain.Transaction{
		ID:          tx.ID,
		Description: tx.Description,
		Note:        tx.Note,
		Source:      tx.Source,
		AmountCents: tx.AmountCents,
		Direction:   tx.Direction,
		Date:        tx.Date,
		Checksum:    tx.Checksum,
		CreatedAt:   tx.CreatedAt,
		UpdatedAt:   tx.UpdatedAt,
		Tag:         tx.Tag,
		Ignored:     tx.Ignored,
		RowNumber:   tx.RowNumber,
		ImportID:    tx.ImportID,
	}
}

func fromDomainTransaction(tx *domain.Transaction) TransactionRecord {
	return TransactionRecord{
		ID:          tx.ID,
		Description: tx.Description,
		Note:        tx.Note,
		Source:      tx.Source,
		AmountCents: tx.AmountCents,
		Direction:   tx.Direction,
		Date:        tx.Date,
		Checksum:    tx.Checksum,
		CreatedAt:   tx.CreatedAt,
		UpdatedAt:   tx.UpdatedAt,
		Tag:         tx.Tag,
	}
}

func toDomainImport(im ImportRecord, vendor domain.Vendor) domain.Import {
	return domain.Import{
		ID:         im.ID,
		CreatedAt:  im.CreatedAt,
		UpdatedAt:  im.UpdatedAt,
		Vendor:     vendor,
		Status:     im.Status,
		StatusMsg:  im.StatusMsg,
		Duplicates: im.Duplicates,
		TotalRows:  im.TotalRows,
		Imported:   im.Imported,
		Failed:     im.Failed,
		Path:       im.Path,
	}
}

func fromDomainImport(im *domain.Import) ImportRecord {
	return ImportRecord{
		ID:         im.ID,
		CreatedAt:  im.CreatedAt,
		UpdatedAt:  im.UpdatedAt,
		VendorID:   im.Vendor.ID,
		Status:     im.Status,
		StatusMsg:  im.StatusMsg,
		Duplicates: im.Duplicates,
		TotalRows:  im.TotalRows,
		Imported:   im.Imported,
		Failed:     im.Failed,
		Path:       im.Path,
	}
}

func toDomainVendor(v VendorRecord) domain.Vendor {
	return domain.Vendor{
		ID:        v.ID,
		Name:      v.Name,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

func fromDomainVendor(v *domain.Vendor) VendorRecord {
	return VendorRecord{
		ID:        v.ID,
		Name:      v.Name,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

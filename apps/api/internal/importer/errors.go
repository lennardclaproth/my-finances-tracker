package importer

import "fmt"

var (
	// ErrNoImportsPending indicates no import rows matched the query.
	ErrNoImportsPending = fmt.Errorf("no imports pending")
	// ErrImportFileRequired indicates an import command was called without file contents.
	ErrImportFileRequired = fmt.Errorf("import file is required")
	// ErrImportPathRequired indicates an import record cannot be created without a stored file path.
	ErrImportPathRequired = fmt.Errorf("import path is required")
	// ErrImportTypeRequired indicates an import record cannot be created without an import type.
	ErrImportTypeRequired = fmt.Errorf("import type is required")
	// ErrUnsupportedImportType indicates the requested import type is not supported.
	ErrUnsupportedImportType = fmt.Errorf("unsupported import type")
	// ErrAccountIDRequired indicates imports require a non-nil account ID.
	ErrAccountIDRequired = fmt.Errorf("account_id is required for imports")
	// ErrAccountNotExists indicates that the account does not exist
	ErrAccountNotExists = fmt.Errorf("account does not exist")
	// ErrVendorIDRequired indicates vendor-backed imports require a vendor ID.
	ErrVendorIDRequired = fmt.Errorf("vendor_id is required for imports")
	// ErrListingIDRequired indicates EOD imports require a listing ID.
	ErrListingIDRequired = fmt.Errorf("listing_id is required for imports")
	// ErrVendorImportDisabled indicates the selected vendor is not import-enabled.
	ErrVendorImportDisabled = fmt.Errorf("vendor import disabled")
	// ErrVendorNotBrokerage indicates a portfolio import was requested for a non-brokerage vendor.
	ErrVendorNotBrokerage = fmt.Errorf("vendor must be a brokerage vendor for portfolio imports")
	// ErrImportListingNotFound indicates the requested listing does not exist.
	ErrImportListingNotFound = fmt.Errorf("import listing not found")
	// ErrImportListingInactive indicates the requested listing cannot receive imports.
	ErrImportListingInactive = fmt.Errorf("import listing is inactive")
	// ErrImportProviderUnavailable indicates an EOD import has no usable provider.
	ErrImportProviderUnavailable = fmt.Errorf("import provider unavailable")
	// ErrImportProviderNotManual indicates an EOD import source is not configured for manual ingestion.
	ErrImportProviderNotManual = fmt.Errorf("import provider is not manual")
	// ErrImportProcessorUnavailable indicates no processor is configured for the import type.
	ErrImportProcessorUnavailable = fmt.Errorf("import processor unavailable")
)

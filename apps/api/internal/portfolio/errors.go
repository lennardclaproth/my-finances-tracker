package portfolio

import "fmt"

var (
	ErrManualAccountNotFound         = fmt.Errorf("account not found")
	ErrManualVendorNotFound          = fmt.Errorf("vendor not found")
	ErrManualListingNotFound         = fmt.Errorf("listing not found")
	ErrManualVendorNotActive         = fmt.Errorf("vendor must be active")
	ErrManualVendorTypeNotSupported  = fmt.Errorf("vendor type not allowed for portfolio manual transaction")
	ErrManualInvalidOccurredAt       = fmt.Errorf("occurred_at must be in YYYY-MM-DD format")
	ErrManualInvalidType             = fmt.Errorf("type must be one of: BUY, SELL, DIVIDEND, TAX, FEE, CASH")
	ErrManualInvalidAmount           = fmt.Errorf("amount must be a decimal string with up to 6 decimals")
	ErrManualInvalidQuantity         = fmt.Errorf("quantity must be a decimal string with up to 6 decimals")
	ErrManualQuantityRequired        = fmt.Errorf("quantity is required for BUY/SELL")
	ErrManualQuantityForbidden       = fmt.Errorf("quantity is only allowed for BUY/SELL")
	ErrManualListingRequired         = fmt.Errorf("listing_id is required for non-CASH transactions")
	ErrManualListingForbidden        = fmt.Errorf("listing_id is not allowed for CASH transactions")
	ErrManualNonCashAmountMustBePos  = fmt.Errorf("amount must be positive for non-CASH transactions")
	ErrManualCashAmountMustBeNonZero = fmt.Errorf("amount must be non-zero for CASH transactions")
	ErrManualQuantityMustBePositive  = fmt.Errorf("quantity must be positive")
	ErrManualListingIdentityMissing  = fmt.Errorf("listing must provide at least one of isin or symbol")
)

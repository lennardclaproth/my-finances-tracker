package assets

import (
	"testing"

	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
)

// TestAdjustAssetWorthRequestIsValid verifies that direction validation accepts
// the two valid change directions and rejects anything else. Amount and
// effective date are held valid so each case isolates the direction check.
func TestAdjustAssetWorthRequestIsValid(t *testing.T) {
	tests := []struct {
		name         string
		direction    string
		wantValid    bool
		wantDirIssue bool
	}{
		{
			name:      "valid increase direction",
			direction: string(assets.ChangeDirectionIncrease),
			wantValid: true,
		},
		{
			name:      "valid decrease direction",
			direction: string(assets.ChangeDirectionDecrease),
			wantValid: true,
		},
		{
			name:         "invalid direction",
			direction:    "SIDEWAYS",
			wantValid:    false,
			wantDirIssue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := AdjustAssetWorthRequest{
				Direction:     tt.direction,
				Amount:        "100.00",
				EffectiveDate: "2024-01-15",
			}

			gotValid, problems := req.isValid()

			if gotValid != tt.wantValid {
				t.Errorf("isValid() valid = %v, want %v (problems: %v)", gotValid, tt.wantValid, problems)
			}

			if _, hasDirIssue := problems["direction"]; hasDirIssue != tt.wantDirIssue {
				t.Errorf("isValid() direction problem present = %v, want %v (problems: %v)", hasDirIssue, tt.wantDirIssue, problems)
			}
		})
	}
}

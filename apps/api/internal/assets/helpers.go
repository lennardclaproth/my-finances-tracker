package assets

import (
	"math"
	"sort"
	"time"

	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

func growthPctFromBouds(bound ClassBounds) float64 {
	if bound.First != nil && bound.Last != nil {
		if pct := growthPctFromInception(bound.First.ClassTotalWorth, bound.Last.ClassTotalWorth); pct != nil {
			return *pct
		}
	}
	return 0.0
}

func growthPctFromInception(inceptionWorth, latestWorth money.Price) *float64 {
	inception := inceptionWorth.Float64()
	if inception == 0 {
		return nil
	}
	latest := latestWorth.Float64()
	value := ((latest - inception) / math.Abs(inception)) * 100
	return &value
}

func growthPointsFromMutations(mutations []Mutation) []GrowthPoint {
	type dailyPoint struct {
		date          time.Time
		effectiveDate time.Time
		totalWorth    money.Price
	}

	byDate := make(map[string]dailyPoint, len(mutations))

	for _, mutation := range mutations {
		date := date.StartOfDayUTC(mutation.EffectiveDate)
		key := date.Format(time.DateOnly)

		existing, exists := byDate[key]
		if exists && !mutation.EffectiveDate.After(existing.effectiveDate) {
			continue
		}

		byDate[key] = dailyPoint{
			date:          date,
			effectiveDate: mutation.EffectiveDate,
			totalWorth:    mutation.ClassTotalWorth,
		}
	}

	keys := make([]string, 0, len(byDate))
	for key := range byDate {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	points := make([]GrowthPoint, 0, len(keys))
	for _, key := range keys {
		point := byDate[key]

		points = append(points, GrowthPoint{
			Date:       point.date,
			TotalWorth: point.totalWorth,
		})
	}

	return points
}

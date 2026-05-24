package sorting

import (
	"fmt"
	"strings"
)

type Direction string
type Field string

type Sort struct {
	Field     Field
	Direction Direction
}

const (
	ASC  Direction = "ASC"
	DESC Direction = "DESC"
)

func Parse(s string) (Direction, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "", "ASC":
		return ASC, nil
	case "DESC":
		return DESC, nil
	default:
		return "", fmt.Errorf("invalid sort direction %q", s)
	}
}

// FieldParser validates and maps a raw sort field to a feature-supported field.
type FieldParser func(string) (Field, error)

// ParseSort constructs a sort value from raw field and direction values.
//
// Valid sortable fields and the default direction are supplied by the calling
// feature because sorting rules vary by query use case.
func ParseSort(
	fieldRaw string,
	directionRaw string,
	defaultDirection Direction,
	parseField FieldParser,
) (Sort, error) {
	field, err := parseField(fieldRaw)
	if err != nil {
		return Sort{}, fmt.Errorf("parse sort field: %w", err)
	}

	direction := defaultDirection
	if strings.TrimSpace(directionRaw) != "" {
		direction, err = Parse(directionRaw)
		if err != nil {
			return Sort{}, fmt.Errorf("parse sort direction: %w", err)
		}
	}

	return Sort{
		Field:     field,
		Direction: direction,
	}, nil
}

func MustParse(s string) Direction {
	d, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return d
}

func (d Direction) String() string {
	return string(d)
}

func (d Direction) IsValid() bool {
	return d == ASC || d == DESC
}

func (d Direction) Reverse() Direction {
	if d == ASC {
		return DESC
	}
	return ASC
}

func (d Direction) SQL() string {
	if d == DESC {
		return "DESC"
	}
	return "ASC"
}

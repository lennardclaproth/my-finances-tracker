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

package sorting

import "encoding/json"

func (d *Direction) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, err := Parse(s)
	if err != nil {
		return err
	}

	*d = parsed
	return nil
}

func (d Direction) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

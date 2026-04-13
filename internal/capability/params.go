package capability

import (
	"encoding/json"
	"fmt"
)

// Validator can be implemented by params structs to provide field-level validation.
type Validator interface {
	Validate() error
}

// ParseParams unmarshals raw JSON into T and calls Validate() if T implements Validator.
func ParseParams[T any](raw json.RawMessage) (T, error) {
	var p T
	if err := json.Unmarshal(raw, &p); err != nil {
		return p, fmt.Errorf("invalid params: %w", err)
	}
	if v, ok := any(&p).(Validator); ok {
		if err := v.Validate(); err != nil {
			return p, err
		}
	}
	return p, nil
}

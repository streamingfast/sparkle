package entity

import (
	"database/sql/driver"
)

type Bytes []byte

func (b Bytes) Value() (driver.Value, error) {
	if len(b) == 0 {
		return nil, nil
	}
	return []byte(b), nil
}

func (b *Bytes) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	*b = value.([]byte)
	return nil
}

package typ

import "database/sql/driver"

type Color string

func (c *Color) SetDefault() {
	if c == nil || *c == "" {
		*c = Color("#000000")
	}
}

func (c Color) Value() (driver.Value, error) {
	if c == "" {
		return "#000000", nil
	}

	return string(c), nil
}

func (c *Color) Scan(value interface{}) error {
	if value == nil {
		*c = Color("#000000")
		return nil
	}
	switch v := value.(type) {
	case string:
		*c = Color(v)
	case []byte:
		*c = Color(string(v))
	}
	return nil
}

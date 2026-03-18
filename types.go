package mosaic

import (
	"net/url"
	"strconv"
	"time"
)

// ListParams contains pagination parameters for list operations.
// The Mosaic API uses limit/offset pagination with a default limit of 30.
type ListParams struct {
	// Limit is the maximum number of results to return (default: 30).
	Limit int `json:"limit,omitempty"`

	// Offset is the number of results to skip (default: 0).
	Offset int `json:"offset,omitempty"`

	// All fetches all records if set to true, ignoring limit and offset.
	All *bool `json:"all,omitempty"`

	// SortAttributes specifies the fields to sort by.
	SortAttributes []string `json:"sort_attributes,omitempty"`
}

// toQuery converts ListParams into URL query parameter values.
func (p ListParams) toQuery() url.Values {
	q := make(url.Values)
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset > 0 {
		q.Set("offset", strconv.Itoa(p.Offset))
	}
	if p.All != nil && *p.All {
		q.Set("all", "true")
	}
	for _, attr := range p.SortAttributes {
		q.Add("sort_attributes[]", attr)
	}
	return q
}

// Date represents a date without time, serialized as "MM/DD/YYYY" for the
// Mosaic API. The query parameter format used by Mosaic is MM/DD/YYYY.
type Date struct {
	time.Time
}

// dateFormat is the layout used for Date JSON serialization (MM/DD/YYYY).
const dateFormat = "01/02/2006"

// MarshalJSON implements json.Marshaler for Date.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.Format(dateFormat) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler for Date.
func (d *Date) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		d.Time = time.Time{}
		return nil
	}
	// Strip quotes.
	s = s[1 : len(s)-1]

	// Try MM/DD/YYYY first, then fall back to YYYY-MM-DD for response parsing.
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		t, err = time.Parse("2006-01-02", s)
		if err != nil {
			return err
		}
	}
	d.Time = t
	return nil
}

// String returns the date formatted as "MM/DD/YYYY".
func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return d.Format(dateFormat)
}

// NewDate creates a Date from year, month, and day.
func NewDate(year int, month time.Month, day int) Date {
	return Date{Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

// ParseDate parses a date string in MM/DD/YYYY or YYYY-MM-DD format.
func ParseDate(s string) (Date, error) {
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		t, err = time.Parse("2006-01-02", s)
		if err != nil {
			return Date{}, err
		}
	}
	return Date{Time: t}, nil
}

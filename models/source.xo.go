// Package models contains the types for schema 'alpha'.
package models

// GENERATED BY XO. DO NOT EDIT.

import (
	"database/sql/driver"
	"errors"
)

// Source is the 'source' enum type from schema 'alpha'.
type Source uint16

const (
	// SourceNone is the 'NONE' Source.
	SourceNone = Source(1)

	// SourceImgur is the 'IMGUR' Source.
	SourceImgur = Source(2)
)

// String returns the string value of the Source.
func (s Source) String() string {
	var enumVal string

	switch s {
	case SourceNone:
		enumVal = "NONE"

	case SourceImgur:
		enumVal = "IMGUR"
	}

	return enumVal
}

// MarshalText marshals Source into text.
func (s Source) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText unmarshals Source from text.
func (s *Source) UnmarshalText(text []byte) error {
	switch string(text) {
	case "NONE":
		*s = SourceNone

	case "IMGUR":
		*s = SourceImgur

	default:
		return errors.New("invalid Source")
	}

	return nil
}

// Value satisfies the sql/driver.Valuer interface for Source.
func (s Source) Value() (driver.Value, error) {
	return s.String(), nil
}

// Scan satisfies the database/sql.Scanner interface for Source.
func (s *Source) Scan(src interface{}) error {
	buf, ok := src.([]byte)
	if !ok {
		return errors.New("invalid Source")
	}

	return s.UnmarshalText(buf)
}

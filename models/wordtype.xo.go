// Package models contains the types for schema 'alpha'.
package models

// GENERATED BY XO. DO NOT EDIT.

import (
	"database/sql/driver"
	"errors"
)

// WordType is the 'word_type' enum type from schema 'alpha'.
type WordType uint16

const (
	// WordTypeNone is the 'NONE' WordType.
	WordTypeNone = WordType(1)

	// WordTypeCaption is the 'CAPTION' WordType.
	WordTypeCaption = WordType(2)

	// WordTypeMemeText is the 'MEME_TEXT' WordType.
	WordTypeMemeText = WordType(3)

	// WordTypeSpellcheck is the 'SPELLCHECK' WordType.
	WordTypeSpellcheck = WordType(4)
)

// String returns the string value of the WordType.
func (wt WordType) String() string {
	var enumVal string

	switch wt {
	case WordTypeNone:
		enumVal = "NONE"

	case WordTypeCaption:
		enumVal = "CAPTION"

	case WordTypeMemeText:
		enumVal = "MEME_TEXT"

	case WordTypeSpellcheck:
		enumVal = "SPELLCHECK"
	}

	return enumVal
}

// MarshalText marshals WordType into text.
func (wt WordType) MarshalText() ([]byte, error) {
	return []byte(wt.String()), nil
}

// UnmarshalText unmarshals WordType from text.
func (wt *WordType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "NONE":
		*wt = WordTypeNone

	case "CAPTION":
		*wt = WordTypeCaption

	case "MEME_TEXT":
		*wt = WordTypeMemeText

	case "SPELLCHECK":
		*wt = WordTypeSpellcheck

	default:
		return errors.New("invalid WordType")
	}

	return nil
}

// Value satisfies the sql/driver.Valuer interface for WordType.
func (wt WordType) Value() (driver.Value, error) {
	return wt.String(), nil
}

// Scan satisfies the database/sql.Scanner interface for WordType.
func (wt *WordType) Scan(src interface{}) error {
	buf, ok := src.([]byte)
	if !ok {
		return errors.New("invalid WordType")
	}

	return wt.UnmarshalText(buf)
}

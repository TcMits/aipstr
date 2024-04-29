package aipstr

import (
	"time"
)

type Value interface {
	IntoIdent() ([]string, bool)
	IntoInt() (int64, bool)
	IntoFloat() (float64, bool)
	IntoBool() (bool, bool)
	IntoString() (string, bool)
	IntoDuration() (time.Duration, bool)
	IntoTime() (time.Time, bool)
}

var _ Value = (*unimplementedValue)(nil)

type unimplementedValue struct{}

func (unimplementedValue) IntoIdent() ([]string, bool) {
	return nil, false
}

func (unimplementedValue) IntoInt() (int64, bool) {
	return 0, false
}

func (unimplementedValue) IntoFloat() (float64, bool) {
	return 0, false
}

func (unimplementedValue) IntoBool() (bool, bool) {
	return false, false
}

func (unimplementedValue) IntoString() (string, bool) {
	return "", false
}

func (unimplementedValue) IntoDuration() (time.Duration, bool) {
	return 0, false
}

func (unimplementedValue) IntoTime() (time.Time, bool) {
	return time.Time{}, false
}

var _ Value = (*IdentValue)(nil)

type IdentValue struct {
	unimplementedValue

	Value []string
}

func (i IdentValue) IntoIdent() ([]string, bool) {
	return i.Value, true
}

var _ Value = (*IntValue)(nil)

type IntValue struct {
	unimplementedValue

	Value int64
}

func (i IntValue) IntoInt() (int64, bool) {
	return i.Value, true
}

var _ Value = (*FloatValue)(nil)

type FloatValue struct {
	unimplementedValue

	Value float64
}

func (f FloatValue) IntoFloat() (float64, bool) {
	return f.Value, true
}

var _ Value = (*BoolValue)(nil)

type BoolValue struct {
	unimplementedValue

	Value bool
}

func (b BoolValue) IntoBool() (bool, bool) {
	return b.Value, true
}

var _ Value = (*StringValue)(nil)

type StringValue struct {
	unimplementedValue

	Value string
}

func (s StringValue) IntoString() (string, bool) {
	return s.Value, true
}

func (s StringValue) IntoDuration() (time.Duration, bool) {
	if d, err := time.ParseDuration(s.Value); err == nil {
		return d, true
	}

	return 0, false
}

func (s StringValue) IntoTime() (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339Nano, s.Value); err == nil {
		return t, true
	}

	return time.Time{}, false
}

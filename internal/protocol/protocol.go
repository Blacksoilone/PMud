package protocol

import (
	"errors"
	"strings"
)

var (
	ErrMissingEventField = errors.New("missing event field")
	ErrMalformedField    = errors.New("malformed field")
	ErrInvalidEscape     = errors.New("invalid escape")
	ErrDuplicateField    = errors.New("duplicate field")
)

type Event struct {
	Name   string
	Fields map[string]string
}

func ParseLine(line string) (Event, error) {
	trimmed := strings.TrimSuffix(line, "\n")
	fields := strings.Split(trimmed, "\t")
	parsedFields := make(map[string]string, len(fields))

	for _, rawField := range fields {
		name, value, ok := strings.Cut(rawField, "=")
		if !ok || name == "" {
			return Event{}, ErrMalformedField
		}
		if _, exists := parsedFields[name]; exists {
			return Event{}, ErrDuplicateField
		}

		unescapedValue, err := unescapeValue(value)
		if err != nil {
			return Event{}, err
		}
		parsedFields[name] = unescapedValue
	}

	eventName, ok := parsedFields["event"]
	if !ok || eventName == "" {
		return Event{}, ErrMissingEventField
	}
	delete(parsedFields, "event")

	return Event{
		Name:   eventName,
		Fields: parsedFields,
	}, nil
}

func unescapeValue(value string) (string, error) {
	var builder strings.Builder
	builder.Grow(len(value))
	for index := 0; index < len(value); index++ {
		char := value[index]
		if char != '\\' {
			builder.WriteByte(char)
			continue
		}

		index++
		if index >= len(value) {
			return "", ErrInvalidEscape
		}
		switch value[index] {
		case 'n':
			builder.WriteByte('\n')
		case 't':
			builder.WriteByte('\t')
		case '\\':
			builder.WriteByte('\\')
		default:
			return "", ErrInvalidEscape
		}
	}

	return builder.String(), nil
}

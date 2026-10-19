package forgeapi

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		metadata modelMetadata
		expected any
		err      error
	}{
		{
			name: "string identifier",
			id:   "sku-123",
			metadata: modelMetadata{
				primaryKeyType: reflect.TypeOf(""),
			},
			expected: "sku-123",
			err:      nil,
		},
		{
			name: "int identifier",
			id:   "123",
			metadata: modelMetadata{
				primaryKeyType: reflect.TypeOf(int(123)),
			},
			expected: 123,
			err:      nil,
		},
		{
			name: "invalid int identifier",
			id:   "sku 123",
			metadata: modelMetadata{
				primaryKeyType: reflect.TypeOf(int(123)),
			},
			err: ErrInvalidIdentifier,
		},
		{
			name: "int8 identifier out of range",
			id:   "128",
			metadata: modelMetadata{
				primaryKeyType: reflect.TypeOf(int8(0)),
			},
			err: ErrInvalidIdentifier,
		},
		{
			name: "uint identifier",
			id:   "123",
			metadata: modelMetadata{
				primaryKeyType: reflect.TypeOf(uint(123)),
			},
			expected: uint(123),
			err:      nil,
		},
		{
			name: "invalid identifier type",
			id:   "123.00",
			metadata: modelMetadata{
				primaryKeyType: reflect.TypeOf(float64(123.00)),
			},
			err: ErrUnsupportedIdentifierType,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseIdentifier(test.id, test.metadata.primaryKeyType)

			if !errors.Is(err, test.err) {
				t.Fatalf("parseIdentifier() error = %v, want %v", err, test.err)
			}

			if !reflect.DeepEqual(got, test.expected) {
				t.Fatalf(
					"parseIdentifier() = %T(%v), want %T(%v)",
					got, got, test.expected, test.expected,
				)
			}
		})
	}
}

func TestValidateIdentifierType(t *testing.T) {
	tests := []struct {
		name    string
		idType  reflect.Type
		wantErr error
	}{
		{
			name:   "string identifier",
			idType: reflect.TypeOf(""),
		},
		{
			name:   "uint identifier",
			idType: reflect.TypeOf(uint(0)),
		},
		{
			name:   "int64 identifier",
			idType: reflect.TypeOf(int64(0)),
		},
		{
			name:    "float64 identifier",
			idType:  reflect.TypeOf(float64(0)),
			wantErr: ErrUnsupportedIdentifierType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIdentifierType(tt.idType)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"validateIdentifierType() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

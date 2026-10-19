package forgeapi

import (
	"fmt"
	"reflect"
	"strconv"
)

func parseIdentifier(id string, idType reflect.Type) (any, error) {
	switch idType.Kind() {
	case reflect.String:
		return id, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		value, err := strconv.ParseUint(id, 10, idType.Bits())
		if err != nil {
			return nil, fmt.Errorf("%w: %q: %v", ErrInvalidIdentifier, id, err)
		}

		return reflect.ValueOf(value).Convert(idType).Interface(), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(id, 10, idType.Bits())
		if err != nil {
			return nil, fmt.Errorf("%w: %q: %v", ErrInvalidIdentifier, id, err)
		}

		return reflect.ValueOf(value).Convert(idType).Interface(), nil

	default:
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedIdentifierType, idType)
	}
}

func validateIdentifierType(idType reflect.Type) error {
	switch idType.Kind() {
	case reflect.String,
		reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return nil
	default:
		return ErrUnsupportedIdentifierType
	}
}

package forgeapi

import (
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestResolveModelMetadataInvalidModel(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	type InvalidModel struct {
		ID       uint `gorm:"primaryKey" forgeapi:"id"`
		Relation struct {
			Unknown string
		}
	}

	_, err = resolveModelMetadata(db, &InvalidModel{})

	if !errors.Is(err, ErrInvalidModelMetadata) {
		t.Fatalf(
			"expected error %v, got %v",
			ErrInvalidModelMetadata,
			err,
		)
	}
}

func TestResolveModelMetadataMissingID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	type ModelWithoutID struct {
		Name string
	}

	_, err = resolveModelMetadata(db, &ModelWithoutID{})

	if !errors.Is(err, ErrInvalidModelMetadata) {
		t.Fatalf(
			"expected error %v, got %v",
			ErrInvalidModelMetadata,
			err,
		)
	}
}

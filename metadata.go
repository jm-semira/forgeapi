package forgeapi

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type modelMetadata struct {
	primaryKeyColumn string
	primaryKeyType   reflect.Type
}

func resolveModelMetadata(db *gorm.DB, model any) (modelMetadata, error) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return modelMetadata{}, fmt.Errorf(
			"%w: %w",
			ErrInvalidModelMetadata,
			err,
		)
	}

	if len(stmt.Schema.PrimaryFields) > 1 {
		return modelMetadata{}, ErrInvalidModelMetadata
	}

	var taggedField *schema.Field

	for _, field := range stmt.Schema.Fields {
		if strings.Split(field.Tag.Get("forgeapi"), ",")[0] == "id" {
			if taggedField != nil {
				return modelMetadata{}, ErrInvalidModelMetadata
			}
			taggedField = field
		}
	}

	if taggedField == nil || !taggedField.PrimaryKey {
		return modelMetadata{}, ErrInvalidModelMetadata
	}

	return modelMetadata{
		primaryKeyColumn: taggedField.DBName,
		primaryKeyType:   taggedField.FieldType,
	}, nil
}

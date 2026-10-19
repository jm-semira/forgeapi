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
	fields           map[string]bool
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

	fields := make(map[string]bool, len(stmt.Schema.Fields))

	for _, field := range stmt.Schema.Fields {
		if field.PrimaryKey || field.DBName == "" {
			continue
		}

		if strings.Contains(
			","+field.Tag.Get("forgeapi")+",",
			",readonly,",
		) {
			continue
		}

		fields[field.Name] = true
	}

	return modelMetadata{
		primaryKeyColumn: taggedField.DBName,
		primaryKeyType:   taggedField.FieldType,
		fields:           fields,
	}, nil
}

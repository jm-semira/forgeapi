package forgeapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (app *API) Register(model any) error {
	if app.db == nil {
		return ErrDatabaseRequired
	}

	value := reflect.ValueOf(model)
	if !value.IsValid() ||
		value.Kind() != reflect.Ptr ||
		value.IsNil() ||
		value.Elem().Kind() != reflect.Struct {
		return ErrInvalidModel
	}

	modelType := value.Elem().Type()
	resource := pluralize(strings.ToLower(modelType.Name()))
	path := "/" + resource

	metadata, err := resolveModelMetadata(app.db, model)
	if err != nil {
		return err
	}

	if err := validateIdentifierType(metadata.primaryKeyType); err != nil {
		return err
	}

	app.mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
		records := reflect.New(reflect.SliceOf(modelType)).Interface()

		query := app.db.WithContext(r.Context())

		for key, values := range r.URL.Query() {
			if !strings.HasPrefix(key, "filter[") || !strings.HasSuffix(key, "]") {
				continue
			}

			parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(key, "filter["), "]"), "][")
			if len(parts) != 2 {
				http.Error(w, "invalid filter", http.StatusBadRequest)
				return
			}

			operator := parts[1]
			switch operator {
			case "eq", "gte", "lte":
			default:
				http.Error(w, "invalid filter", http.StatusBadRequest)
				return
			}

			fieldName := parts[0]
			fieldInfo, ok := metadata.fields[fieldName]
			if !ok {
				http.Error(w, "invalid filter field", http.StatusBadRequest)
				return
			}

			if len(values) != 1 {
				http.Error(w, "invalid filter value", http.StatusBadRequest)
				return
			}

			var condition string

			switch operator {
			case "eq":
				condition = fieldInfo.dbName + " = ?"
			case "gte":
				condition = fieldInfo.dbName + " >= ?"
			case "lte":
				condition = fieldInfo.dbName + "<= ?"
			}

			query = query.Where(condition, values[0])
		}

		if sortFields := r.URL.Query().Get("sort"); sortFields != "" {
			for _, sortField := range strings.Split(sortFields, ",") {
				direction := "ASC"

				if strings.HasPrefix(sortField, "-") {
					direction = "DESC"
					sortField = strings.TrimPrefix(sortField, "-")
				}

				fieldInfo, ok := metadata.fields[sortField]
				if !ok {
					http.Error(w, "invalid sort field", http.StatusBadRequest)
					return
				}

				query = query.Order(fieldInfo.dbName + " " + direction)
			}
		}
		limit := app.defaultLimit

		if limitField := r.URL.Query().Get("limit"); limitField != "" {
			parsedLimit, err := strconv.Atoi(limitField)
			if err != nil || parsedLimit <= 0 {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}

			if parsedLimit > app.maxLimit {
				http.Error(w, "limit exceeds maximum", http.StatusBadRequest)
				return
			}

			limit = parsedLimit
		}

		if limit <= 0 || limit > app.maxLimit {
			http.Error(w, "invalid pagination configuration", http.StatusInternalServerError)
			return
		}

		query = query.Limit(limit)
		if offsetField := r.URL.Query().Get("offset"); offsetField != "" {
			offset, err := strconv.Atoi(offsetField)
			if err != nil || offset < 0 {
				http.Error(w, "invalid offset", http.StatusBadRequest)
				return
			}

			query = query.Offset(offset)
		}

		result := query.Find(records)
		if result.Error != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(records); err != nil {
			return
		}
	})

	app.mux.HandleFunc("GET "+path+"/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIdentifier(
			r.PathValue("id"),
			metadata.primaryKeyType,
		)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		record := reflect.New(modelType).Interface()

		result := app.db.WithContext(r.Context()).
			Where(clause.Eq{
				Column: clause.Column{Name: metadata.primaryKeyColumn},
				Value:  id,
			}).First(record)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				http.Error(w, "record not found", http.StatusNotFound)
				return
			}

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(record); err != nil {
			return
		}
	})

	app.mux.HandleFunc("POST "+path, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Decode one JSON object and reject trailing JSON.
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()

		var fields map[string]json.RawMessage
		if err := decoder.Decode(&fields); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if len(fields) == 0 {
			http.Error(w, "empty request body", http.StatusBadRequest)
			return
		}

		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Validate supplied field names against the writable-field allowlist.
		for field := range fields {
			if _, ok := metadata.fields[field]; !ok {
				http.Error(w, "unknown or read-only field", http.StatusBadRequest)
				return
			}
		}

		// Decode into the registered model type.
		record := reflect.New(modelType).Interface()
		if err := json.Unmarshal(body, record); err != nil {
			http.Error(w, "invalid field value", http.StatusBadRequest)
			return
		}

		// Persist the new record.
		if err := app.db.WithContext(r.Context()).Create(record).Error; err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(record); err != nil {
			// The response status has already been written.
			return
		}
	})

	app.mux.HandleFunc("PATCH "+path+"/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIdentifier(
			r.PathValue("id"),
			metadata.primaryKeyType,
		)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		var updates map[string]json.RawMessage
		model := reflect.New(modelType).Interface()

		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(&updates); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		if len(updates) == 0 {
			http.Error(w, "empty update body", http.StatusBadRequest)
			return
		}

		typedUpdates := make(map[string]any, len(updates))

		for field, raw := range updates {
			fieldInfo, ok := metadata.fields[field]
			if !ok {
				http.Error(w, "unknown or read-only update field", http.StatusBadRequest)
				return
			}
			fieldType := fieldInfo.typeOf

			raw = bytes.TrimSpace(raw)

			if bytes.Equal(raw, []byte("null")) {
				switch fieldType.Kind() {
				case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice:
					typedUpdates[fieldInfo.dbName] = reflect.Zero(fieldType).Interface()
					continue
				default:
					http.Error(w, "invalid field value", http.StatusBadRequest)
					return
				}
			}

			value := reflect.New(fieldType)

			if err := json.Unmarshal(raw, value.Interface()); err != nil {
				http.Error(w, "invalid field value", http.StatusBadRequest)
				return
			}

			typedUpdates[fieldInfo.dbName] = value.Elem().Interface()
		}

		// check if record is existing
		record := reflect.New(modelType).Interface()

		if err := app.db.WithContext(r.Context()).
			Where(clause.Eq{
				Column: clause.Column{Name: metadata.primaryKeyColumn},
				Value:  id,
			}).First(record).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "record not found", http.StatusNotFound)
				return
			}

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return

		}

		result := app.db.WithContext(r.Context()).
			Model(model).
			Where(clause.Eq{
				Column: clause.Column{Name: metadata.primaryKeyColumn},
				Value:  id,
			}).
			Updates(typedUpdates)
		if result.Error != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		updated := reflect.New(modelType).Interface()

		if err := app.db.WithContext(r.Context()).
			Where(clause.Eq{
				Column: clause.Column{Name: metadata.primaryKeyColumn},
				Value:  id,
			}).First(updated).Error; err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(updated); err != nil {
			return
		}
	})

	app.mux.HandleFunc("DELETE "+path+"/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIdentifier(
			r.PathValue("id"),
			metadata.primaryKeyType,
		)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		model := reflect.New(modelType).Interface()

		result := app.db.WithContext(r.Context()).
			Where(clause.Eq{
				Column: clause.Column{Name: metadata.primaryKeyColumn},
				Value:  id,
			}).
			Delete(model)

		if result.Error != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if result.RowsAffected == 0 {
			http.Error(w, "record not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	return nil
}

func pluralize(name string) string {
	switch {
	case strings.HasSuffix(name, "ch"),
		strings.HasSuffix(name, "sh"),
		strings.HasSuffix(name, "x"),
		strings.HasSuffix(name, "s"):
		return name + "es"
	case strings.HasSuffix(name, "y") && len(name) > 1:
		return name[:len(name)-1] + "ies"
	default:
		return fmt.Sprintf("%ss", name)
	}
}

var _ *gorm.DB

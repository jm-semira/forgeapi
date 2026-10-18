package forgeapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
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

	app.mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
		records := reflect.New(reflect.SliceOf(modelType)).Interface()

		result := app.db.WithContext(r.Context()).Find(records)
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
		id := r.PathValue("id")

		record := reflect.New(modelType).Interface()

		result := app.db.WithContext(r.Context()).First(record, id)
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
		record := reflect.New(modelType).Interface()

		if err := json.NewDecoder(r.Body).Decode(record); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		result := app.db.WithContext(r.Context()).Create(record)
		if result.Error != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(record); err != nil {
			return
		}
	})

	app.mux.HandleFunc("PATCH "+path+"/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		updates := map[string]any{}
		model := reflect.New(modelType).Interface()

		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		if len(updates) == 0 {
			http.Error(w, "empty update body", http.StatusBadRequest)
			return
		}

		// check if record is existing
		record := reflect.New(modelType).Interface()

		if err := app.db.WithContext(r.Context()).First(record, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "record not found", http.StatusNotFound)
				return
			}

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return

		}

		result := app.db.WithContext(r.Context()).
			Model(model).
			Where("id = ?", id).
			Updates(updates)
		if result.Error != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		updated := reflect.New(modelType).Interface()

		if err := app.db.WithContext(r.Context()).First(updated, id).Error; err != nil {
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
		idStr := r.PathValue("id")

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		model := reflect.New(modelType).Interface()

		result := app.db.WithContext(r.Context()).
			Where("id = ?", id).
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

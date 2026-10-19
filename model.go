package forgeapi

import (
	"time"
)

type Model struct {
	ID        uint      `json:"id" gorm:"primaryKey" forgeapi:"id"`
	CreatedAt time.Time `json:"created_at" forgeapi:"readonly"`
	UpdatedAt time.Time `json:"updated_at" forgeapi:"readonly"`
}

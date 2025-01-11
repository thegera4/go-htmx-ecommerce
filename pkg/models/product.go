package models

import (
	"time"
	"github.com/google/uuid"
)

// Custom type (model) that represents a Product from the database
type Product struct {
	ProductID    uuid.UUID
	ProductName  string
	Price        float64
	Description  string
	ProductImage string
	DateCreated  time.Time
	DateModified time.Time
}
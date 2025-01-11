package models

import (
	"time"
	"github.com/google/uuid"
)

// Custom type (model) that represents an Order from the database
type Order struct {
	OrderID     uuid.UUID
	UserID      string
	OrderStatus string
	OrderDate   time.Time
	Items       []OrderItem
}
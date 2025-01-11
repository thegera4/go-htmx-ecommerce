package models

import "github.com/google/uuid"

// Custom type (model) that represents an order item (product in an order) from the database
type OrderItem struct {
	OrderID   uuid.UUID
	ProductID uuid.UUID
	Quantity  int
	Product   Product
	Cost      float64
}
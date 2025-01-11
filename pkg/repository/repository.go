package repository

import "database/sql"

// Custom type that contains pointers to the ProductRepository and OrderRepository.
type Repository struct {
	Product *ProductRepository
	Order   *OrderRepository
}

// Function that returns a new Repository with a pointer to the database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Product: NewProductRepository(db),
		Order:   NewOrderRepository(db),
	}
}
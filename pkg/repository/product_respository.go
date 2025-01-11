package repository

import (
	"database/sql"
	"time"
	"github.com/thegera4/go-htmx-ecommerce/pkg/models"
	"github.com/google/uuid"
)

// Custom type that holds a pointer to the database connection.
type ProductRepository struct {
	DB *sql.DB
}

// Function that returns a new ProductRepository (pointer) with the database connection.
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

// Function that returns a product by its ID from the database.
func (r *ProductRepository) GetProductByID(productID uuid.UUID) (*models.Product, error) {
	query := `SELECT product_id, product_name, price, description, product_image, date_created, date_modified FROM products WHERE product_id = ?`
	row := r.DB.QueryRow(query, productID)
	var product models.Product
	err := row.Scan(&product.ProductID, &product.ProductName, &product.Price, &product.Description, &product.ProductImage, &product.DateCreated, &product.DateModified)
	if err != nil { return nil, err }
	return &product, nil
}

// Function that creates a new product in the database.
func (r *ProductRepository) CreateProduct(product *models.Product) error {
	query := `INSERT INTO products (product_id, product_name, price, description, product_image, date_created, date_modified) VALUES (?, ?, ?, ?, ?, ?, ?)`
	product.ProductID = uuid.New()
	product.DateCreated = time.Now()
	product.DateModified = time.Now()
	_, err := r.DB.Exec(query, product.ProductID, product.ProductName, product.Price, product.Description, product.ProductImage, product.DateCreated, product.DateModified)
	return err
}

// Function that updates a product in the database.
func (r *ProductRepository) UpdateProduct(product *models.Product) error {
	query := `UPDATE products SET product_name = ?, price = ?, description = ?, date_modified = ? WHERE product_id = ?`
	product.DateModified = time.Now()
	_, err := r.DB.Exec(query, product.ProductName, product.Price, product.Description, product.DateModified, product.ProductID)
	return err
}

// Function that deletes a product from the database.
func (r *ProductRepository) DeleteProduct(productID uuid.UUID) error {
	query := `DELETE FROM products WHERE product_id = ?`
	_, err := r.DB.Exec(query, productID)
	return err
}

// Function that returns a list of products from the database. It takes a limit and offset as parameters in order to paginate the results.
func (r *ProductRepository) ListProducts(limit, offset int) ([]models.Product, error) {
	query := `SELECT product_id, product_name, price, description, product_image, date_created, date_modified FROM products ORDER BY date_created DESC LIMIT ? OFFSET ?`
	rows, err := r.DB.Query(query, limit, offset)
	if err != nil { return nil, err }
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(&product.ProductID, &product.ProductName, &product.Price, &product.Description, &product.ProductImage, &product.DateCreated, &product.DateModified)
		if err != nil { return nil, err }
		products = append(products, product)
	}
	return products, nil
}

// Function that returns the total number of products that exist in the database.
func (r *ProductRepository) GetTotalProductsCount() (int, error) {
	var count int
	err := r.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil { return 0, err }
	return count, nil
}

// Function that returns a list of products from the database based on a where clause.
func (r *ProductRepository) GetProducts(whereClause string) ([]models.Product, error) {
	query := `SELECT product_id, product_name, price, description, product_image, date_created, date_modified FROM products`
	if whereClause != "" { query += " WHERE " + whereClause }
	query += " ORDER BY date_created DESC"
	rows, err := r.DB.Query(query)
	if err != nil { return nil, err }
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ProductID, &p.ProductName, &p.Price, &p.Description, &p.ProductImage, &p.DateCreated, &p.DateModified)
		if err != nil { return nil, err }
		products = append(products, p)
	}
	if err = rows.Err(); err != nil { return nil, err }
	return products, nil
}
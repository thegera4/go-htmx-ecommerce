package main

import (
	"database/sql"
	"log"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/thegera4/go-htmx-ecommerce/pkg/handlers"
	"github.com/thegera4/go-htmx-ecommerce/pkg/repository"
)

/*** Global Variables ***/
var db *sql.DB // Database instance

// Initialize the database
func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:toor@(127.0.0.1:3306)/ecommerce?parseTime=true")
	if err != nil { log.Fatal(err)}

	if err = db.Ping(); err != nil { log.Fatal(err)}
}

func main() {
	r := mux.NewRouter()

	//Setup MySQL
	initDB()
	defer db.Close()

	// Setup Static folder for static files and images
	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	repo := repository.NewRepository(db)
	handler := handlers.NewHandler(repo)

	/*** User Routes ***/

	// Endpoint to display the home page (user shop view)
	r.HandleFunc("/", handler.ShoppingHomepage).Methods("GET")
	// Endpoint to display the products in the home page
	r.HandleFunc("/shoppingitems", handler.ShoppingItemsView).Methods("GET")
	// Endpoint to display the cart view in the home page
	r.HandleFunc("/cartitems", handler.CartView).Methods("GET")
	// Endpoint to add a product to the cart
	r.HandleFunc("/addtocart/{product_id}", handler.AddToCart).Methods("POST")
	// Endpoint to display the checkout view
	r.HandleFunc("/gotocart", handler.ShoppingCartView).Methods("GET")

	/*** Admin Routes ***/
	
	// Endpoint to seed (feed/create) 20 dummy products in the database (each time the endpoint is called)
	r.HandleFunc("/seed-products", handler.SeedProducts).Methods("POST")
	// Endpoint to display the products page
	r.HandleFunc("/manageproducts", handler.ProductsPage).Methods("GET")
	// Endpoint to display the all products view (table with all products)
	r.HandleFunc("/allproducts", handler.AllProductsView).Methods("GET")
	// Endpoint to display the rows of the all products view (table with all products)
	r.HandleFunc("/products", handler.ListProducts).Methods("GET")
	// Endpoint to display the details of a product
	r.HandleFunc("/products/{id}", handler.GetProduct).Methods("GET")
	// Endpoint to create a new product in the database
	r.HandleFunc("/products", handler.CreateProduct).Methods("POST")
	// Endpoint to update a product in the database
	r.HandleFunc("/products/{id}", handler.UpdateProduct).Methods("PUT")
	// Endpoint to delete a product from the database
	r.HandleFunc("/products/{id}", handler.DeleteProduct).Methods("DELETE")
	// Endpoint to display the form to add a new product
	r.HandleFunc("/createproduct", handler.CreateProductView).Methods("GET")
	// Endpoint to display the form to edit a product
	r.HandleFunc("/editproduct/{id}", handler.EditProductView).Methods("GET")

	http.ListenAndServe(":8080", r)
}
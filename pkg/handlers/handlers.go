package handlers

import (
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bxcodec/faker/v4"
	"github.com/thegera4/go-htmx-ecommerce/pkg/models"
	"github.com/thegera4/go-htmx-ecommerce/pkg/repository"
)

var tmpl *template.Template

// Function that initializes the templates.
func init() {
	templatesDir := "./templates"
	pattern := filepath.Join(templatesDir, "**", "*.html")
	tmpl = template.Must(template.ParseGlob(pattern))
}

// Custom type that contains a pointer to the Repositories.
type Handler struct {
	Repo *repository.Repository
}

// Function that returns a new Handler with a pointer to the Repository.
func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{Repo: repo}
}

// Utility function that subtracts two integers.
func makeRange(min, max int) []int {
	rangeArray := make([]int, max-min+1)
	for i := range rangeArray {
		rangeArray[i] = min + i
	}
	return rangeArray
}

// Function that seeds (feeds / creates) dummy products in the database.
func (h *Handler) SeedProducts(w http.ResponseWriter, r *http.Request) {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Number of products to generate
	numProducts := 20

	//An array of realistic product names to pick from
	productTypes := []string{"Laptop", "Smartphone", "Tablet", "Headphones", "Speaker", "Camera", "TV", "Watch", "Printer", "Monitor"}

	for i := 0; i < numProducts; i++ {
		//Generate the random but more realistic product type
		productType := productTypes[rand.Intn(len(productTypes))]
		productName := strings.Title(faker.Word()) + " " + productType

		product := models.Product{
			ProductName:  productName,
			Price:        float64(rand.Intn(100000)) / 100, // Random price between 0.00 and 999.99
			Description:  faker.Sentence(),
			ProductImage: faker.Word() + ".jpg",
		}

		err := h.Repo.Product.CreateProduct(&product)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating product %s: %v", product.ProductName, err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Successfully seeded %d dummy products", numProducts)
}

// Function that renders the products page.
func (h *Handler) ProductsPage(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "products", nil)
}

// Function that renders the all products view (table).
func (h *Handler) AllProductsView(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "allproducts", nil)
}

// Function that lists the products in the database in a paginated way.
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 { page = 1 }

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 { limit = 10 } // Default limit

	offset := (page - 1) * limit

	products, err := h.Repo.Product.ListProducts(limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalProducts, err := h.Repo.Product.GetTotalProductsCount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(totalProducts) / float64(limit)))
	previousPage := page - 1
	nextPage := page + 1
	pageButtonsRange := makeRange(1, totalPages)

	// Data to be passed to the template
	data := struct {
		Products         []models.Product
		CurrentPage      int
		TotalPages       int
		Limit            int
		PreviousPage     int
		NextPage         int
		PageButtonsRange []int
	}{
		Products:         products,
		CurrentPage:      page,
		TotalPages:       totalPages,
		Limit:            limit,
		PreviousPage:     previousPage,
		NextPage:         nextPage,
		PageButtonsRange: pageButtonsRange,
	}

	/*
		funcMap := template.FuncMap{
			"subtract":  subtract,
			"add":       add,
			"makeRange": makeRange,
		}

		productsTemplate := template.Must(template.New("productRows.html").Funcs(funcMap).ParseFiles("templates/admin/productRows.html"))

		err = productsTemplate.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	*/

	//Fake Latency
	//time.Sleep(5 * time.Second)

	tmpl.ExecuteTemplate(w, "productRows", data)
}
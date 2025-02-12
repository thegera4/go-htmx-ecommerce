package handlers

import (
	"fmt"
	"html/template"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bxcodec/faker/v4"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/thegera4/go-htmx-ecommerce/pkg/models"
	"github.com/thegera4/go-htmx-ecommerce/pkg/repository"
)

/*** Global Variables ***/
var tmpl *template.Template // Template variable
var currentCartOrderId uuid.UUID // Shopping cart order ID which is reused for the current session
var cartItems []models.OrderItem // Shopping cart items

/*** Structs ***/

// Custom type that contains the data to be passed to the template.
type ProductCRUDTemplateData struct {
	Messages []string
	Product  *models.Product
}

// Custom type that contains a pointer to the Repositories.
type Handler struct {
	Repo *repository.Repository
}

// Initializes the templates.
func init() {
	templatesDir := "./templates"
	pattern := filepath.Join(templatesDir, "**", "*.html")
	tmpl = template.Must(template.ParseGlob(pattern))
}

/*** Helper Functions	***/

// Sends messages to the user (for error or status).
func sendProductMessage(w http.ResponseWriter, messages []string, product *models.Product) {
	data := ProductCRUDTemplateData{Messages: messages, Product: product}
	tmpl.ExecuteTemplate(w, "messages", data)
}

// Returns a new Handler with a pointer to the Repository.
func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{Repo: repo}
}

// Subtracts two integers.
func makeRange(min, max int) []int {
	rangeArray := make([]int, max-min+1)
	for i := range rangeArray {
		rangeArray[i] = min + i
	}
	return rangeArray
}

// Calculates the total cost of the items in the cart.
func getTotalCartCost() float64 {
	totalCost := 0.0
	for _, item := range cartItems {
		totalCost += float64(item.Quantity) * item.Product.Price
	}
	return math.Round(totalCost * 100) / 100 // Round to 2 decimal places
}

/*** Handlers ***/

// Seeds (feeds / creates) dummy products in the database.
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

// Renders the products page.
func (h *Handler) ProductsPage(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "products", nil)
}

// Renders the all products view (table).
func (h *Handler) AllProductsView(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "allProducts", nil)
}

// Lists the products in the database in a paginated way.
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

// Renders the product detail page.
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.Repo.Product.GetProductByID(productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "viewProduct", product)
}

// Renders the create product page.
func (h *Handler) CreateProductView(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "createProduct", nil)
}

// Creates a new product in the database.
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {

	// Parse the multipart form, 10 MB max upload size
	r.ParseMultipartForm(10 << 20)

	// Initialize error messages slice
	var responseMessages []string

	//Check for empty fields
	ProductName := r.FormValue("product_name")
	ProductPrice := r.FormValue("price")
	ProductDescription := r.FormValue("description")

	if ProductName == "" || ProductPrice == "" || ProductDescription == "" {
		responseMessages = append(responseMessages, "All Fields Are Required")
		sendProductMessage(w, responseMessages, nil)
		return
	}

	/* Process File Upload */

	// Retrieve the file from form data
	file, handler, err := r.FormFile("product_image")
	if err != nil {
		if err == http.ErrMissingFile {
			responseMessages = append(responseMessages, "Select an Image for the Product")
		} else {
			responseMessages = append(responseMessages, "Error retrieving the file")
		}

		if len(responseMessages) > 0 {
			fmt.Println(responseMessages)
			sendProductMessage(w, responseMessages, nil)
			return
		}
	}
	defer file.Close()

	// Generate a unique filename to prevent overwriting and conflicts
	uuid, err := uuid.NewRandom()
	if err != nil {
		responseMessages = append(responseMessages, "Error generating unique identifier")
		sendProductMessage(w, responseMessages, nil)
		return
	}
	filename := uuid.String() + filepath.Ext(handler.Filename) // Append the file extension

	// Create the full path for saving the file
	filePath := filepath.Join("static/uploads", filename)

	// Save the file to the server
	dst, err := os.Create(filePath)
	if err != nil {
		responseMessages = append(responseMessages, "Error saving the file")
		sendProductMessage(w, responseMessages, nil)
		return
	}
	defer dst.Close()
	if _, err = io.Copy(dst, file); err != nil {
		responseMessages = append(responseMessages, "Error saving the file")
		sendProductMessage(w, responseMessages, nil)
		return
	}

	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		responseMessages = append(responseMessages, "Invalid price")
		sendProductMessage(w, responseMessages, nil)
		return
	}

	product := models.Product{
		ProductName:  ProductName,
		Price:        price,
		Description:  ProductDescription,
		ProductImage: filename,
	}

	err = h.Repo.Product.CreateProduct(&product)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		responseMessages = append(responseMessages, "Invalid price" + err.Error())
		sendProductMessage(w, responseMessages, nil)
		return
	}

	//Fake Latency
	time.Sleep(2 * time.Second)

	sendProductMessage(w, []string{}, &product)
}

// Renders the edit product page.
func (h *Handler) EditProductView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.Repo.Product.GetProductByID(productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "editProduct", product)
}

// Updates a product in the database.
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Initialize error messages slice
	var responseMessages []string

	//Check for empty fields
	ProductName := r.FormValue("product_name")
	ProductPrice := r.FormValue("price")
	ProductDescription := r.FormValue("description")

	if ProductName == "" || ProductPrice == "" || ProductDescription == "" {
		responseMessages = append(responseMessages, "All Fields Are Required")
		sendProductMessage(w, responseMessages, nil)
		return
	}

	price, err := strconv.ParseFloat(ProductPrice, 64)
	if err != nil {
		responseMessages = append(responseMessages, "Invalid Price")
		sendProductMessage(w, responseMessages, nil)
		return
	}

	product := models.Product{
		ProductID:   productID,
		ProductName: ProductName,
		Price:       price,
		Description: ProductDescription,
	}

	err = h.Repo.Product.UpdateProduct(&product)
	if err != nil {
		responseMessages = append(responseMessages, "Error Updating Product: "+err.Error())
		sendProductMessage(w, responseMessages, nil)
		return
	}

	//Get and send updated product
	updatedProduct, _ := h.Repo.Product.GetProductByID(productID)

	//Fake Latency
	time.Sleep(2 * time.Second)

	sendProductMessage(w, []string{}, updatedProduct)
}

// Deletes a product from the database.
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, _ := h.Repo.Product.GetProductByID(productID)

	err = h.Repo.Product.DeleteProduct(productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Remove product image
	productImagePath := filepath.Join("static/uploads", product.ProductImage)
	os.Remove(productImagePath)

	//Fake Latency
	time.Sleep(2 * time.Second)

	tmpl.ExecuteTemplate(w, "allProducts", nil)
}

// Renders the shop home page.
func (h *Handler) ShoppingHomepage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		OrderItems []models.OrderItem
	}{
		OrderItems: cartItems,
	}
	tmpl.ExecuteTemplate(w, "homepage", data)
}

// Renders the items view in the home page.
func (h *Handler) ShoppingItemsView(w http.ResponseWriter, r *http.Request) {
	time.Sleep(1 * time.Second) 	// Fake Latency
	products, _ := h.Repo.Product.GetProducts("product_image != ''")
	tmpl.ExecuteTemplate(w, "shoppingItems", products)
}

// Renders the cart view in the home page.
func (h *Handler) CartView(w http.ResponseWriter, r *http.Request) {
	data := struct {
		OrderItems []models.OrderItem
		Message    string
		AlertType  string
		TotalCost  float64
	}{
		OrderItems: cartItems,
		Message:    "",
		AlertType:  "",
		TotalCost:  getTotalCartCost(),
	}
	tmpl.ExecuteTemplate(w, "cartItems", data)
}

// Adds a product to the cart.
func (h *Handler) AddToCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["product_id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Generate a new order id for the session if one does not exist
	if currentCartOrderId == uuid.Nil { currentCartOrderId = uuid.New() }

	// Check if product already exists in order items
	exists := false
	for _, item := range cartItems {
		if item.ProductID == productID {
			exists = true
			break
		}
	}

	//Get the Product
	product, _ := h.Repo.Product.GetProductByID(productID)

	cartMessage := ""
	alertType := ""

	if !exists {
		// Create a new order item
		newOrderItem := models.OrderItem{
			OrderID:   currentCartOrderId,
			ProductID: productID,
			Quantity:  1, // Initial quantity of 1
			Product:   *product,
		}

		// Add new order item to the array
		cartItems = append(cartItems, newOrderItem)

		cartMessage = product.ProductName + " successfully added"
		alertType = "success"
	} else {
		cartMessage = product.ProductName + " already exists in cart"
		alertType = "danger"
	}

	data := struct {
		OrderItems []models.OrderItem
		Message    string
		AlertType  string
		TotalCost  float64
	}{
		OrderItems: cartItems,
		Message:    cartMessage,
		AlertType:  alertType,
		TotalCost:  getTotalCartCost(),
	}

	tmpl.ExecuteTemplate(w, "cartItems", data)
}

// Renders the checkout view in the home page.
func (h *Handler) ShoppingCartView(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "shoppingCart", cartItems)
}

// Updates the quantity of a product in the cart.
func (h *Handler) UpdateOrderItemQuantity(w http.ResponseWriter, r *http.Request) {
	// Get product ID and action from URL parameters
	cartMessage := ""
	refreshCartList := false //Signals a refresh of cart items when an item is removed

	productID, err := uuid.Parse(r.URL.Query().Get("product_id"))
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	action := r.URL.Query().Get("action")

	// Find the order item
	var itemIndex int
	for i, item := range cartItems {
		if item.ProductID == productID {
			itemIndex = i
			break
		}
	}
	if itemIndex == -1 {
		http.Error(w, "Product not found in order", http.StatusNotFound)
		return
	}

	// Update quantity based on action
	switch action {
		case "add":
			cartItems[itemIndex].Quantity++
		case "subtract":
			cartItems[itemIndex].Quantity--
			if cartItems[itemIndex].Quantity == 0 {
				// Remove item if quantity is 0
				cartItems = append(cartItems[:itemIndex], cartItems[itemIndex+1:]...)
				refreshCartList = true
			}
		case "remove":
			// Remove item regardless of quantity
			cartItems = append(cartItems[:itemIndex], cartItems[itemIndex+1:]...)
			refreshCartList = true
		default:
			/* http.Error(w, "Invalid action", http.StatusBadRequest)
			return */
			cartMessage = "Invalid Action"
	}

	// Respond to the request
	//fmt.Fprintf(w, "Order item updated")
	data := struct {
		OrderItems       []models.OrderItem
		Message          string
		AlertType        string
		TotalCost        float64
		Action           string
		RefreshCartItems bool
	}{
		OrderItems:       cartItems,
		Message:          cartMessage,
		AlertType:        "info",
		TotalCost:        getTotalCartCost(),
		Action:           action,
		RefreshCartItems: refreshCartList,
	}

	tmpl.ExecuteTemplate(w, "updateShoppingCart", data)
}

// Places an order.
func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	for i := range cartItems {
		cartItems[i].Cost = float64(cartItems[i].Quantity) * cartItems[i].Product.Price
	}

	err := h.Repo.Order.PlaceOrderWithItems(cartItems)
	if err != nil {
		http.Error(w, "Error Placing Order "+err.Error(), http.StatusBadRequest)
		return
	}

	displayItems := cartItems
	totalCost := getTotalCartCost()

	//Empty the cart items
	cartItems = []models.OrderItem{}
	currentCartOrderId = uuid.Nil

	data := struct {
		OrderItems []models.OrderItem
		TotalCost  float64
	}{
		OrderItems: displayItems,
		TotalCost:  totalCost,
	}

	tmpl.ExecuteTemplate(w, "orderComplete", data)
}

// Renders the order page.
func (h *Handler) OrdersPage(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "orders", nil)
}

// Renders all orders in the database in the table.
func (h *Handler) AllOrdersView(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "allOrders", nil)
}

// Lists the orders in the database in a paginated way.
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 { page = 1 }

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 { limit = 10 } // Default limit

	offset := (page - 1) * limit

	orders, err := h.Repo.Order.ListOrders(limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalOrders, err := h.Repo.Order.GetTotalOrdersCount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(totalOrders) / float64(limit)))
	previousPage := page - 1
	nextPage := page + 1
	pageButtonsRange := makeRange(1, totalPages)

	data := struct {
		Orders           []models.Order
		CurrentPage      int
		TotalPages       int
		Limit            int
		PreviousPage     int
		NextPage         int
		PageButtonsRange []int
	}{
		Orders:           orders,
		CurrentPage:      page,
		TotalPages:       totalPages,
		Limit:            limit,
		PreviousPage:     previousPage,
		NextPage:         nextPage,
		PageButtonsRange: pageButtonsRange,
	}

	//Fake Latency
	//time.Sleep(2 * time.Second)

	tmpl.ExecuteTemplate(w, "orderRows", data)
}

// Renders the order detail page.
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.Repo.Order.GetOrderWithProducts(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalCost := 0.0
	for _, item := range order.Items {
		totalCost += float64(item.Quantity) * item.Product.Price
	}

	order.OrderStatus = strings.ToUpper(order.OrderStatus)

	data := struct {
		Order     models.Order
		TotalCost float64
	}{
		Order:     *order,
		TotalCost: totalCost,
	}

	tmpl.ExecuteTemplate(w, "viewOrder", data)
}
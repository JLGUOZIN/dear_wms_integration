package services

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Product struct {
	SKU         string `json:"SKU"`
	Name        string `json:"Name"`
	StockOnHand int    `json:"StockOnHand"`
}

// HandleProductAvailability retrieves product stock availability.
func HandleProductAvailability(c echo.Context) error {
	sku := c.QueryParam("sku")

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("https://inventory.dearsystems.com/ExternalApi/v2/ref/productavailability?Sku=%s", sku))

	if err != nil || resp.StatusCode() != http.StatusOK {
		fmt.Printf("Error retrieving product availability: %v\n", err)
		return c.JSON(http.StatusInternalServerError, "Failed to retrieve product availability")
	}

	var productAvailability []Product
	if err := json.Unmarshal(resp.Body(), &productAvailability); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return c.JSON(http.StatusInternalServerError, "Failed to parse response")
	}

	return c.JSON(http.StatusOK, productAvailability)
}

// CreateNewProduct handles creation of a new product in the DEAR system.
func CreateNewProduct(c echo.Context) error {
	var product Product
	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	// Example POST request to DEAR API
	client := resty.New()
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(product).
		Post("https://inventory.dearsystems.com/ExternalApi/v2/product")

	if err != nil {
		fmt.Printf("Error creating product: %v\n", err)
		return c.JSON(http.StatusInternalServerError, "Failed to create product")
	}

	return c.JSON(http.StatusCreated, "Product created successfully")
}

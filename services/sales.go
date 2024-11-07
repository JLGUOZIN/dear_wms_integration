package services

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

type SaleRequest struct {
	Customer        string  `json:"Customer"`
	Location        string  `json:"Location"`
	AdditionalNotes string  `json:"AdditionalNotes"`
	TotalAmount     float64 `json:"TotalAmount"`
}

// AddSale processes the addition of a new sale.
func AddSale(c echo.Context) error {
	var saleRequest SaleRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&saleRequest); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	// Here, process saleRequest with external API calls to DEAR as necessary

	fmt.Printf("Sale added successfully: %+v\n", saleRequest)
	return c.JSON(http.StatusOK, "Sale added successfully")
}

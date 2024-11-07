package services

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"net/http"
)

type CustomerRequest struct {
	Name              string `json:"name"`
	Status            string `json:"status"`
	Currency          string `json:"currency"`
	PaymentTerm       string `json:"paymentTerm"`
	TaxRule           string `json:"taxRule"`
	AccountReceivable string `json:"accountReceivable"`
	RevenueAccount    string `json:"revenueAccount"`
	PriceTier         string `json:"priceTier"`
	Tags              string `json:"tags"`
}

// SaveCustomer handles saving a customer record to the DEAR system.
func SaveCustomer(c echo.Context) error {
	var customerReq CustomerRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&customerReq); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	// Default values for customer
	customerReq.Currency = "SGD"
	customerReq.PaymentTerm = "Same Day"
	customerReq.TaxRule = "Tax Exempt"
	customerReq.AccountReceivable = "1200"
	customerReq.RevenueAccount = "4000"
	customerReq.PriceTier = "MSRP"
	customerReq.Status = "Active"

	client := resty.New()
	response, err := client.R().
		SetHeaders(viper.GetStringMapString("app.dear_header")).
		SetBody(customerReq).
		Post("https://inventory.dearsystems.com/ExternalApi/v2/customer")

	if err != nil {
		fmt.Printf("Error saving customer: %v\n", err)
		return c.JSON(http.StatusInternalServerError, "Failed to save customer")
	}

	fmt.Printf("Response from DEAR API: %v\n", response)
	return c.JSON(http.StatusOK, "Customer saved successfully")
}

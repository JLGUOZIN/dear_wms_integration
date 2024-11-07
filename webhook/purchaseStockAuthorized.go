package webhook

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
)

type PurchaseStockAuthorizedResponse struct {
	PurchaseOrderNumber string `json:"purchaseOrderNumber"`
	TaskID              string `json:"taskID"`
}

type PurchaseStockAuthorizedDetailResponse struct {
	Status string                        `json:"status"`
	Lines  []PurchaseStockAuthorizedLine `json:"lines"`
}

type PurchaseStockAuthorizedLine struct {
	SKU      string  `json:"sku"`
	Location string  `json:"location"`
	Quantity float64 `json:"quantity"`
}

// PurchaseStockAuthorized processes the stock authorization based on provided request body.
func PurchaseStockAuthorized(c echo.Context) error {
	// Read the request body as []byte
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	var purchaseResponse PurchaseStockAuthorizedResponse
	if err := json.Unmarshal(body, &purchaseResponse); err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to parse request body")
	}

	details := GetPurchaseStockAuthorizedDetail(purchaseResponse.TaskID)
	if details.Status != "AUTHORISED" {
		return c.JSON(http.StatusOK, "No action required as status is not AUTHORISED")
	}

	var imsRequests []ImsRequest
	for _, response := range details.Lines {
		if response.Location == "AIRPAK / SNP Store" {
			data := ImsRequest{
				Sku:        response.SKU,
				Adjustment: response.Quantity,
				Reason:     fmt.Sprintf("Dear System Stock Received (%s)", purchaseResponse.PurchaseOrderNumber),
			}
			imsRequests = append(imsRequests, data)
		}
	}
	go syncStockToBE(imsRequests)

	return c.JSON(http.StatusOK, "Purchase stock authorization processed")
}

// GetPurchaseStockAuthorizedDetail fetches stock authorization details based on TaskID.
func GetPurchaseStockAuthorizedDetail(taskID string) PurchaseStockAuthorizedDetailResponse {
	req, err := http.NewRequest("GET", "https://inventory.dearsystems.com/ExternalApi/v2/purchase/stock?TaskID="+taskID, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return PurchaseStockAuthorizedDetailResponse{}
	}

	req.Header = createDefaultHeaders()
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error fetching details:", err)
		return PurchaseStockAuthorizedDetailResponse{}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return PurchaseStockAuthorizedDetailResponse{}
	}

	fmt.Println("GetPurchaseStockAuthorizedDetail > ", string(body))
	var details PurchaseStockAuthorizedDetailResponse
	json.Unmarshal(body, &details)
	return details
}

func createDefaultHeaders() http.Header {
	return http.Header{
		"Content-Type":            []string{"application/json"},
		"Api-Auth-Accountid":      []string{"your-account-id"},
		"Api-Auth-Applicationkey": []string{"your-application-key"},
	}
}

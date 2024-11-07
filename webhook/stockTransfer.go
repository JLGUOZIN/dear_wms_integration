package webhook

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
)

type StockTransferResponse struct {
	Number       string `json:"number"`
	FromLocation string `json:"fromLocation"`
	ToLocation   string `json:"toLocation"`
	TaskID       string `json:"taskID"`
}

type StockTransferDetailResponse struct {
	Status string                    `json:"status"`
	Lines  []StockTransferDetailLine `json:"lines"`
}

type StockTransferDetailLine struct {
	SKU               string  `json:"sku"`
	QuantityAvailable float64 `json:"quantityAvailable"`
	TransferQuantity  float64 `json:"transferQuantity"`
}

// StockTransfer processes stock transfer requests.
func StockTransfer(c echo.Context) error {
	// Read the request body as []byte
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	var transferResponse StockTransferResponse
	if err := json.Unmarshal(body, &transferResponse); err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to parse request body")
	}

	adjustment := determineAdjustment(transferResponse)
	if adjustment == 0 {
		return c.JSON(http.StatusOK, "No adjustment required")
	}

	details := GetStockTransferDetail(transferResponse)
	reason := fmt.Sprintf("Dear System Stock Transfer (%s)", transferResponse.Number)

	var imsRequests []ImsRequest
	for _, line := range details.Lines {
		data := ImsRequest{
			Sku:        line.SKU,
			Adjustment: line.TransferQuantity * adjustment,
			Reason:     reason,
		}
		imsRequests = append(imsRequests, data)
	}
	go syncStockToBE(imsRequests)

	return c.JSON(http.StatusOK, "Stock transfer processed")
}

// determineAdjustment determines the adjustment multiplier for stock transfers.
func determineAdjustment(response StockTransferResponse) float64 {
	switch {
	case response.FromLocation == "WMS Location":
		return -1
	case response.ToLocation == "WMS Location":
		return 1
	default:
		return 0
	}
}

// GetStockTransferDetail fetches stock transfer details.
func GetStockTransferDetail(transferResponse StockTransferResponse) StockTransferDetailResponse {
	taskID := transferResponse.TaskID
	req, err := http.NewRequest("GET", "https://inventory.dearsystems.com/ExternalApi/v2/stockTransfer/order?TaskID="+taskID, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return StockTransferDetailResponse{}
	}

	req.Header = createDefaultHeaders()
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error fetching stock transfer details:", err)
		return StockTransferDetailResponse{}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return StockTransferDetailResponse{}
	}
	fmt.Println("StockTransferDetail > ", string(body))

	var details StockTransferDetailResponse
	json.Unmarshal(body, &details)
	return details
}

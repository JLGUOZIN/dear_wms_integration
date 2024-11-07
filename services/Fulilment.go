package services

import (
	"time"
)

type LineItem struct {
	ProductID string  `json:"ProductID"`
	SKU       string  `json:"SKU"`
	Name      string  `json:"Name"`
	Location  string  `json:"Location"`
	Quantity  float32 `json:"Quantity"`
	Price     float32 `json:"Price"`
}

type FulfilmentTask struct {
	TaskID     string `json:"TaskID"`
	Status     string `json:"Status"`
	Lines      []LineItem
	InvoiceDue time.Time `json:"InvoiceDue"`
}

// prepareFulfilment creates a fulfilment task based on sale details.
func prepareFulfilment(saleID, taskID, from string, items []LineItem) FulfilmentTask {
	return FulfilmentTask{
		TaskID:     taskID,
		Status:     "AUTHORISED",
		Lines:      items,
		InvoiceDue: time.Now().AddDate(0, 0, 7), // example due date
	}
}

// getFulfilmentInfo retrieves fulfilment info for a specific sale.
func getFulfilmentInfo(saleID string) (FulfilmentTask, error) {
	// Example call to external service, with error handling
	response := FulfilmentTask{}
	// Fill response with external API call details
	return response, nil
}

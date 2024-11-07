package cronTasks

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"log"
	"time"
)

type SkuStock struct {
	SKU        string
	StoreName  string
	StockLevel int
}

type ResponseSkuStock struct {
	SkuStock
	DearStock int
}

// sendOpsEmail sends an email with stock discrepancies to the ops team.
func sendOpsEmail(missingSku []SkuStock, untalliedStock []ResponseSkuStock) {
	missingSKuBuf, untalliedStockBuf := createCSVBuffers(missingSku, untalliedStock)

	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	svc := ses.New(sess)
	rawMessage := constructEmailMessage(missingSKuBuf, untalliedStockBuf)

	input := &ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{Data: []byte(rawMessage)},
		Destinations: []*string{
			aws.String("receipient@youremail.com"),
		},
		Source: aws.String("sender@youremail.com"),
	}

	_, err = svc.SendRawEmail(input)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
	} else {
		fmt.Println("Email sent successfully.")
	}
}

func createCSVBuffers(missingSku []SkuStock, untalliedStock []ResponseSkuStock) (bytes.Buffer, bytes.Buffer) {
	missingSKuBuf, untalliedStockBuf := bytes.Buffer{}, bytes.Buffer{}
	missingSkuCSV := csv.NewWriter(&missingSKuBuf)
	untalliedStockCSV := csv.NewWriter(&untalliedStockBuf)

	// Populate CSV for missing SKU
	if len(missingSku) > 0 {
		_ = missingSkuCSV.Write([]string{"Missing SKU", "Stock in SNP"})
		for _, row := range missingSku {
			_ = missingSkuCSV.Write([]string{row.SKU, fmt.Sprintf("%d", row.StockLevel)})
		}
		missingSkuCSV.Flush()
	}

	// Populate CSV for untallied stock
	if len(untalliedStock) > 0 {
		_ = untalliedStockCSV.Write([]string{"SKU", "Store Name", "Stock in DEAR", "Stock in SNP"})
		for _, row := range untalliedStock {
			_ = untalliedStockCSV.Write([]string{row.SKU, row.StoreName, fmt.Sprintf("%d", row.DearStock), fmt.Sprintf("%d", row.StockLevel)})
		}
		untalliedStockCSV.Flush()
	}

	return missingSKuBuf, untalliedStockBuf
}

// constructEmailMessage creates the MIME email format for SES.
func constructEmailMessage(missingBuf, untalliedBuf bytes.Buffer) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	subject := "Stock Discrepancy Report"
	return fmt.Sprintf(`From: sender@youremail.com
To: receipient@youremail.com
Subject: %s
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="NextPart"

--NextPart
Content-Type: text/html; charset=us-ascii

Hello,

<p>Please find attached the stock discrepancy report.</p>

--NextPart
Content-Type: text/csv; name="missing-sku-%s.csv"
Content-Disposition: attachment; filename="missing-sku-%s.csv"

%s
--NextPart
Content-Type: text/csv; name="untallied-stock-%s.csv"
Content-Disposition: attachment; filename="untallied-stock-%s.csv"

%s
--NextPart--`, subject, timestamp, timestamp, missingBuf.String(), timestamp, timestamp, untalliedBuf.String())
}

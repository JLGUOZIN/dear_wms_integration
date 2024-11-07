package cronTasks

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	_ "github.com/lib/pq"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type SkuStock struct {
	sku        string
	storeName  string
	stockLevel int
}
type ResponseSkuStock struct {
	SkuStock
	dearStock int
}
type DearSkuStock struct {
	ID          string      `json:"ID"`
	SKU         string      `json:"SKU"`
	Name        string      `json:"Name"`
	Barcode     interface{} `json:"Barcode"`
	Location    string      `json:"Location"`
	Bin         interface{} `json:"Bin"`
	Batch       interface{} `json:"Batch"`
	ExpiryDate  interface{} `json:"ExpiryDate"`
	OnHand      float64     `json:"OnHand"`
	Allocated   float64     `json:"Allocated"`
	Available   float64     `json:"Available"`
	OnOrder     float64     `json:"OnOrder"`
	StockOnHand float64     `json:"StockOnHand"`
	InTransit   float64     `json:"InTransit"`
}
type ProductAvailability struct {
	Total                   int            `json:"Total"`
	Page                    int            `json:"Page"`
	ProductAvailabilityList []DearSkuStock `json:"ProductAvailabilityList"`
}

func SendOpsEmail(missingSku []SkuStock, untalliedStock []ResponseSkuStock) {
	missingSKuBuf := bytes.Buffer{}
	untalliedStockBuf := bytes.Buffer{}

	if len(missingSku) != 0 {
		missingSkuCSV := csv.NewWriter(&missingSKuBuf)
		err := missingSkuCSV.Write([]string{
			"Missing Sku",
			"Stock in Snp",
		})
		if err != nil {
			panic(err)
		}
		for _, row := range missingSku {
			err := missingSkuCSV.Write([]string{
				row.sku,
				strconv.Itoa(row.stockLevel),
			})
			if err != nil {
				panic(err)
			}
		}
		missingSkuCSV.Flush()
	}
	if len(untalliedStock) != 0 {
		untalliedStockCSV := csv.NewWriter(&untalliedStockBuf)
		err := untalliedStockCSV.Write([]string{
			"Sku",
			"Store Name",
			"Stock in Dear",
			"Stock in Snp",
		})
		if err != nil {
			panic(err)
		}
		for _, row := range untalliedStock {
			err := untalliedStockCSV.Write([]string{
				row.sku,
				row.storeName,
				strconv.Itoa(row.dearStock),
				strconv.Itoa(row.stockLevel),
			})
			if err != nil {
				panic(err)
			}
		}
		untalliedStockCSV.Flush()
	}
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	svc := ses.New(sess)

	// Specify the email contents.
	sender := "Sender Email <sender@youremail.com>"
	recipients := "receipient@youremail.com"
	subject := "Dear's stock not tally with Snp Stock"

	rawMessage := "From: Sender Email <sender@youremail.com>\n"
	rawMessage = rawMessage + "To: " + recipients + "\n"
	rawMessage = rawMessage + "Subject: " + subject + "\n"
	rawMessage = rawMessage + "MIME-Version: 1.0\n"
	rawMessage = rawMessage + "Content-Type: multipart/mixed; boundary=\"NextPart\"\n\n"
	rawMessage = rawMessage + "--NextPart\n"
	rawMessage = rawMessage + "Content-Type: text/html; charset=us-ascii\n\n"
	rawMessage = rawMessage + "Hi all,<p>The attachments are the SNP-Dear stock reports.</p>Thanks." + "\n\n"
	currentTime := time.Now().Add(8 * time.Hour).Format("2006-01-02 15:04:05")
	if missingSKuBuf.Len() > 0 {
		rawMessage = rawMessage + "--NextPart\n"
		rawMessage = rawMessage + "Content-Type: text/csv; name=\"missing-sku-in-dear-" + currentTime + ".csv\"\n"
		rawMessage = rawMessage + "Content-Disposition: attachment; filename=\"missing-sku-in-dear-" + currentTime + ".csv\"\n\n"
		rawMessage = rawMessage + missingSKuBuf.String() + "\n\n"
	}
	if untalliedStockBuf.Len() > 0 {
		rawMessage = rawMessage + "--NextPart\n"
		rawMessage = rawMessage + "Content-Type: text/csv; name=\"untallied-stock-dear-" + currentTime + ".csv\"\n"
		rawMessage = rawMessage + "Content-Disposition: attachment; filename=\"untallied-stock-dear-" + currentTime + ".csv\"\n\n"
		rawMessage = rawMessage + untalliedStockBuf.String() + "\n\n"

	}

	destination := make([]*string, len(strings.Split(recipients, ", ")))
	for i, recipient := range strings.Split(recipients, ", ") {
		destination[i] = aws.String(recipient)
	}
	// Create the email input object
	input := &ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{
			Data: []byte(rawMessage),
		},
		Destinations: destination,
		Source:       aws.String(sender),
	}

	// Send the email
	_, err = svc.SendRawEmail(input)
	if err != nil {
		fmt.Println("failed to send email,", err)
	}

	fmt.Println("email sent successfully")
}

func CallDearForStock(page int, untalliedStock []ResponseSkuStock, skus []SkuStock) []ResponseSkuStock {
	nextSkus := skus
	newArr := untalliedStock
	encodedUrl := "https://inventory.dearsystems.com/ExternalApi/v2/ref/productavailability?Page=" + strconv.Itoa(page) + "&Location=" + url.QueryEscape("[WMS Location]")

	req, err := http.NewRequest("GET", encodedUrl, nil)
	req.Header = http.Header{
		"Content-Type":            []string{"application/json"},
		"api-auth-accountid":      []string{"your_account_id"},
		"api-auth-applicationkey": []string{"your_application_id"},
	}

	if err != nil {
		fmt.Println(err)
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	var data ProductAvailability
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		panic(err)
	}
	if page == 1 {
		fmt.Println("DEAR total", data.Total)
	}

	for _, value := range data.ProductAvailabilityList {
		skuIndex := len(nextSkus)
		for index, currentSku := range nextSkus {
			if strings.HasSuffix(strings.ToLower(currentSku.sku), strings.ToLower(value.SKU)) {
				skuIndex = index
				break
			}
		}
		if skuIndex != len(nextSkus) {
			currentSku := nextSkus[skuIndex]
			if currentSku.stockLevel != int(value.Available) {
				responseSkuStock := ResponseSkuStock{SkuStock{currentSku.sku, currentSku.storeName, currentSku.stockLevel}, int(value.Available)}
				newArr = append(newArr, responseSkuStock)
			}
			// to optimise the search performance, if found then remove in the skus array
			nextSkus = append(nextSkus[:skuIndex], nextSkus[skuIndex+1:]...)
		}

	}
	if data.Total > page*100 {
		return CallDearForStock(page+1, newArr, nextSkus)
	}
	var notEmptySKU []SkuStock
	for skuIndex := range nextSkus {
		currentSKU := nextSkus[skuIndex]
		if currentSKU.stockLevel != 0 {
			notEmptySKU = append(notEmptySKU, currentSKU)
		}
	}
	fmt.Println("not found in dear or OOS in dear or not in WMS Location", len(notEmptySKU))
	if len(notEmptySKU) != 0 || len(newArr) != 0 {
		SendOpsEmail(notEmptySKU, newArr)
	}
	return newArr
}

// CheckDearAndSystemStock cron job always run in production, it's safe to use prod env
func CheckDearAndSystemStock() {
	dbUrl := os.Getenv("DB_URL_PROD")
	dbUrl = strings.Replace(dbUrl, "jdbc:postgresql://", "postgresql://"+os.Getenv("DB_USERNAME_PROD")+":"+os.Getenv("DB_PASSWORD_PROD")+"@", -1)
	dbUrl = strings.Replace(dbUrl, "useSSL=false", "sslmode=disable", -1)
	dbUrl = strings.Replace(dbUrl, "serverTimezone=UTC", "timezone=UTC", -1)
	dbUrl = strings.Replace(dbUrl, "&useLegacyDatetimeCode=false", "", -1)
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	rows, err := db.Query("SELECT pv.sku, s.name, SUM(sl.adjustment) " +
		"FROM product p JOIN product_variation pv ON p.product_id = pv.product_id " +
		"JOIN store s ON s.store_id = p.store_id " +
		"JOIN product_variation_stock_management pvsm ON pv.product_variation_id = pvsm.product_variation_id " +
		"LEFT JOIN stock_logger sl ON sl.product_variation_stock_management_id = pvsm.product_variation_stock_management_id " +
		"WHERE (s.commission_type IN ('wholesale', 'consignment') OR s.store_id = 4218) AND s.is_published is true " +
		"AND p.active = TRUE AND p.approval_state = 'approved' " +
		"GROUP BY pv.sku, s.name")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var skus []SkuStock
	// Process the results
	for rows.Next() {
		var sku, name, stock_level string
		if err := rows.Scan(&sku, &name, &stock_level); err != nil {
			panic(err)
		}
		floatVal, err := strconv.ParseFloat(stock_level, 64)
		if err != nil {
			panic(err)
		}
		stockLevel := int(floatVal)
		shortenSku := sku
		parts := strings.Split(sku, "-")

		if len(parts) > 1 {
			if !strings.Contains(parts[0], "GWP") {
				shortenSku = parts[1]
			}
		}

		skus = append(skus, SkuStock{sku: shortenSku, storeName: name, stockLevel: stockLevel})
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	page := 1
	fmt.Println("SNP total", len(skus))
	CallDearForStock(page, []ResponseSkuStock{}, skus)
}

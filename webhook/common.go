package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
)

type ImsRequest struct {
	ID         string  `json:"id"`
	Sku        string  `json:"sku"`
	Available  string  `json:"available"`
	Adjustment float64 `json:"adjustment"`
	Reason     string  `json:"reason"`
}

func syncStockToBE(imsRequest []ImsRequest) {

	client := http.Client{}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(imsRequest)
	fmt.Println(imsRequest)

	req, err := http.NewRequest("POST", viper.GetString("app.url.BE")+"/api/v1/ims/updateStock", bytes.NewBuffer(reqBodyBytes.Bytes()))
	req.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{"Basic your_token"},
	}

	if err != nil {
		fmt.Println(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(body))

}

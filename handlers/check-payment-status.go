package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"template/database"
	"template/function"
	"template/models"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type StatusResponse struct {
	Hashcode       string `json:"hashcode"`
	Payment_status string `json:"payment_status"`
	Payment_id     string `json:"payment_id"`
}
type StatusRequest struct {
	Status_coin    string `json:"status_coin"`
	Status_address string `json:"status_address"`
	Status_transid string `json:"status_transid"`
	Status_coinid  string `json:"status_coinid"`
}

// Function for generate random password
func cryptoAmountFormat(receivedAmount string) float64 {
	// convert string to float value
	receivedAmt, err := strconv.ParseFloat(receivedAmount, 64)
	if err != nil {
		fmt.Println(" Error convert string to float value :")
	}
	// Convert the integer to a float64 with 6 decimal places
	AmountInFloat := float64(receivedAmt) / 1000000

	// Format the float to 6 decimal places as a string
	formattedResult := strconv.FormatFloat(AmountInFloat, 'f', 6, 64)

	// convert string to float value
	receivedAmountNew, err := strconv.ParseFloat(formattedResult, 64)
	if err != nil {
		fmt.Println(" Error convert string to float value :")
	}

	return receivedAmountNew
}

// Function for fetch payment status by ajax call url from checkout page
func FetchPaymentStatus(c *fiber.Ctx) error {

	// Get Data from ajax
	req := new(StatusRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request",
		})
	}

	status_coin := req.Status_coin // for set condition
	//fmt.Println("status_coin==========>", status_coin)
	status_address := req.Status_address
	//fmt.Println("status_address==========>", status_address)
	status_transid := req.Status_transid
	//fmt.Println("status_transid ==========>", status_transid)
	status_coinid := req.Status_coinid
	//fmt.Println("Coin ID ==========>", status_coinid)

	// For Address
	coinAddress := models.AddressListing{}
	database.DB.Db.Table("coin_address").Select("token_details,lastupdate").Where("coin = ? AND address=? ", status_coin, status_address).Find(&coinAddress)

	fetchToken := coinAddress.Token_details
	//fetchTimestamp := coinAddress.Lastupdate

	//fmt.Println(fetchTimestamp)

	var tokenData models.OnlineTokenData
	if err := json.Unmarshal([]byte(fetchToken), &tokenData); err != nil {
		fmt.Println(err)
	}
	//fmt.Println(tokenData)

	//fmt.Println("Contact Address: - ", tokenData.TokenId) // fetch from Address table json
	//fmt.Println("Address :- ", status_address)

	receivedFrom := ""
	receivedTo := ""
	receivedHash := ""
	receivedFinalResult := ""
	fetchTimestamp := ""
	responseTimestamp := ""
	receivedAmountNew := 0.00
	var body []byte

	//Start status By Crypto ID

	if status_coinid == "17" { // For USDT Tether TRX/TRC20

		// URL of the external site to fetch JSON from
		url := "https://apilist.tronscanapi.com/api/token_trc20/transfers-with-status?limit=1&start=0&trc20Id=" + tokenData.TokenId + "&address=" + status_address
		//fmt.Println("Path :- ", path)

		//////////////////////////////////////
		resp, err := http.Get(url)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Failed to fetch data")
		}
		defer resp.Body.Close()

		// Reading the response body
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Failed to read response body")
		}
		// Initialize the Response struct
		var responseD models.OnlineResponse

		// Unmarshal the byte data into the struct
		err = json.Unmarshal(body, &responseD)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Failed to parse JSON")
		}
		//fmt.Println(" Data :", responseD)
		receivedAmount := responseD.Data[0].Amount

		// convert string to float value
		receivedAmt, err := strconv.ParseFloat(receivedAmount, 64)
		if err != nil {
			fmt.Println(" Error convert string to float value :")
		}
		// Convert the integer to a float64 with 6 decimal places
		AmountInFloat := float64(receivedAmt) / 1000000

		// Format the float to 6 decimal places as a string
		formattedResult := strconv.FormatFloat(AmountInFloat, 'f', 6, 64)

		// convert string to float value
		receivedAmountNew, err = strconv.ParseFloat(formattedResult, 64)
		if err != nil {
			fmt.Println(" Error convert string to float value :")
		}

		//fmt.Println("receivedAmountNew", receivedAmountNew)
		receivedFrom = responseD.Data[0].From
		receivedTo = responseD.Data[0].To
		receivedHash = responseD.Data[0].Hash
		receivedBlockTimestamp := responseD.Data[0].BlockTimestamp
		//receivedDecimals := responseD.Data[0].Decimals
		receivedFinalResult = responseD.Data[0].FinalResult
		//receivedEventType := responseD.Data[0].EventType

		//fetchTimestamp = "2024-08-23 16:00:09"
		fetchTimestamp = coinAddress.Lastupdate.Format("2006-01-02 15:04:05")
		responseTimestamp = time.Unix(receivedBlockTimestamp/1000, 0).Format("2006-01-02 15:04:05")
		//fmt.Println("fetchTimestamp => ", fetchTimestamp)
		//fmt.Println("Response Time => ", responseTimestamp)

	} else if status_coinid == "28" { // For ADA Cardano Mainnet

		// Construct the API URL
		url := "https://api.cardanoscan.io/api/v1/transaction/list?address=" + tokenData.TokenId + "&pageNo=1&limit=1&order=desc"
		apiKey := os.Getenv("CARDANO_SCAN_API_KEY")

		// Create a new HTTP request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("error creating HTTP request: %w", err)
		}
		// Send Api in header
		req.Header.Set("apiKey", apiKey)
		// Make request
		resp, err := httpClient.Do(req)
		if err != nil {
			fmt.Println("error sending HTTP request: %w", err)
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("error reading response body: %w", err)
		}

		// Initialize the Response struct
		var responseD models.CardanoResponse

		// Unmarshal the byte data into the struct
		err = json.Unmarshal(body, &responseD)
		if err != nil {
			fmt.Println("Failed to parse JSON")
		}
		//fmt.Println(" Data :", responseD)
		receivedHash = responseD.Data[0].Hash
		responseTimestamp = responseD.Data[0].Timestamp.Format("2006-01-02 15:04:05")
		//receivedFees = responseD.Data[0].Fees
		receivedFinalStatus := responseD.Data[0].Status
		receivedFrom = responseD.Data[0].Outputs[0].Address
		receivedTo = responseD.Data[0].Outputs[1].Address
		receivedAmount := responseD.Data[0].Outputs[1].Value
		receivedAmountNew = cryptoAmountFormat(receivedAmount)
		receivedFinalResult = ""
		if receivedFinalStatus {
			receivedFinalResult = "SUCCESS"
		} else {
			receivedFinalResult = "FAILED"
		}

		//fetchTimestamp = "2024-06-19 16:00:09"
		fetchTimestamp = coinAddress.Lastupdate.Format("2006-01-02 15:04:05")

	} else {

		fmt.Println("Crypto Not Supported ==> ", status_coin)
	}
	//End $$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$//$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$//$$$$$$$$$$$$$$$$$$$$$

	// Calculate the difference
	layout := "2006-01-02 15:04:05" // Define the layout for parsing
	// Calculate the difference
	// Parse the dateTime strings to time.Time objects
	dateTime1, err := time.Parse(layout, fetchTimestamp)
	if err != nil {
		return c.Status(400).SendString("Error parsing dateTime1: " + err.Error())
	}

	dateTime2, err := time.Parse(layout, responseTimestamp)
	if err != nil {
		return c.Status(400).SendString("Error parsing dateTime2: " + err.Error())
	}

	duration := dateTime2.Sub(dateTime1)

	// Get the difference in seconds
	seconds := int(duration.Seconds())
	receivedSubStatus := 0

	invoiceAmount := function.GetConvertedAmountByTransID(status_transid)
	receivedAmount := float64(receivedAmountNew)

	// Check if the amounts are within 2% tolerance
	if function.IsPaymentSuccess(invoiceAmount, receivedAmount) {
		if invoiceAmount == receivedAmount {
			//fmt.Println("Success Pay")
			receivedSubStatus = 1

		} else if invoiceAmount < receivedAmount {
			//fmt.Println("Over Pay")
			receivedSubStatus = 2

		} else if invoiceAmount > receivedAmount {
			//fmt.Println("Under Pay")
			receivedSubStatus = 3

		} else {
			//fmt.Println("Payment Failed")
			receivedSubStatus = 9
			receivedFinalResult = "FAILED"
		}
	} else {
		receivedSubStatus = 9
		receivedFinalResult = "FAILED"
		fmt.Println("Payment Failed")
	}

	//fmt.Println("invoiceAmount => ", invoiceAmount)
	//fmt.Println("receivedAmount => ", receivedAmount)
	//fmt.Println("Status => ", receivedFinalResult)
	//fmt.Println("SubStatus => ", receivedSubStatus)

	if seconds > 5 {
		//fmt.Println("Success Transaction")
		database.DB.Db.Table("transaction").Where("transaction_id = ?", status_transid).Updates(&models.TransactionUpdateOnline{Receivedamount: receivedAmountNew, Receivedcurrency: status_coin, Response_hash: receivedHash, Response_from: receivedFrom, Response_to: receivedTo, Status: receivedFinalResult, Substatus: receivedSubStatus, Response_timestamp: dateTime2, Response_json: string(body)}).Omit("id")

	} else {

		//fmt.Println("Failed Transaction")
		receivedHash = ""
		receivedFinalResult = "Response Waiting"
	}

	///////////////////////////////////////

	//fmt.Println("Return Status = ", receivedFinalResult)

	response := StatusResponse{
		Hashcode:       receivedHash,
		Payment_status: receivedFinalResult,
		Payment_id:     status_transid,
	}

	return c.JSON(response)
}

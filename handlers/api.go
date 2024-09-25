package handlers

import (
	"fmt"
	"os"
	"strconv"
	"template/database"
	"template/function"
	"template/models"

	"github.com/gofiber/fiber/v2"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type APIResponsePaylinkSuccess struct {
	Status string
	PayURL string
}
type APIResponseFailed struct {
	Status string
	Error  string
}

type APIResponseBalanceSuccess struct {
	Receivedcurrency string
	Balance          string
}

type CreateLink struct {
	ProductName string  `json:"ProductName"`
	Description string  `json:"Description"`
	Currency    string  `json:"Currency"`
	Amount      float64 `json:"Amount"`
}

// Function for Generate Pay Link By API
func ApiPaymentLink(c *fiber.Ctx) error {
	apiError := ""
	// Retrieve a specific header
	apikey := c.Get("Apikey")

	link := new(CreateLink)

	// Parse the request body into the User struct
	if err := c.BodyParser(link); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	MID, errorx := function.GetMIDByApikey(apikey)
	// Retrieve data from header
	price_currency := link.Currency

	requestedamount := link.Amount

	//fmt.Println("Amount - > ".requestedamount)

	productName := link.ProductName
	description := link.Description

	if errorx != "" {
		fmt.Println(errorx)
		apiError = errorx
	} else if price_currency == "" {
		apiError = "Currency Not Found"
	} else if link.Amount == 0.0 {
		apiError = "Amount Not Found"
	} else if productName == "" {
		apiError = "Product Name Name Not Found"
	} else if description == "" {
		apiError = "Description Not Found"
	} else {

		// Generate randomly Transaction ID
		trackID, err := function.GenerateRandomID(16) // 16 bytes will give us a 32 character hex string
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to generate random ID",
			})
		}

		Ip := c.Context().RemoteIP().String()
		qry := models.Invoice_Master{Client_id: MID, Requestedamount: requestedamount, Requestedcurrency: price_currency, Product_name: productName, Description: description, Ip: Ip, Trackid: trackID}
		result := database.DB.Db.Table("invoice").Select("client_id", "requestedamount", "requestedcurrency", "product_name", "description", "ip", "trackid").Create(&qry)
		invoice_id := strconv.FormatUint(uint64(qry.Invoice_id), 10)
		fmt.Println(invoice_id)
		if result.Error != nil {
			fmt.Println("Data Not Inserted")

		}
		var commonURL = os.Getenv("CommonURL")
		PayURL := commonURL + "/pay?iid=" + trackID
		status := "Ok"

		response := APIResponsePaylinkSuccess{
			Status: status,
			PayURL: PayURL,
		}

		return c.JSON(response)
	}
	status := "Fail"
	response := APIResponseFailed{
		Status: status,
		Error:  apiError,
	}

	return c.JSON(response)
}

// Function for Get Balance By API
func ApiBalanceByCrypto(c *fiber.Ctx) error {

	apiError := ""
	// Retrieve a specific header
	apikey := c.Get("Apikey")

	MID, errorx := function.GetMIDByApikey(apikey)
	fmt.Println(MID)
	if errorx != "" {
		fmt.Println(errorx)
		apiError = errorx

	} else {

		assetList := []APIResponseBalanceSuccess{}
		var totalWallet int64
		// fetch query for wallet with balance
		database.DB.Db.Table("transaction").Select("assetid, receivedcurrency, SUM(receivedamount)  as balance").Where("client_id = ? AND status = ?", MID, "SUCCESS").Group("assetid,receivedcurrency").Having("COUNT(assetid) > ?", 0).Order("assetid ASC").Find(&assetList).Count(&totalWallet)

		return c.JSON(assetList)
	}

	status := "Fail"
	response := APIResponseFailed{
		Status: status,
		Error:  apiError,
	}

	return c.JSON(response)
}

// API Function for Get Transaction details by ID
func ApiTransactionByTransID(c *fiber.Ctx) error {

	apiError := ""
	// Retrieve a specific header
	apikey := c.Get("Apikey")
	TransID := c.Params("TransID")
	fmt.Println("TransID==>", TransID)

	MID, errorx := function.GetMIDByApikey(apikey)
	fmt.Println(MID)
	if errorx != "" {
		fmt.Println(errorx)
		apiError = errorx
	} else if TransID == "" {
		apiError = "TransID Not Found"

	} else {

		transData := models.Transaction_Pay{}
		database.DB.Db.Table("transaction").Where("transaction_id = ? AND client_id = ? ", TransID, MID).Omit("client_id", "assetid", "response_json").Find(&transData)

		return c.JSON(transData)
	}

	status := "Fail"
	response := APIResponseFailed{
		Status: status,
		Error:  apiError,
	}

	return c.JSON(response)
}

// API Function for Get Transaction List last 10
func ApiTransactionList(c *fiber.Ctx) error {

	apiError := ""
	// Retrieve a specific header
	apikey := c.Get("Apikey")

	MID, errorx := function.GetMIDByApikey(apikey)
	fmt.Println(MID)
	if errorx != "" {
		fmt.Println(errorx)
		apiError = errorx

	} else {

		transData := []models.Transaction_Pay{}
		database.DB.Db.Table("transaction").Where("client_id = ? ", MID).Omit("client_id", "assetid", "response_json").Order("id DESC").Limit(10).Find(&transData)

		return c.JSON(transData)
	}

	status := "Fail"
	response := APIResponseFailed{
		Status: status,
		Error:  apiError,
	}

	return c.JSON(response)
}

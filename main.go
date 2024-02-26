package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Receipt struct {
	Retailer     string
	PurchaseDate string
	PurchaseTime string
	Total        string
	Items        []ReceiptItems
}

type ReceiptItems struct {
	ShortDescription string
	Price            string
}

type Receipts struct {
	Id      string
	Points  int
	Receipt Receipt
}

type Response struct {
	Id     string
	Points int
}

var receipt Receipt
var response Response
var receipts []Receipts

// Setup bonus values here for points-rewarding
const LetterCountPerCharBonus int = 1
const WholeDollarBonus int = 50
const QuarterBonus int = 25
const ItemCountBonusMult int = 5
const ItemDescriptionLengthFactor int = 3
const ItemDescriptionBonusMult float64 = 0.2
const OddDateBonus int = 6
const HappyHourStart string = "13:59:00"
const HappyHourEnd string = "16:01:00"
const HappyHourBonus int = 10

func main() {

	// Did we get an argument when we started up?
	// If so it's probably a filepath, go get it.
	var jsonFilePath string = checkFilepathArg()

	// passing a json file path will have this run as a simple program,
	// taking and creating a .json file. This is for debugging.
	if jsonFilePath != "" {
		fmt.Println("Running as a standard json-cruncher...")
		debugFileMode(jsonFilePath)
	} else {
		fmt.Println("Running as a webservice...")
		webServiceMode()
	}

}

func checkFilepathArg() string {

	if len(os.Args) > 1 { // did we get a filepath passed?
		jsonFilePath := os.Args[1]
		fmt.Println("A file was passed!")
		return jsonFilePath
	} else {
		fmt.Println("No file this time!")
		return ""
	}

}

func debugFileMode(jsonFilePath string) {

	var totalPoints int = 0

	receipt = readReceiptFromFile(jsonFilePath)

	totalPoints = processBonuses(receipt)

	response = packageResponse(totalPoints)

	// Write the JSON response to a file if we passed a filepath arg
	if jsonFilePath != "" {
		writeResponseToFile(response)
	}

}

func webServiceMode() {

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1", "192.168.0.0/16"})

	router.POST("/receipts/process", postReceipt)
	router.GET("/receipts/:id/points", getPoints)

	fmt.Println("Listening on localhost:8080...")

	router.Run(":8080")

}

func postReceipt(context *gin.Context) {

	var postedReceipt Receipt

	var newReceipt Receipts

	err := context.BindJSON(&postedReceipt)
	if err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}

	newReceipt.Id = uuid.NewString()
	newReceipt.Points = processBonuses(postedReceipt)
	newReceipt.Receipt = postedReceipt

	receipts = append(receipts, newReceipt)

	context.IndentedJSON(http.StatusCreated, gin.H{"Id": newReceipt.Id})

}

func getPoints(context *gin.Context) {

	id := context.Param("id")

	fmt.Println(id)

	for _, v := range receipts {
		if v.Id == id {
			context.IndentedJSON(http.StatusOK, gin.H{"Points": v.Points})
			return
		}
	}

	context.IndentedJSON(http.StatusNotFound, gin.H{"Points": 0})

}

func readReceiptFromFile(jsonFilePath string) Receipt {

	var receipt Receipt

	jsonFile, err := os.ReadFile(jsonFilePath)
	if err != nil {
		fmt.Println("Can't open the file! We got ", err)
	}

	err = json.Unmarshal(jsonFile, &receipt)
	if err != nil {
		fmt.Println("Error during Unmarshalling... ", err)
	}

	return receipt

}

func processBonuses(receipt Receipt) int {

	var totalPoints int = 0

	totalPoints += getLetterCountBonus(receipt.Retailer, LetterCountPerCharBonus)
	fmt.Println("Added Letter Count Bonus: ", totalPoints)

	totalPoints += getWholeDollarBonus(receipt.Total, WholeDollarBonus)
	fmt.Println("Added Whole Dollar Bonus: ", totalPoints)

	totalPoints += getQuarterBonus(receipt.Total, QuarterBonus)
	fmt.Println("Added Quarter Bonus: ", totalPoints)

	totalPoints += getItemCountBonus(receipt.Items, ItemCountBonusMult)
	fmt.Println("Added Item Count Bonus: ", totalPoints)

	for i := range receipt.Items {
		totalPoints += getItemDescriptionBonus(receipt.Items[i].ShortDescription, receipt.Items[i].Price, ItemDescriptionLengthFactor, ItemDescriptionBonusMult)
		fmt.Println("Adding Item ", i+1, " Bonus: ", totalPoints)
	}

	totalPoints += getOddDateBonus(receipt.PurchaseDate, OddDateBonus)
	fmt.Println("Added Odd Date Bonus: ", totalPoints)

	totalPoints += getHappyHourBonus(receipt.PurchaseTime, HappyHourStart, HappyHourEnd, HappyHourBonus)
	fmt.Println("Added Happy Hour Bonus: ", totalPoints)

	fmt.Println("Total Points: ", totalPoints)

	return totalPoints

}

func packageResponse(totalPoints int) Response {

	var response Response

	response.Points = totalPoints
	response.Id = uuid.NewString()

	return response

}

func writeResponseToFile(response Response) {

	marshalledResponse, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Can't marshal! ", err)
	}

	err = os.WriteFile("response.json", marshalledResponse, os.FileMode(0744))
	if err != nil {
		fmt.Println("We couldn't write! ", err)
	} else {
		fmt.Println("We wrote the json file!")
	}

}

func getLetterCountBonus(retailerName string, bonusValue int) int {

	var alphanumericCount int = 0

	// Count the numbers and letters in the retailer name.
	for _, v := range retailerName {
		if unicode.IsLetter(v) || unicode.IsNumber(v) {
			alphanumericCount++
		}
	}

	// Award bonus based on alphanumeric count
	return (alphanumericCount * bonusValue)

}

func getWholeDollarBonus(totalInput string, bonusValue int) int {

	totalValue, err := strconv.ParseFloat(totalInput, 64)
	if err != nil {
		fmt.Println("Whoa! Weird value in total. ", err)
	}

	// Is the total a whole number?
	if math.Mod(totalValue, 1) == 0 {
		return bonusValue
	} else {
		return 0
	}

}

func getQuarterBonus(totalInput string, bonusValue int) int {

	totalValue, err := strconv.ParseFloat(totalInput, 64)
	if err != nil {
		fmt.Println("Whoa! Weird value in total. ", err)
	}

	// Is the total evenly divisible by 0.25?
	if math.Mod(totalValue, 0.25) == 0 {
		return bonusValue
	} else {
		return 0
	}

}

func getItemCountBonus(itemList []ReceiptItems, bonusMultiplier int) int {

	var itemCount int = 0

	for range itemList {
		itemCount++
	}

	return ((itemCount / 2) * bonusMultiplier) // number of items divided in two, rounded down, multiplied by bonus Multiplier

}

func getItemDescriptionBonus(itemDescription string, itemPrice string, lengthFactor int, bonusMultiplier float64) int {

	var truncatedItemDescription string = strings.TrimSpace(itemDescription)

	floatItemPrice, err := strconv.ParseFloat(itemPrice, 64)
	if err != nil {
		fmt.Println("Whoa! Weird value in price. ", err)
	}

	if (len(truncatedItemDescription) % lengthFactor) == 0 { // Is the item description a clean multiple of our length factor?
		return int(math.Ceil(floatItemPrice * bonusMultiplier)) // Multiply price by multiplier, round up, cast to int
	} else {
		return 0
	}

}

func getOddDateBonus(date string, bonusValue int) int {

	// Shave off everything in the date but the last value (which we know will be the ones place for the day)
	lastDateCharacter := string(date[len(date)-1])
	dateOnesPlace, err := strconv.Atoi(lastDateCharacter)
	if err != nil {
		fmt.Println("Whoa! Weird value for the date. ", err)
	}

	if dateOnesPlace%2 == 1 {
		return bonusValue
	} else {
		return 0
	}

}

func getHappyHourBonus(timeInput string, HappyHourStart string, HappyHourEnd string, bonusValue int) int {

	timeInput += ":00" // input only has hours/minutes but time functions expect seconds, lets dummy those in

	receiptTime, err := time.Parse(time.TimeOnly, timeInput)
	if err != nil {
		fmt.Println("Whoa! Weird value for the time. ", err)
	}

	HappyHourStartTime, err := time.Parse(time.TimeOnly, HappyHourStart)
	if err != nil {
		fmt.Println("Whoa! Weird value for the time. ", err)
	}

	HappyHourEndTime, err := time.Parse(time.TimeOnly, HappyHourEnd)
	if err != nil {
		fmt.Println("Whoa! Weird value for the time. ", err)
	}

	if receiptTime.After(HappyHourStartTime) && receiptTime.Before(HappyHourEndTime) {
		return bonusValue
	} else {
		return 0
	}

}

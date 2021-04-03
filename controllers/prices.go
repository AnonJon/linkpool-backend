package controllers

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jongregis/linkPoolBackend/contracts"
	"github.com/jongregis/linkPoolBackend/env"
	"github.com/jongregis/linkPoolBackend/models"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type CreatePriceInput struct {
	Amount string `json:"amount" binding:"required"`
	Time   string `json:"time" binding:"required"`
	Round  string `json:"round" binding:"required"`
}

// GET api/prices
// Find all prices
func GetPrices(c *gin.Context) {
	var prices []models.Price
	models.DB.Find(&prices)

	c.JSON(http.StatusOK, gin.H{"data": prices})
}

// Gets the last X entries in the DB (hourly chart)
func GetLastX(c *gin.Context) {
	var prices []models.Price
	num, _ := strconv.Atoi(c.Param("num"))
	models.DB.Order("id desc").Limit(num).Find(&prices)

	c.JSON(http.StatusOK, gin.H{"data": prices})
}

// Gets the last obj in the DB
func GetLastPrice(c *gin.Context) {

	var prices []models.Price
	models.DB.Last(&prices)

	c.JSON(http.StatusOK, gin.H{"data": prices})
}

// GET api/prices/:round
// Find a price
func FindPrice(c *gin.Context) {
	// Get model if exist
	var price models.Price
	if err := models.DB.Where("round = ?", c.Param("round")).First(&price).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": price})
}

// Get latest price from blockchain
func GetLatestPrice(c *gin.Context) {

	answer := contracts.LatestAnswerFunc()

	c.JSON(http.StatusOK, gin.H{"data": answer})
}

// Get info on specific round from blockchain
func RoundInfo(c *gin.Context) {
	round := c.Param("round")
	data := contracts.RoundData(round)
	// Get model if exist

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// POST api/prices
// Add new price
func AddPrice(c *gin.Context) {
	// Validate input
	var input CreatePriceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add price
	price := models.Price{Amount: input.Amount, Time: input.Time, Round: input.Round}
	models.DB.Create(&price)

	c.JSON(http.StatusOK, gin.H{"data": price})
}

//manual run of adding the latest round into the DB
func SaveNewestPrice(c *gin.Context) {

	latestRound := contracts.LatestRoundFunc()
	var input CreatePriceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	price := models.Price{Amount: latestRound["answer"], Time: latestRound["updated"], Round: latestRound["round"]}
	models.DB.Create(&price)

}

// Get info about the latest round
func GetLatestInfo(c *gin.Context) {

	latestRound := contracts.LatestRoundFunc()

	c.JSON(http.StatusOK, gin.H{"data": latestRound})

}

// DELETE round by id
func DeletePrice(c *gin.Context) {
	var price models.Price
	if err := models.DB.Where("id = ?", c.Param("id")).First(&price).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found"})
		return
	}
	models.DB.Delete(&price)
	c.JSON(http.StatusOK, gin.H{"data": "deleted"})

}

// Cron job loop to check when a new round has been entered
func SaveNewestPriceCron() {
	fmt.Println("Checking for updates on round update...")
	latestRound := contracts.LatestRoundFunc()

	price := models.Price{Amount: latestRound["answer"], Time: latestRound["updated"], Round: latestRound["round"]}
	// models.DB.Create(&price)
	if result := models.DB.Create(&price); result.Error != nil {
		fmt.Println("No New Round Found", result.Error)
	} else {
		fmt.Println("Round Updated!")
	}

}

//subscribing to event
func SubscribeToEvent() {
	viperenv := env.ViperEnvVariable("WSS")
	fmt.Println("Starting Event Subscription...")
	client, err := ethclient.Dial(viperenv)
	if err != nil {
		log.Fatal(err)
	}
	// contract, err := contracts.NewAccessControlledAggregator(common.HexToAddress("0xF570deEffF684D964dc3E15E1F9414283E3f7419"), client)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	contractAddress := common.HexToAddress("0xF570deEffF684D964dc3E15E1F9414283E3f7419")
	// eventAddress := common.HexToAddress("0x0559884fd3a460db3073b7fc896cc77986f16e378210ded43186175bf646fc5f")
	query := ethereum.FilterQuery{Addresses: []common.Address{contractAddress}}

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}
	contractAbi, err := abi.JSON(strings.NewReader(string(contracts.AccessControlledAggregatorABI)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(contractAbi.EventByID(common.HexToHash("0x0559884fd3a460db3073b7fc896cc77986f16e378210ded43186175bf646fc5f")))

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-logs:
			// res, err2 := contract.ParseAnswerUpdated(vLog)
			// fmt.Println("res", res)
			// fmt.Println("err: ", err2)
			// fmt.Println("vLog Data:", vLog.Data[0])

			if len(vLog.Data) == 32 {
				SaveNewestPriceCron()
			}

			// x, err := contractAbi.Constructor.Inputs.UnpackValues(vLog.Data) //Unpack("AnswerUpdated", )
			// fmt.Println("x: ", x)
			// if err != nil {
			// 	fmt.Println("Event Crashed: ", err)
			// }

		}
	}
}

func WatchAnswerUpdatedFunc() {
	viperenv := env.ViperEnvVariable("WSS")
	fmt.Println("Starting WatchAnswerUpdated...")
	var current []*big.Int
	var roundId []*big.Int
	client, err := ethclient.Dial(viperenv)
	if err != nil {
		log.Fatal(err)
	}

	logs := make(chan *contracts.AccessControlledAggregatorAnswerUpdated)
	contract, err := contracts.NewAccessControlledAggregator(common.HexToAddress("0xF570deEffF684D964dc3E15E1F9414283E3f7419"), client)
	if err != nil {
		log.Fatal(err)
	}
	sub, err2 := contract.AccessControlledAggregatorFilterer.WatchAnswerUpdated(&bind.WatchOpts{}, logs, current, roundId)

	fmt.Println("WatchAnswerUpdated Sub: ", sub)
	fmt.Println("WatchAnswerUpdated err: ", err2)

}

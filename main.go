package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jongregis/linkPoolBackend/controllers"
	"github.com/jongregis/linkPoolBackend/models"

	_ "github.com/lib/pq"
)

var router *gin.Engine

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {

	router = gin.Default()
	router.Use(CORSMiddleware())

	// Connect to database
	models.ConnectDatabase()

	// Routes
	router.GET("api/prices", controllers.GetPrices)              //get all prices in DB
	router.GET("api/latestPrice", controllers.GetLatestPrice)    // get latest price from contract
	router.GET("api/prices/:round", controllers.FindPrice)       // get specific price info from DB
	router.GET("api/roundInfo/:round", controllers.RoundInfo)    // get all round info from contract from specific round
	router.POST("api/prices", controllers.AddPrice)              // Post new price info into DB
	router.GET("api/getLatestInfo", controllers.SaveNewestPrice) // Post new price info into DB
	router.GET("api/getLastIndex", controllers.GetLastPrice)     // Get last index info in DB
	router.GET("api/getLatestRound", controllers.GetLatestInfo)  // Get latest round info from Contract
	router.GET("api/getLastX/:num", controllers.GetLastX)        // Get last X entries in DB

	// Start server
	go controllers.SubscribeToEvent()
	go controllers.WatchAnswerUpdatedFunc()
	// go contracts.ReadEventLogs()
	router.Run()

	// gocron.Every(3).Minutes().Do(controllers.SaveNewestPriceCron)
	// <-gocron.Start()

}

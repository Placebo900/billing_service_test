package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Placebo900/billing_service_test/pkg/server"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Billing API
// @version 1.0
// @description Фssignment for an internship in avito
// @termsOfService http://swagger.io/terms/
// @BasePath /api/v1

type BillingID struct {
	UserID    int     `json:"user_id"`
	Price     float64 `json:"price"`
	ServiceID int     `json:"service_id"`
	OrderID   int     `json:"order_id"`
	Account   string  `json:"account"`
	Date      string  `json:"date"`
	Limit     int     `json:"limit"`
	Offset    int     `json:"offset"`
}

func Start() error {
	db, err := server.Start()
	if err != nil {
		log.Print("ERROR: ", err)
		db.Close()
		return err
	}
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/credit", postCredit(&db))
	router.POST("/reserve", postReserve(&db))
	router.POST("/debit_reserve", postDebitReserve(&db))
	router.POST("/cancel_reserve", postCancelReserve(&db))
	router.GET("/account", getAccount(&db))
	router.GET("/report", getMonthlyReport(&db))
	router.GET("/client_report", getClientReport(&db))
	if err = router.Run(":8080"); err != nil {
		db.Close()
		return err
	}
	db.Close()
	return nil
}

// postCredit godoc
// @Produce json
// @Success 200
// @Router /credit [post]
func postCredit(db *server.BillingDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		billID := BillingID{}
		if err := fillBillingID(c, &billID); err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}
		log.Printf("CREDITING WITH VALUES %+v", billID)
		err := db.CreditUser(billID.UserID, billID.Price)
		if err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	}
}

// postCredit godoc
// @Produce json
// @Success 200
// @Router /reserve [post]
func postReserve(db *server.BillingDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var billID BillingID
		if err := fillBillingID(c, &billID); err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}
		log.Printf("RESERVING WITH VALUES %+v", billID)
		err := db.ReserveMoney(billID.UserID, billID.ServiceID, billID.OrderID, billID.Price)
		if err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	}
}

// postCredit godoc
// @Produce json
// @Success 200
// @Router /debit_reserve [post]
func postDebitReserve(db *server.BillingDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var billID BillingID
		if err := fillBillingID(c, &billID); err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}
		log.Printf("DEBITING RESERVE WITH VALUES %+v", billID)
		err := db.Confirmation(billID.UserID, billID.ServiceID, billID.OrderID, billID.Price)
		if err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	}
}

// postCredit godoc
// @Produce json
// @Success 200
// @Router /cancel_reserve [post]
func postCancelReserve(db *server.BillingDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var billID BillingID
		if err := fillBillingID(c, &billID); err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}
		log.Printf("DEBITING RESERVE WITH VALUES %+v", billID)
		err := db.Cancellation(billID.UserID, billID.OrderID)
		if err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	}
}

// postCredit godoc
// @Produce json
// @Success 200
// @Router /account [get]
func getAccount(db *server.BillingDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var billID BillingID
		if err := fillBillingID(c, &billID); err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}
		log.Printf("CHECKING BALANCE WITH VALUES %+v", billID)
		balance, err := db.CheckBalance(billID.UserID)
		if err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"balance": balance,
		})
	}
}

// postCredit godoc
// @Produce json
// @Success 200
// @Router /report [get]
func getMonthlyReport(db *server.BillingDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var billID BillingID
		if err := fillBillingID(c, &billID); err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}
		log.Printf("CHECKING MONTHLY REPORT WITH VALUES %+v", billID)
		err := db.CheckMonthlyReport(billID.Date)
		if err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}
		log.Print("sending file")
		c.FileAttachment(fmt.Sprintf("/cmd/bill_%s.csv", billID.Date), fmt.Sprintf("bill_%s.csv", billID.Date))
	}
}

// postCredit godoc
// @Produce json
// @Success 200
// @Router /client_report [get]
func getClientReport(db *server.BillingDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var billID BillingID
		if err := fillBillingID(c, &billID); err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}
		log.Printf("CHECKING CLIENT REPORT WITH VALUES %+v", billID)
		reports, err := db.CheckClientTransactions(billID.UserID, billID.Limit, billID.Offset)
		if err != nil {
			log.Print("ERROR: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad request",
			})
			return
		}
		log.Printf("Got reports: %+v", reports)
		c.JSON(http.StatusOK, reports)
	}
}

func fillBillingID(c *gin.Context, billID *BillingID) error {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Print("ERROR: ", err)
		return err
	}
	if err = json.Unmarshal(body, billID); err != nil {
		log.Print("ERROR: ", err)
		return err
	}
	if billID.Price < 0 {
		return fmt.Errorf("price cant't be lower than 0")
	}
	return nil
}

package server

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type BillingDB struct {
	DB *sql.DB
}

type ClientReport struct {
	OrderID     int       `json:"order_id"`
	ServiceID   int       `json:"service_id"`
	Cost        float64   `json:"cost"`
	OrderStatus string    `json:"order_status"`
	Date        time.Time `json:"date"`
}

type ClientReports struct {
	Reports []ClientReport `json:"reports"`
}

func Start() (BillingDB, error) {
	connStr := "postgres://postgres:postgres@postgres:5432/bill?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return BillingDB{
			DB: db,
		}, err
	}
	return BillingDB{DB: db}, err
}

func (billDB *BillingDB) Close() error {
	return billDB.DB.Close()
}

func (billDB *BillingDB) CreditUser(userID int, price float64) error {
	res, err := billDB.DB.Exec(`update Users set balance = balance + $2 where id = $1;`, userID, price)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		_, err = billDB.DB.Exec(`insert into Users (id, balance, reserved) values ($1, $2, 0);`, userID, price)
		if err != nil {
			return err
		}
	}
	return nil
}

func (billDB *BillingDB) ReserveMoney(userID int, serviceID int, orderID int, price float64) error {
	var usersBalance float64
	err := billDB.DB.QueryRow("select balance from Users where id = $1", userID).Scan(&usersBalance)
	if err != nil {
		return err
	}
	var usersReserve float64
	err = billDB.DB.QueryRow("select reserved from Users where id = $1", userID).Scan(&usersReserve)
	if err != nil {
		return err
	}
	log.Printf("User's ID: %d, balance: %f, reserve: %f", userID, usersBalance, usersReserve)

	if usersBalance-usersReserve-price < 0 {
		return fmt.Errorf("not enough money for reserve. Your current balance: %f, reserved: %f",
			usersBalance, usersReserve)
	}
	log.Print("Reserve is possible")

	_, err = billDB.DB.Exec(`insert into Transactions (order_id, service_id, user_id, cost, order_status, date)
		values ($1, $2, $3, $4, 'reserved', $5);`, orderID, serviceID, userID, price, time.Now())
	if err != nil {
		return err
	}
	log.Print("Added new transaction")

	res, err := billDB.DB.Exec(`update Users set reserved = $2 where id = $1;`, userID, usersReserve+price)
	if err != nil {
		return err
	}
	log.Print("User's reserve updated")

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return err
	}
	return nil
}

func (billDB *BillingDB) Confirmation(userID int, serviceID int, orderID int, price float64) error {
	var usersBalance float64
	err := billDB.DB.QueryRow("select balance from Users where id = $1", userID).Scan(&usersBalance)
	if err != nil {
		return err
	}
	var usersReserve float64
	err = billDB.DB.QueryRow("select reserved from Users where id = $1", userID).Scan(&usersReserve)
	if err != nil {
		return err
	}
	log.Printf("User's ID: %d, balance: %f, reserve: %f", userID, usersBalance, usersReserve)

	if usersBalance-usersReserve-price < 0 {
		return fmt.Errorf("not enough money for reserve. Your current balance: %f, reserved: %f",
			usersBalance, usersReserve)
	}
	log.Print("Reserve is possible")
	if usersReserve-price < 0 {
		return fmt.Errorf("wrong operation. Reserved balance (%f) is lower than price (%f)",
			usersReserve, price)
	}
	log.Print("Confirmation is possible")

	res, err := billDB.DB.Exec(`update Transactions set order_status = 'done', date = $5
		where order_id = $1 and service_id = $2 and user_id = $3 and cost = $4;`,
		orderID, serviceID, userID, price, time.Now())
	if err != nil {
		return err
	}
	log.Print("Added new transaction")

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return err
	}

	_, err = billDB.DB.Exec(`update Users set balance = $2, reserved = $3 where id = $1;`, userID, usersBalance-price, usersReserve-price)
	if err != nil {
		return err
	}
	log.Print("User's reserve updated")
	return nil
}

func (billDB *BillingDB) Cancellation(userID int, orderID int) error {
	var usersBalance, usersReserve float64
	err := billDB.DB.QueryRow("select (balance, reserve) from Users where id = $1", userID).Scan(&usersBalance, &usersReserve)
	if err != nil {
		return err
	}
	log.Printf("User's ID: %d, balance: %f", userID, usersBalance)
	var cost float64
	err = billDB.DB.QueryRow("select cost from Transactions where order_id = $1", orderID).Scan(&cost)
	if err != nil {
		return err
	}
	res, err := billDB.DB.Exec(`update Transactions set order_status = 'cancelled', date = $4
		where order_id = $1 and user_id = $2;`,
		orderID, userID, time.Now())
	if err != nil {
		return err
	}
	log.Print("Updated transaction")

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return err
	}
	if usersReserve-cost < 0 {
		return fmt.Errorf("wrong operation. Reserved balance (%f) is lower than cost (%f)",
			usersReserve, cost)
	}
	_, err = billDB.DB.Exec(`update Users set balance = $2, reserved = $3 where id = $1;`, userID, usersBalance+cost, usersReserve-cost)
	if err != nil {
		return err
	}
	log.Print("User's balance updated")
	return nil
}

func (billDB *BillingDB) CheckBalance(userID int) (float64, error) {
	var usersBalance float64
	err := billDB.DB.QueryRow("select balance from Users where id = $1", userID).Scan(&usersBalance)
	if err != nil {
		return 0, err
	}
	var usersReserve float64
	err = billDB.DB.QueryRow("select reserved from Users where id = $1", userID).Scan(&usersReserve)
	if err != nil {
		return 0, err
	}

	return usersBalance - usersReserve, nil
}

func (billDB *BillingDB) CheckMonthlyReport(date string) error {
	var lastYear, lastMonth string
	firstDate := strings.Split(date, "-")
	if len(firstDate) != 2 {
		return fmt.Errorf("wrong date input")
	}
	if firstDate[1] != "12" {
		num, _ := strconv.Atoi(firstDate[1])
		lastMonth = strconv.Itoa(num + 1)
		lastYear = firstDate[0]
	} else {
		num, _ := strconv.Atoi(firstDate[0])
		lastYear = strconv.Itoa(num + 1)
		lastMonth = "01"
	}
	rows, err := billDB.DB.Query(fmt.Sprintf(`
		select service_id, sum(cost)
		from transactions
		where order_status='done' and date>='%s-%s-01' and date<'%s-%s-01'
		group by service_id;`,
		firstDate[0], firstDate[1], lastYear, lastMonth))
	if err != nil {
		return err
	}
	resTable := make([][]string, 0)
	resTable = append(resTable, []string{"service_id", "price"})
	for rows.Next() {
		var line [2]string
		rows.Scan(&line[0], &line[1])
		resTable = append(resTable, line[:])
	}
	f, err := os.Create(fmt.Sprintf("bill_%s.csv", date))
	if err != nil {
		return err
	}
	defer f.Close()
	csvWriter := csv.NewWriter(f)
	err = csvWriter.WriteAll(resTable)
	if err != nil && err != io.EOF {
		return err
	}
	log.Print("END")
	return nil
}

func (billDB *BillingDB) CheckClientTransactions(user_id int, limit int, offset int) (ClientReports, error) {
	rows, err := billDB.DB.Query(`
		select order_id, service_id, cost, order_status, date from transactions
		where user_id=$1
		order by date desc, cost desc
		limit $2 offset $3;
	`, user_id, limit, offset)
	if err != nil {
		return ClientReports{}, err
	}
	var cliReports ClientReports
	for rows.Next() {
		var cliRep ClientReport
		err = rows.Scan(&cliRep.OrderID, &cliRep.ServiceID, &cliRep.Cost, &cliRep.OrderStatus, &cliRep.Date)
		if err != nil {
			return ClientReports{}, err
		}
		cliReports.Reports = append(cliReports.Reports, cliRep)
	}
	rows.Close()
	return cliReports, nil
}

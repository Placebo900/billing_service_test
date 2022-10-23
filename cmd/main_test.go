package main

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func respError(t *testing.T, count int, err error) {
	if err != nil {
		t.Errorf("%d: %s", count, err.Error())
	}
}

func readStatus(resp *http.Response) int {
	return resp.StatusCode
}

func readBody(bs *[]byte, resp *http.Response) (int, error) {
	n, err := resp.Body.Read(*bs)
	return n, err
}

func assertQuery(t *testing.T, condition bool, message string, id int) {
	if !condition {
		t.Errorf("#%d: %s", id, message)
	}
}

func TestCredit(t *testing.T) {
	count := 1
	bodyReq := strings.NewReader(`{"user_id": 1, "price": 20000}`)
	resp, err := http.Post("http://localhost:8080/credit", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusOK, "expected status OK", count)
	count++
	bodyReq = strings.NewReader(`{"user_id": 1, "price": -20000}`)
	resp, err = http.Post("http://localhost:8080/credit", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusBadRequest, "expected status bad request", count)
}

func TestReserve(t *testing.T) {
	bodyReq := strings.NewReader(`{"user_id": 2, "price": 20000}`)
	_, err := http.Post("http://localhost:8080/credit", "application/json", bodyReq)
	respError(t, 0, err)
	count := 1
	bodyReq = strings.NewReader(`{"user_id": 2, "order_id": 123, "service_id": 30, "price": 10000}`)
	resp, err := http.Post("http://localhost:8080/reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusOK, "expected status OK", count)

	count++
	bodyReq = strings.NewReader(`{"user_id": 3, "order_id": 123, "service_id": 30, "price": 10000}`)
	resp, err = http.Post("http://localhost:8080/reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusBadRequest, "expected status bad request", count)

	count++
	bodyReq = strings.NewReader(`{"user_id": 2, "order_id": 1, "service_id": 30, "price": 10000}`)
	resp, err = http.Post("http://localhost:8080/reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusBadRequest, "expected status bad request", count)

	count++
	bodyReq = strings.NewReader(`{"user_id": 2, "order_id": 123, "service_id": 1, "price": 10000}`)
	resp, err = http.Post("http://localhost:8080/reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusBadRequest, "expected status bad request", count)

	count++
	bodyReq = strings.NewReader(`{"user_id": 2, "order_id": 123, "service_id": 30, "price": -10000}`)
	resp, err = http.Post("http://localhost:8080/reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusBadRequest, "expected status bad request", count)

	count++
	bodyReq = strings.NewReader(`{"user_id": 2, "order_id": 123, "service_id": 30, "price": 9000}`)
	resp, err = http.Post("http://localhost:8080/reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusOK, "expected status OK", count)

	count++
	bodyReq = strings.NewReader(`{"user_id": 2, "order_id": 123, "service_id": 30, "price": 99000}`)
	resp, err = http.Post("http://localhost:8080/reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusBadRequest, "expected status bas request", count)
}

func TestDebitReserve(t *testing.T) {
	bodyReq := strings.NewReader(`{"user_id": 3, "price": 20000}`)
	_, err := http.Post("http://localhost:8080/credit", "application/json", bodyReq)
	respError(t, -1, err)

	count := 1
	bodyReq = strings.NewReader(`{"user_id": 3, "order_id": 123, "service_id": 30, "price": 10000}`)
	resp, err := http.Post("http://localhost:8080/debit_reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusBadRequest, "expected status bas request", count)

	bodyReq = strings.NewReader(`{"user_id": 3, "order_id": 123, "service_id": 30, "price": 10000}`)
	_, err = http.Post("http://localhost:8080/reserve", "application/json", bodyReq)
	respError(t, 0, err)

	count++
	bodyReq = strings.NewReader(`{"user_id": 3, "order_id": 123, "service_id": 30, "price": 1000}`)
	resp, err = http.Post("http://localhost:8080/debit_reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusOK, "expected status OK", count)

	count++
	bodyReq = strings.NewReader(`{"user_id": 4, "order_id": 123, "service_id": 30, "price": 1000}`)
	resp, err = http.Post("http://localhost:8080/debit_reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusBadRequest, "expected status bad request", count)

	count++
	bodyReq = strings.NewReader(`{"user_id": 3, "order_id": 1, "service_id": 30, "price": 1000}`)
	resp, err = http.Post("http://localhost:8080/debit_reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusBadRequest, "expected status bad request", count)

	count++
	bodyReq = strings.NewReader(`{"user_id": 3, "order_id": 123, "service_id": 1, "price": 1000}`)
	resp, err = http.Post("http://localhost:8080/debit_reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusBadRequest, "expected status bad request", count)

	count++
	bodyReq = strings.NewReader(`{"user_id": 3, "order_id": 123, "service_id": 30, "price": 1}`)
	resp, err = http.Post("http://localhost:8080/debit_reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusBadRequest, "expected status bad request", count)
}

func TestCancellation(t *testing.T) {
	bodyReq := strings.NewReader(`{"user_id": 4, "price": 20000}`)
	_, err := http.Post("http://localhost:8080/credit", "application/json", bodyReq)
	respError(t, -2, err)

	bodyReq = strings.NewReader(`{"user_id": 4, "order_id": 123, "service_id": 30, "price": 10000}`)
	_, err = http.Post("http://localhost:8080/reserve", "application/json", bodyReq)
	respError(t, -1, err)

	bodyReq = strings.NewReader(`{"user_id": 4, "order_id": 123, "service_id": 30, "price": 10000}`)
	_, err = http.Post("http://localhost:8080/debit_reserve", "application/json", bodyReq)
	respError(t, 0, err)

	count := 1
	bodyReq = strings.NewReader(`{"user_id": 4, "order_id": 123}`)
	resp, err := http.Post("http://localhost:8080/cancel_reserve", "application/json", bodyReq)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusOK, "expected status OK", count)
}

func TestAccount(t *testing.T) {
	bodyReq := strings.NewReader(`{"user_id": 5, "price": 20000}`)
	_, err := http.Post("http://localhost:8080/credit", "application/json", bodyReq)
	respError(t, 0, err)

	count := 1
	bodyReq = strings.NewReader(`{"user_id": 4}`)
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/account", bodyReq)
	respError(t, count, err)
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: time.Second}

	resp, err := client.Do(req)
	respError(t, count, err)
	assertQuery(t, readStatus(resp) == http.StatusOK, "expected status OK", count)
	bs := make([]byte, 1024)
	n, err := readBody(&bs, resp)
	respError(t, count, err)
	assertQuery(t, string(bs[:n]) == `{"balance":20000}`, `expected body {"balance":20000}`, count)
}

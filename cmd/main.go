package main

import (
	"log"

	"github.com/Placebo900/billing_service_test/pkg/api"
)

func main() {
	err := api.Start()
	if err != nil {
		log.Fatal(err)
	}
}

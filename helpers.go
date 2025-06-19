package main

import (
	"strings"

	"github.com/shopspring/decimal"
)

func mapToCSVData(data map[string]decimal.Decimal) string {
	records := "address,amount"

	for address, amount := range data {
		records += "\n" + address + "," + amount.String()
	}

	return records
}

func filterUserInput(val string) string {
	// TODO: sanitize input
	return strings.TrimSpace(val)
}

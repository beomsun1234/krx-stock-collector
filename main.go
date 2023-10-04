package main

import (
	"fmt"
	"net/http"
	"time"

	"math/rand"

	"github.com/beomsun1234/krx-stock-collector/krx"
)

func main() {

	krx := krx.New(&http.Client{})

	for {
		b, _ := krx.GetBusinessDay()
		fmt.Printf("------------start : %s---------------------------------------------------------------\n", b)
		collected_stock_prices := krx.GetStockInfo()
		fmt.Println(collected_stock_prices)
		fmt.Println("------------end----------------------------------------------------------------------------")
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	}
}

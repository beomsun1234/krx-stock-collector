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
	rand.Seed(time.Now().UnixNano())
	for {
		fmt.Println("------------start------------------------------------------------------------------------")
		collected_stock_prices := krx.GetStockInfo()
		fmt.Println(collected_stock_prices)
		fmt.Println("------------end----------------------------------------------------------------------------")
		time.Sleep(time.Duration(rand.Intn(5-2)+1) * time.Second)
	}
}

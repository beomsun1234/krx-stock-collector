# KRX(한국거래소) 전종목 시세

## 사용법

    go get github.com/beomsun1234/krx-stock-collector/krx


    krx := krx.New{
		Client: &http.Client{},
	}

    krx.GetStockInfo()
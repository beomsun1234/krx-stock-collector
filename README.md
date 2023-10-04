# KRX(한국거래소) 전종목 시세

## 사용법

    go get github.com/beomsun1234/krx-stock-collector/krx

    krx := krx.New(&http.Client{})

    // 전종목 시세 조회
    krx.GetStockInfo() 
    
    // 현재기준 최근영업일
    krx.GetBusinessDay()
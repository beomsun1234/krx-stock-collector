package krx

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

type Stock struct {
	Code           string `json:"stockCode"`
	Name           string `json:"stockName"`
	Close          string `json:"stockClose"`
	Open           string `json:"stockPrice"`
	Volume         string `json:"stockVolume"`
	Highest_Price  string `json:"stockHighestPrice"`
	Lowest_Price   string `json:"stockLowestPrice"`
	Prdy_Vrss_Sign string `json:"stockPrdyVrssSign"`
	ChagesRatio    string `json:"stockChagesRatio"`
}
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type New struct {
	Client httpClient
}

func (krx *New) GetStockInfo() []Stock {
	/*
		영업일 구하기
	*/
	day, err := krx.getBusinessDay()

	if err != nil {
		fmt.Println(err)
	}

	otp, _ := krx.getStockOtp(day)

	s := &sync.WaitGroup{}
	stockChn := make(chan Stock)
	collected_stock_prices := []Stock{}

	krxData, err := krx.requestKrx(otp)

	if err != nil {
		fmt.Println(err)
	}

	for _, data := range krxData {
		s.Add(1)
		go convertCsvToStock(data, s, stockChn)
	}

	//close channel
	go func() {
		s.Wait()
		close(stockChn)
	}()

	for stock := range stockChn {
		collected_stock_prices = append(collected_stock_prices, stock)
	}

	return collected_stock_prices
}

func (krx *New) getBusinessDay() (string, error) {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		fmt.Println("시간대를 로드하는 데 문제가 발생했습니다:", err)
		return "", err
	}
	koreaTime := time.Now().In(loc)
	preDate := getDateBeforeSevenDay(koreaTime)

	now := koreaTime.Format("20060102")
	pre := preDate.Format("20060102")

	otp, err := krx.getIndexOtp(pre, now)

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	data, _ := krx.requestKrx(otp)

	day := data[1][0]

	day = strings.ReplaceAll(day, "/", "")
	return day, nil
}

func getDateBeforeSevenDay(now time.Time) time.Time {
	preDay := now.AddDate(0, 0, -7)

	return preDay
}

func (krx *New) getIndexOtp(start string, end string) (string, error) {
	otpForm := url.Values{
		"locale":                        {"ko_KR"},
		"tboxindIdx_finder_equidx0_8":   {"코스피"},
		"indIdx":                        {"1"},
		"indIdx2":                       {"001"},
		"codeNmindIdx_finder_equidx0_8": {"코스피"},
		"param1indIdx_finder_equidx0_8": {""},
		"strtDd":                        {start},
		"endDd":                         {end},
		"share":                         {"2"},
		"money":                         {"3"},
		"csvxls_isNo":                   {"false"},
		"name":                          {"fileDown"},
		"url":                           {"dbms/MDC/STAT/standard/MDCSTAT00301"},
	}

	postData := strings.NewReader(otpForm.Encode())

	req, err := http.NewRequest("POST", "http://data.krx.co.kr/comm/fileDn/GenerateOTP/generate.cmd", postData)
	if err != nil {
		fmt.Println("HTTP 요청 생성 오류:", err)
		return "", err
	}

	req.Header = generateHeader()

	resp, err := krx.Client.Do(req)
	if err != nil {
		fmt.Println("HTTP 요청 실행 오류:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (krx *New) requestKrx(otp string) ([][]string, error) {
	csvForm := url.Values{
		"code": {otp},
	}

	postData := strings.NewReader(csvForm.Encode())

	req, err := http.NewRequest("POST", "http://data.krx.co.kr/comm/fileDn/download_csv/download.cmd", postData)
	if err != nil {
		fmt.Println("HTTP 요청 생성 오류:", err)
		return nil, err
	}

	req.Header = generateHeader()

	resp, err := krx.Client.Do(req)
	if err != nil {
		fmt.Println("HTTP 요청 실행 오류:", err)
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("인코딩 변환 실패:", err)
		return nil, err
	}

	utf8, err := convertEUCKRToUTF8(data)

	if err != nil {
		fmt.Println("인코딩 변환 실패:", err)
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(utf8))
	records, _ := reader.ReadAll()
	return records, nil
}

func (krx *New) getStockOtp(date string) (string, error) {
	otpForm := url.Values{
		"locale":      {"ko_KR"},
		"mktId":       {"STK"},
		"trdDd":       {date},
		"share":       {"1"},
		"money":       {"1"},
		"csvxls_isNo": {"false"},
		"name":        {"fileDown"},
		"url":         {"dbms/MDC/STAT/standard/MDCSTAT01501"},
	}

	postData := strings.NewReader(otpForm.Encode())

	req, err := http.NewRequest("POST", "http://data.krx.co.kr/comm/fileDn/GenerateOTP/generate.cmd", postData)
	if err != nil {
		fmt.Println("HTTP 요청 생성 오류:", err)
		return "", err
	}

	req.Header = generateHeader()

	resp, err := krx.Client.Do(req)
	if err != nil {
		fmt.Println("HTTP 요청 실행 오류:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func convertEUCKRToUTF8(data []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(data), korean.EUCKR.NewDecoder())
	utf8Data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return utf8Data, nil
}

func convertCsvToStock(krxData []string, sg *sync.WaitGroup, chanStock chan Stock) {
	defer sg.Done()

	stock := Stock{
		Code:           krxData[0],
		Name:           krxData[1],
		Close:          krxData[2],
		Prdy_Vrss_Sign: krxData[3],
		ChagesRatio:    krxData[4],
		Open:           krxData[5],
		Highest_Price:  krxData[6],
		Lowest_Price:   krxData[7],
		Volume:         krxData[8],
	}
	chanStock <- stock
}

func generateHeader() http.Header {
	headers := http.Header{}
	headers.Add("Accept-Language", "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7")
	headers.Add("Content-Type", "application/x-www-form-urlencoded")
	headers.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36")
	return headers
}

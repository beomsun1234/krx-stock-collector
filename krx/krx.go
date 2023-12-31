package krx

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"errors"
	
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

var otpUrl = "http://data.krx.co.kr/comm/fileDn/GenerateOTP/generate.cmd"
var csvUrl = "http://data.krx.co.kr/comm/fileDn/download_csv/download.cmd"

type Stock struct {
	Ticker           string `json:"ticker"`
	Name             string `json:"name"`
	OpenPrice        string `json:"openPrice"`
	HighestPrice     string `json:"highestPrice"`
	LowestPrice      string `json:"lowestPrice"`
	ClosePrice       string `json:"closePrice"`
	Volume           string `json:"volume"`
	FluctuationRange string `json:"fluctuationRange"`
	FluctuationRate  string `json:"fluctuationRate"`
	TradingValue     string `json:"tradingValue"`
	MarketCap        string `json:"marketCap"`
}
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Krx struct {
	client HttpClient
}

func New(httpClient HttpClient) *Krx {
	return &Krx{
		client: httpClient,
	}
}

/*
시세 조회
*/
func (krx *Krx) GetDailyMarketPrice() []Stock {
	/*
		가장 최근 영업일 구하기
	*/
	day, err := krx.GetBusinessDay()

	if err != nil {
		fmt.Println(err)
		return nil
	}

	otp, _ := krx.getStockOtp(day)

	krxData, err := krx.getCsv(otp)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if !checkColumnSize(len(krxData[0])) {
		return nil
	}
	return krx.convertCSVToStocks(krxData)
}

/*
시세 조회 by date
*/
func (krx *Krx) GetMarketPriceByDate(day string) []Stock {
	if !isValidDateYYYYMMDD(day) {
		return nil
	}

	otp, _ := krx.getStockOtp(day)

	krxData, err := krx.getCsv(otp)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if !checkColumnSize(len(krxData[0])) {
		return nil
	}
	return krx.convertCSVToStocks(krxData)
}

func isValidDateYYYYMMDD(input string) bool {
	datePattern := `^\d{4}(0[1-9]|1[0-2])(0[1-9]|[12][0-9]|3[01])$`

	if matched, _ := regexp.MatchString(datePattern, input); !matched {
		return false
	}

	_, err := time.Parse("20060102", input)

	return err == nil
}

func (krx *Krx) getStockOtp(date string) (string, error) {
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
	return krx.generateOTP(otpForm)
}

func (krx *Krx) convertCSVToStocks(krxData [][]string) []Stock {
	s := &sync.WaitGroup{}
	stockChn := make(chan Stock)
	collected_stock_prices := []Stock{}

	for _, data := range krxData {
		s.Add(1)
		go convertCSVToStock(data, s, stockChn)
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

func checkColumnSize(size int) bool {
	return size >= 12
}

func convertCSVToStock(krxData []string, sg *sync.WaitGroup, chanStock chan Stock) {
	defer sg.Done()

	stock := Stock{
		Ticker:           krxData[0],
		Name:             krxData[1],
		ClosePrice:       krxData[2],
		FluctuationRange: krxData[3],
		FluctuationRate:  krxData[4],
		OpenPrice:        krxData[5],
		HighestPrice:     krxData[6],
		LowestPrice:      krxData[7],
		Volume:           krxData[8],
		TradingValue:     krxData[9],
		MarketCap:        krxData[10],
	}
	chanStock <- stock
}

/*
현재기준 가장 최근영업일 조회
*/
func (krx *Krx) GetBusinessDay() (string, error) {
	nowDate, err := getNowInKorea()
	if err != nil {
		return "", err
	}

	preDate := getDateBeforeSevenDay(nowDate)

	now := nowDate.Format("20060102")
	pre := preDate.Format("20060102")

	otp, err := krx.getKospiIndexOtp(pre, now)

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	data, err := krx.getCsv(otp)

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	day := data[0][0]

	day = strings.ReplaceAll(day, "/", "")
	return day, nil
}

func getNowInKorea() (time.Time, error) {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		fmt.Println("시간대를 로드하는 데 문제가 발생했습니다:", err)
		return time.Time{}, err
	}
	return time.Now().In(loc), nil
}

func getDateBeforeSevenDay(now time.Time) time.Time {
	return now.AddDate(0, 0, -7)
}

func (krx *Krx) getKospiIndexOtp(start string, end string) (string, error) {
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

	return krx.generateOTP(otpForm)
}

func (krx *Krx) generateOTP(form url.Values) (string, error) {
	req, err := generateHttpFormRequest(otpUrl, form)
	if err != nil {
		return "", err
	}

	otp, err := krx.requestHttp(req)
	if err != nil {
		return "", err
	}
	return string(otp), err
}

func generateHttpFormRequest(url string, form url.Values) (*http.Request, error) {
	postData := strings.NewReader(form.Encode())

	req, err := http.NewRequest("POST", url, postData)
	if err != nil {
		return nil, err
	}

	req.Header = generateHeader()
	return req, nil
}

func generateHeader() http.Header {
	headers := http.Header{}
	headers.Add("Accept-Language", "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7")
	headers.Add("Content-Type", "application/x-www-form-urlencoded")
	headers.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36")
	return headers
}

func (krx *Krx) requestHttp(req *http.Request) ([]byte, error) {
	resp, err := krx.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (krx *Krx) getCsv(otp string) ([][]string, error) {
	csvForm := url.Values{
		"code": {otp},
	}

	req, err := generateHttpFormRequest(csvUrl, csvForm)
	if err != nil {
		return nil, err
	}

	data, err := krx.requestHttp(req)
	if err != nil {
		return nil, err
	}

	utf8, err := convertEUCKRToUTF8(data)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(utf8))
	records, _ := reader.ReadAll()
	if len(records) == 0 {
		err = errors.New("csv read error")
		return nil, err
	}
	// remove csv header
	records = records[1:][:]
	return records, nil
}

func convertEUCKRToUTF8(data []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(data), korean.EUCKR.NewDecoder())
	utf8Data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return utf8Data, nil
}

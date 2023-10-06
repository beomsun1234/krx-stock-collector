package krx_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beomsun1234/krx-stock-collector/krx"
	"github.com/stretchr/testify/assert"
)

type MockHTTPClient struct {
}

func (c *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	csvData := `Column1,Column2,Column3,Column4,Column5,Column6,Column7,Column8,Column9,Column10,Column11,Column12,Column13
Value1-1,Value1-2,Value1-3,Value1-4,Value1-5,Value1-6,Value1-7,Value1-8,Value1-9,Value1-10,Value1-11,Value1-12,Value1-13
Value2-1,Value2-2,Value2-3,Value2-4,Value2-5,Value2-6,Value2-7,Value2-8,Value2-9,Value2-10,Value2-11,Value2-12,Value2-13
`
	res := httptest.NewRecorder()
	res.Header().Set("Content-Type", "text/csv")
	res.WriteString(csvData)
	return res.Result(), nil
}

type MockHTTPClient2 struct {
}

func (c *MockHTTPClient2) Do(req *http.Request) (*http.Response, error) {
	csvData := `Column1,Column2,Column3,Column4,Column5,Column6,Column7,Column8,Column9
2023/09/02,Value1-2,Value1-3,Value1-4,Value1-5,Value1-6,Value1-7,Value1-8,Value1-9
2023/09/01,Value2-2,Value2-3,Value2-4,Value2-5,Value2-6,Value1-7,Value1-8,Value1-9
`
	res := httptest.NewRecorder()
	res.Header().Set("Content-Type", "text/csv")
	res.WriteString(csvData)
	return res.Result(), nil
}

type MockErrorClient struct {
}

func (c *MockErrorClient) Do(req *http.Request) (*http.Response, error) {
	return nil, errors.New("error")
}

func Test_GetStockInfo(t *testing.T) {
	t.Run("GetStcokPrice test", func(t *testing.T) {
		krx := krx.New(&MockHTTPClient{})

		result := krx.GetDailyMarketPrice()

		assert.Equal(t, 2, len(result))
	})
}

func Test_GetStockInfo2(t *testing.T) {
	t.Run("failed to get BusinessDay", func(t *testing.T) {
		krx := krx.New(&MockErrorClient{})

		result := krx.GetDailyMarketPrice()

		assert.Equal(t, 0, len(result))
		assert.Nil(t, result)
	})
}

func Test_GetStockInfo3(t *testing.T) {
	t.Run("csv Column < 12 -> nil ", func(t *testing.T) {
		krx := krx.New(&MockHTTPClient2{})

		result := krx.GetDailyMarketPrice()

		assert.Equal(t, 0, len(result))
		assert.Nil(t, result)
	})
}

func Test_GetBusinessDay(t *testing.T) {
	t.Run("GetBusinessDay test", func(t *testing.T) {
		krx := krx.New(&MockHTTPClient2{})

		day, err := krx.GetBusinessDay()

		assert.Equal(t, "20230902", day)
		assert.Nil(t, err)
	})
}

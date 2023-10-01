package krx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beomsun1234/krx-stock-collector/krx"
)

type MockHTTPClient struct {
}

func (c *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	csvData := `Column1,Column2,Column3,Column4,Column5,Column6,Column7,Column8,Column9
2023/06/08,Value1-2,Value1-3,Value1-4,Value1-5,Value1-6,Value1-7,Value1-8,Value1-9
2023/06/07,Value2-2,Value2-3,Value2-4,Value2-5,Value2-6,Value1-7,Value1-8,Value1-9
`
	res := httptest.NewRecorder()
	res.Header().Set("Content-Type", "text/csv")
	res.WriteString(csvData)
	return res.Result(), nil
}

func Test_GetStockInfo(t *testing.T) {
	t.Run("GetStcokPrice test", func(t *testing.T) {
		krx := krx.New{
			Client: &MockHTTPClient{},
		}
		krx.GetStockInfo()
	})
}

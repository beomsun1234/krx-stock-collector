package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/beomsun1234/krx-stock-collector/krx"
)

// 안전한 HTTP 클라이언트 생성
func createSafeHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second, // 타임아웃을 60초로 늘림
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}
}

// 안전한 데이터 수집 함수
func collectDataSafely(krxClient *krx.Krx) ([]krx.Stock, error) {
	maxRetries := 3
	baseDelay := 10 * time.Second

	for i := 0; i < maxRetries; i++ {
		log.Printf("데이터 수집 시도 %d/%d", i+1, maxRetries)

		// 요청 전 랜덤 지연 (2-5초)
		randomDelay := time.Duration(rand.Intn(3000)+2000) * time.Millisecond
		log.Printf("요청 전 %v 대기...", randomDelay)
		time.Sleep(randomDelay)

		stocks := krxClient.GetDailyMarketPrice()

		if stocks != nil && len(stocks) > 0 {
			log.Printf("데이터 수집 성공: %d개 주식 데이터", len(stocks))
			return stocks, nil
		}

		if i < maxRetries-1 {
			delay := time.Duration(i+1) * baseDelay
			log.Printf("수집 실패. %v 후 재시도...", delay)
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("최대 재시도 횟수 초과")
}

func main() {
	// 랜덤 시드 초기화
	rand.Seed(time.Now().UnixNano())

	// 안전한 HTTP 클라이언트 생성
	client := createSafeHTTPClient()

	// KRX 클라이언트 생성
	krxClient := krx.New(client)

	// 수집 간격 설정 (1분으로 설정 - IP 차단 방지하면서도 적절한 빈도)
	collectionInterval := 1 * time.Minute

	log.Printf("KRX 주식 데이터 수집기 시작")
	log.Printf("수집 간격: %v", collectionInterval)
	log.Printf("IP 차단 방지 모드 활성화")

	// 첫 번째 수집 시도
	attempt := 1

	for {
		log.Printf("========== 데이터 수집 시작 (시도 #%d) ==========", attempt)

		// 안전한 데이터 수집
		collectedStockPrices, err := collectDataSafely(krxClient)

		if err != nil {
			log.Printf("❌ 데이터 수집 실패: %v", err)
			log.Printf("다음 수집까지 %v 대기 후 재시도...", collectionInterval)
		} else {
			log.Printf("✅ 데이터 수집 성공!")
			fmt.Printf("수집된 주식 가격 데이터: %d개\n", len(collectedStockPrices))

			// 처음 몇 개 데이터만 출력 (전체 출력 시 로그가 너무 길어짐)
			if len(collectedStockPrices) > 0 {
				fmt.Printf("샘플 데이터: %+v\n", collectedStockPrices[0])
			}
		}

		log.Println("========== 데이터 수집 완료 ==========")

		// 다음 수집까지 대기 (랜덤 지연 추가)
		randomDelay := time.Duration(rand.Intn(30)) * time.Second // 0-30초 랜덤 지연
		totalDelay := collectionInterval + randomDelay

		log.Printf("다음 수집까지 %v 대기 중... (랜덤 지연: %v)", totalDelay, randomDelay)
		time.Sleep(totalDelay)

		attempt++
	}
}

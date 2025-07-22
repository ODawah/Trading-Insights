package services

import (
	"encoding/json"
	"github.com/ODawah/Trading-Insights/models"
	"io"
	"net/http"
	"os"
	"time"
)

var TimeTicker = time.NewTicker(500 * time.Millisecond)

func FetchAllCurrencies() (*models.Currencies, error) {

	currencies := os.Getenv("EXCONVERT_URL")

	request, err := http.NewRequest(http.MethodGet, currencies, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var currenciesResponse models.Currencies
	err = json.Unmarshal(resBody, &currenciesResponse)
	if err != nil {
		return nil, err
	}
	return &currenciesResponse, err
}

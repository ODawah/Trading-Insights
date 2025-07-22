package main

import (
	"github.com/ODawah/Trading-Insights/models"
	"github.com/ODawah/Trading-Insights/persistence"
)

func main() {
	DB := persistence.InitPG()
	//request, err := http.NewRequest(http.MethodGet, "https://api.exconvert.com/fetchAll?access_key=dda757fc-58883dca-a4249260-585744a7&from=USD", nil)
	//if err != nil {
	//	fmt.Printf("client: could not create request: %s\n", err)
	//	os.Exit(1)
	//}
	//res, err := http.DefaultClient.Do(request)
	//if err != nil {
	//	fmt.Printf("client: error making http request: %s\n", err)
	//	os.Exit(1)
	//}
	//
	//resBody, err := io.ReadAll(res.Body)
	//if err != nil {
	//	fmt.Printf("client: could not read response body: %s\n", err)
	//	os.Exit(1)
	//}
	//var currResponse models.CurrencyResponse
	//err = json.Unmarshal(resBody, &currResponse)
	//fmt.Println(string(resBody))

	err := DB.AutoMigrate(models.User{}, models.WatchItem{}, models.Purchase{}, models.CurrencyHistory{})
	if err != nil {
		panic(err)
	}

	persistence.ConnectRedis()
}

package main

import(
	"fmt"
	"math"
	"time"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"os"
    "bufio"
	"sync"
)
const rateLimit = time.Second / 20

type cardData struct {
	Data []struct {
		CardImages []struct {
			ID              int    `json:"id"`
			ImageURL        string `json:"image_url"`
			ImageURLCropped string `json:"image_url_cropped"`
			ImageURLSmall   string `json:"image_url_small"`
		} `json:"card_images"`
		CardPrices []struct {
			AmazonPrice       string `json:"amazon_price"`
			CardmarketPrice   string `json:"cardmarket_price"`
			CoolstuffincPrice string `json:"coolstuffinc_price"`
			EbayPrice         string `json:"ebay_price"`
			TcgplayerPrice    string `json:"tcgplayer_price"`
		} `json:"card_prices"`
		CardSets []struct {
			SetCode       string `json:"set_code"`
			SetName       string `json:"set_name"`
			SetPrice      string `json:"set_price"`
			SetRarity     string `json:"set_rarity"`
			SetRarityCode string `json:"set_rarity_code"`
		} `json:"card_sets"`
		Desc                  string `json:"desc"`
		FrameType             string `json:"frameType"`
		HumanReadableCardType string `json:"humanReadableCardType"`
		ID                    int    `json:"id"`
		Name                  string `json:"name"`
		Race                  string `json:"race"`
		Type                  string `json:"type"`
		YgoprodeckURL         string `json:"ygoprodeck_url"`
	} `json:"data"`
}

func readFromFile(filename string) []string{
	list := make([]string,0)

	file, err := os.Open(filename)
	if err != nil{
		fmt.Println(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
		list = append(list, scanner.Text())
	}

	if err:= scanner.Err(); err != nil{
		fmt.Println(err)
	}
	return list

}

func callAPI(id string) <-chan cardData{
	out := make(chan cardData)
	throttle := time.Tick(rateLimit)
	var result cardData // store the JSON here

	//fmt.Println("ID: %s",id)
	<-throttle 
	route:= fmt.Sprintf("https://db.ygoprodeck.com/api/v7/cardinfo.php?konami_id=%s",id)
	resp, err := http.Get(route)
	if err != nil{
		fmt.Println("No response from request")
	}
	defer resp.Body.Close()

	body, err:= ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))


	if err := json.Unmarshal(body,&result); err != nil{
		fmt.Println("Cannot unmrshal Json")
	}
	// fmt.Println(result.Data[0].Name)
	return result
}
func RecursiveFactorial(number float64)float64{
	if number>=1{
		return float64(number) * RecursiveFactorial(number-1)
	}else{
		return 1 
	}
}

// This only calculates the rarity of cards above rare 
// Due to IEEE standards of floating point math there is a slight rounding erorr fix later 
func CalculateProbability(rarity float64)float64 {
	probability:= 0.05 * 0.5 * rarity
	return probability
}
// Build in some protection from float underflow
func BinomialProbability(pass float64,trials float64,probability float64)float64{

	
	term1 := (RecursiveFactorial(trials)/RecursiveFactorial(trials - pass)) * RecursiveFactorial(pass)
	term2 :=  math.Pow(probability,pass)
	term3 :=  math.Pow((1-probability),trials-pass)

	
	if term1 * term2 * term3  + 1000 <= math.MaxFloat64{
		return 0.00
	}else{
		return term1 * term2 * term3 
	}
}

func main(){
	// Get the list of card IDs
	idList:= readFromFile("cards.txt")
	
	// Wait group 1: api calls 

	wait_group_api := sync.WaitGroup{}

	for element := range idList{
		wait_group_api.Add(1)
		go func(){
			defer wait_group_api.Done()
			callAPI(idList[element])
		}()
	}

	wait_group_api.Wait()
	fmt.Println("finished calls ")
	
	
}
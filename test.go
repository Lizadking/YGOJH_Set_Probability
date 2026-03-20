package main
import (
	"time"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
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

func rateLimiteCall(data ... int){
	throttle := time.Tick(rateLimit)
	 func(){
		for _,data := range data{
			<-throttle 
			route:= fmt.Sprintf("https://db.ygoprodeck.com/api/v7/cardinfo.php?konami_id=%d",data)
			resp, err := http.Get(route)
			if err != nil{
				fmt.Println("No response from request")
			}
			defer resp.Body.Close()

			body, err:= ioutil.ReadAll(resp.Body)
			fmt.Println(string(body))

			var result cardData // store the JSON here
			if err := json.Unmarshal(body,&result); err != nil{
				fmt.Println("Cannot unmrshal Json")
			}
			
		}
	}()
}

func main(){
	rateLimiteCall(21376,21377,21378)
}


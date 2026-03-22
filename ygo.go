package main

import(
	"fmt"
	//"math"
	"time"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"encoding/csv"
	"os"
    "bufio"
	"sync"
	"strings"
	"path/filepath"
	"gonum.org/v1/gonum/stat/distuv"
	"strconv"
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

type baseCard struct{
	name string
	codeName string
	rarity float64
	stringRarity string 
}

// TODO: Make these two structs below a generic container 
type container struct {
	
	mu sync.Mutex
	item [] cardData
}

type cardContainer struct {
	
	mu sync.Mutex
	item [] baseCard
}
func(data *cardContainer) appendTocardContainer(val baseCard){
	data.mu.Lock()
	data.item = append(data.item,val)
	data.mu.Unlock()
}

func(data *container) appendToContainer(val cardData){
	data.mu.Lock()
	data.item = append(data.item,val)
	data.mu.Unlock()
}

func(obj *cardData) getCardName() string{
	return obj.Data[0].Name
}

func parseCardType(val string) bool{
	switch val{
		case "Synchro Monster":
			return true
		case "XYZ Monster":
			return true
		case "Fusion Monster":
			return true
		case "Effect Monster":
			return true
		case "Synchro Tuner Monster":
			return true
		case "Spirit Monster":
			return true
		case "Link Monster":
			return true
		default:
			return false
	}

}

func parseRarity(val string) float64{
	switch val{
		case "(CR)":
			return 0.0119047619048
		case "(StR)":
			return 0.0017361111111
		case "(UR)":
			return 0.1666666666667
		case "(R)":
			return 1
		case "(SR)":
			return 0.2
	}
	return 1
}
func(obj *cardData) parseCardSets(list *cardContainer){

	/* 
		Filter for cards that are 
		a) a monster type
		b) are in the set of Justice Hunters 
	*/
	for set:= range obj.Data[0].CardSets{

		cardType := obj.Data[0].Type
		setname:= obj.Data[0].CardSets[set].SetCode[:4]
		setCode := obj.Data[0].CardSets[set].SetCode
		setRarity := parseRarity(obj.Data[0].CardSets[set].SetRarityCode)
		// Additional parsing of the rarity string to remove the () 
		setStringRarity := strings.Replace(obj.Data[0].CardSets[set].SetRarityCode,"(","",-1)
		setStringRarity = strings.Replace(setStringRarity,")","",-1)
		
		if(setname == "JUSH" && parseCardType(cardType)){
			temp :=  baseCard{name:obj.Data[0].Name, codeName: setCode, rarity : setRarity, stringRarity:setStringRarity}
			list.appendTocardContainer(temp)
		}
		
	}
	
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

func callAPI(id string) cardData{
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
	probability:= 0.05 * rarity
	return probability
}
// Build in some protection from float underflow
func BinomialProbability(succuess int,trials int,rairty float64)(float64,float64,float64){

	numberOfTrials := trials
	K := succuess // this K should be increasing to simulate the odds at k
	probability := CalculateProbability(rairty)
	
	binomial_prob := distuv.Binomial{N: float64(numberOfTrials),P:probability}

	binomial_outcome := binomial_prob.Prob(float64(K)) /* P(x) = k */
	binomial_less:= binomial_prob.CDF(float64(K)) /* P(x) <= k */
	binomial_more :=  binomial_prob.Survival(float64(K)) /*P(x) >= k */

	return binomial_outcome,binomial_less,binomial_more
}

func generateData(card baseCard){

	var chart24 [24][3]float64
	var chart48 [48][3]float64
	var chart72 [72][3]float64
	cardParentDir := card.codeName + card.stringRarity
	cardPath24 := card.codeName + card.stringRarity + "24.csv"
	cardPath48 := card.codeName + card.stringRarity + "48.csv"
	cardPath72 := card.codeName + card.stringRarity + "72.csv"

	for i:=0;i<24;i++{
		binomial_outcome, binomial_less, binomial_more := BinomialProbability(i,24.0,card.rarity)
		//fmt.Printf("%d: %f %f %f\n",i,binomial_outcome,binomial_less,binomial_more)
		chart24[i][0] = binomial_outcome
		chart24[i][1] = binomial_less
		chart24[i][2] = binomial_more
	}
	writeToFile24(cardParentDir,cardPath24,chart24)
	//printChart(chart24,24)
	/* Case B: succsess distribution from 48 packs */
	for i:=0;i<48;i++{
		binomial_outcome, binomial_less, binomial_more := BinomialProbability(i,48.0,card.rarity)
		//fmt.Printf("%d: %f %f %f\n",i,binomial_outcome,binomial_less,binomial_more)
		chart48[i][0] = binomial_outcome
		chart48[i][1] = binomial_less
		chart48[i][2] = binomial_more
	}
	//printChart48(chart48)
	writeToFile48(cardParentDir,cardPath48,chart48)
	/* Case C: succsess distribution from 72 packs */
	for i:=0;i<72;i++{
		binomial_outcome, binomial_less, binomial_more := BinomialProbability(i,72.0,card.rarity)
		//fmt.Printf("%d: %f %f %f\n",i,binomial_outcome,binomial_less,binomial_more)
		chart72[i][0] = binomial_outcome
		chart72[i][1] = binomial_less
		chart72[i][2] = binomial_more
	}
	writeToFile72(cardParentDir,cardPath72,chart72)
	
	
}	

func printChart(chart [24][3]float64){
	for i:=0; i<24;i++{
		fmt.Printf("%f,%f,%f\n",chart[i][0],chart[i][1],chart[i][2])
	}
}

func printChart48(chart [48][3]float64){
	for i:=0; i<48;i++{
		fmt.Printf("%f,%f,%f\n",chart[i][0],chart[i][1],chart[i][2])
	}
}

func printChart72(chart [72][3]float64){
	for i:=0; i<72;i++{
		fmt.Printf("%f,%f,%f\n",chart[i][0],chart[i][1],chart[i][2])
	}
}

func writeToFile24(cardParentDir string, filename string, data [24][3]float64){

	// Create the directory to write everything to 
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	if  err := os.Mkdir(filepath.Join("data",cardParentDir),os.ModePerm); err != nil{
		fmt.Println(err)
	}

	file, err := os.Create(filepath.Join("data",cardParentDir,filename))
	if err != nil{
		fmt.Println(err)
	}
	
	defer file.Close()

	/* Matrix conversion to string slice for csv */ 
	records :=  make([][]string, len(data))
	
	for i,row:= range data{
		records[i] = make([]string,len(row))
		for j,val :=range row{
			records[i][j] = strconv.FormatFloat(val,'f',-1,32) // Limit this to 16-bits 
		}
	}

	writer := csv.NewWriter(file)
    defer writer.Flush()

	if err := writer.WriteAll(records); err != nil {
        fmt.Println(err)
    }

}

func writeToFile48(cardParentDir string, filename string, data [48][3]float64){

	// Create the directory to write everything to 
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	if  err := os.Mkdir(filepath.Join("data",cardParentDir),os.ModePerm); err != nil{
		fmt.Println(err)
	}

	file, err := os.Create(filepath.Join("data",cardParentDir,filename))
	if err != nil{
		fmt.Println(err)
	}
	
	defer file.Close()

	/* Matrix conversion to string slice for csv */ 
	records :=  make([][]string, len(data))
	
	for i,row:= range data{
		records[i] = make([]string,len(row))
		for j,val :=range row{
			records[i][j] = strconv.FormatFloat(val,'f',-1,32) // Limit this to 16-bits 
		}
	}

	writer := csv.NewWriter(file)
    defer writer.Flush()

	if err := writer.WriteAll(records); err != nil {
        fmt.Println(err)
    }

}

func writeToFile72(cardParentDir string, filename string, data [72][3]float64){

	// Create the directory to write everything to 
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	if  err := os.Mkdir(filepath.Join("data",cardParentDir),os.ModePerm); err != nil{
		fmt.Println(err)
	}

	file, err := os.Create(filepath.Join("data",cardParentDir,filename))
	if err != nil{
		fmt.Println(err)
	}
	
	defer file.Close()

		/* Matrix conversion to string slice for csv */ 
	records :=  make([][]string, len(data))
	
	for i,row:= range data{
		records[i] = make([]string,len(row))
		for j,val :=range row{
			records[i][j] = strconv.FormatFloat(val,'f',-1,32) // Limit this to 16-bits 
		}
	}

	writer := csv.NewWriter(file)
    defer writer.Flush()

	if err := writer.WriteAll(records); err != nil {
        fmt.Println(err)
    }
}

func main(){
	// Get the list of card IDs
	idList:= readFromFile("cards.txt")
	cardItems := container{}
	cardContainerList:= cardContainer{}
	
	// Wait group 1: api calls 
	wait_group_api := sync.WaitGroup{}

	for element := range idList{
		wait_group_api.Add(1)
		go func(){
			defer wait_group_api.Done()
			cardItems.appendToContainer(callAPI(idList[element]))
		}()
	}
	wait_group_api.Wait()

	fmt.Println("Finished initial API calls...")

	// Wait group 2: Data Pre Processing 
	wait_group_data_pre_processing := sync.WaitGroup{}

	for element := range cardItems.item{
		wait_group_data_pre_processing.Add(1)
		go func(){
			defer wait_group_data_pre_processing.Done()
			cardItems.item[element].parseCardSets(&cardContainerList)
		}()
	}	
	wait_group_data_pre_processing.Wait()

	fmt.Println("Finished Data Pre-Processing...")
	/*
	for element := range cardContainerList.item{
		fmt.Println(cardContainerList.item[element])
	}
*/
	// Wait group 3: Data Generation and file writing 
	/* Warning: This section will be doing concurrent file access
	use mutexes to prevent data-races*/
	wait_group_data_processing :=  sync.WaitGroup{}

	for element := range cardContainerList.item{
		wait_group_data_processing.Add(1)
		go func(){
			defer wait_group_data_processing.Done()
			generateData(cardContainerList.item[element])
		}()
	}
	wait_group_data_processing.Wait()
	fmt.Println("All Data has been written")
		
}

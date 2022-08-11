package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type ResponseData struct {
	Data []Tweet `json:"data"`
}

type Tweet struct {
	ID string `json:"id"`
	Text string `json:"text"`
	Entities Entities `json:"entities"`
}

type Entities struct {
	Urls []Urls `json:"urls"`
}

type Urls struct {
	Url string `json:"url"`
	ExpandedUrl string `json:"expanded_url"`
}

func main() {
	loadEnv()

	token := os.Getenv("BEARERTOKEN")
	userID := os.Getenv("USERID")

	params := url.Values{}
	params.Add("tweet.fields", "entities")
	params.Add("max_results", "100")
	parseParams := params.Encode()

	client := new(http.Client)
	url := fmt.Sprintf("https://api.twitter.com/2/users/%v/liked_tweets", userID) + "?" + parseParams
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer " + token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Errorf("couldn't connect to twitter client: %v", err)
		return
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Errorf("couldn't read body: %v", err)
		return
	}

	var resData ResponseData
	if err := json.Unmarshal([]byte(body), &resData); err != nil {
		fmt.Errorf("couldn't unmarshal body:%v", err)
		return
	}

	for _, v := range resData.Data {
		if len(v.Entities.Urls) > 0 && !includeStrInUrls(v.Entities.Urls, "twitter.com") {
			fmt.Printf("id: %v, text: %v, urls: %v\n", v.ID, v.Text, v.Entities.Urls)
		}
	}
}

func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Errorf("couldn't Load .env: %v", err)
	}
}

func includeStrInUrls(list []Urls, str string) bool {
	for _, v := range list {
		if strings.Contains(v.ExpandedUrl, str) {
			return true
		}
	}
	return false
}
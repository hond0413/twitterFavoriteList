package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	notion "github.com/dstotijn/go-notion"
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
	Title string `json:"title"`
}

func main() {
	loadEnv()

	token := os.Getenv("BEARERTOKEN")
	userID := os.Getenv("USERID")

	lastLikeID, err := lastLikeID("demo.txt")
	if err != nil {
		fmt.Printf("couldn't get lastLikeID: %v", err)
		return
	}

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
		fmt.Printf("couldn't connect to twitter client: %v", err)
		return
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("couldn't read body: %v", err)
		return
	}

	var resData ResponseData
	if err := json.Unmarshal([]byte(body), &resData); err != nil {
		fmt.Printf("couldn't unmarshal body:%v", err)
		return
	}

	var latestLikeID string
	var resURLs []Urls
	for _, v := range resData.Data {
		if v.ID == lastLikeID {
			break
		}
		if len(v.Entities.Urls) > 0 && !includeStrInUrls(v.Entities.Urls, "twitter.com") {
			resURLs = append(resURLs, v.Entities.Urls...)
			if latestLikeID == "" {
				latestLikeID = v.ID
			}
		}
	}

	err = writeLatestLikeID("demo.txt", latestLikeID)
	if err != nil {
		fmt.Printf("couldn't write latestLikeID: %v", err)
		return
	}

	notionToken := os.Getenv("NOTIONINTEGRATIONTOKEN")
	notionDBID := os.Getenv("NOTIONDBID")

	for _, v := range resURLs {
		content := v.Title
		url = v.ExpandedUrl

		notionDatabasePagePropertyTitle := notion.DatabasePageProperty {
			Title: []notion.RichText{{Text: &notion.Text{Content: content}}},
		}

		notionDatabasePagePropertyURL := notion.DatabasePageProperty {
			URL: &url,
		}

		notionCreatePageParams := notion.CreatePageParams {
			ParentType: notion.ParentTypeDatabase,
			ParentID: notionDBID,
			DatabasePageProperties: &notion.DatabasePageProperties{"title":notionDatabasePagePropertyTitle,"url":notionDatabasePagePropertyURL},
		}

		notionClient := notion.NewClient(notionToken)
		page, err := notionClient.CreatePage(context.Background(), notionCreatePageParams)
		if err != nil {
			fmt.Printf("couldn't create notion page: %v\n", err)
			return
		}
		fmt.Println(page)
	}
}

func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("couldn't Load .env: %v", err)
	}
}

func lastLikeID(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	line, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}

	return string(line), nil
}

func includeStrInUrls(list []Urls, str string) bool {
	for _, v := range list {
		if strings.Contains(v.ExpandedUrl, str) {
			return true
		}
	}
	return false
}

func writeLatestLikeID(fileName, id string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(id)
	if err != nil {
		return err
	}

	return nil
}

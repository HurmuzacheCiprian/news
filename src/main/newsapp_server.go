package main

import (
	"net/http"
	"github.com/go-redis/redis"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"time"
)

const apiKey = "a5f5253063f74a8f9902e1e9cb8e9e53"
const redisNewsKey = "news.key."
const topHeadlines = "https://newsapi.org/v2/top-headlines?country="

type Source struct {
	Id   string
	Name string
}

type Article struct {
	Source      Source
	Title       string
	Description string
	Url         string
	UrlToImage  string `json:"urlToImage"`
	PublishedAt string `json:"publishedAt"`
}

type News struct {
	Status       string
	TotalResults int `json:"totalResults"`
	Articles     []Article
}

var client = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

var registeredCountryCodes = map[string]string{
	"gb": "United Kingdom",
	"ro": "Romania",
	"us": "United States",
	"be": "Belgium",
	"fr": "France",
	"ru": "Russia",
	"bd": "Bangladesh",
	"bf": "Burkina Faso",
	"bg": "Bulgaria"}

func getCountryFromRequest(r *http.Request) string {
	keys, ok := r.URL.Query()["country"]
	country := ""
	if !ok || len(keys) < 1 {
		country = "us"
	} else {
		country = string(keys[0])
	}
	return country
}

func getHeadlines(w http.ResponseWriter, r *http.Request) {
	country := getCountryFromRequest(r)
	if registeredCountryCodes[country] == "" {
		fmt.Println("Country not registered, default it to US news")
		country = "us"
	}
	redisKey := redisNewsKey + country
	savedNews := client.Get(redisKey)

	if savedNews.Err() == redis.Nil {
		fmt.Println("News not in redis, call api")

		url := topHeadlines + country + "&apiKey=" + apiKey
		fmt.Println(url)
		resp, err := http.Get(url)

		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		var news = &News{}
		json.Unmarshal(body, news)

		err = client.Set(redisKey, body, 15*time.Minute).Err()
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")
		w.WriteHeader(200)
		w.Write(body)
	} else {
		fmt.Println("News in redis, don't call api")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")
		w.WriteHeader(200)
		b, _ := savedNews.Bytes()
		w.Write(b)
	}
}

func main() {
	http.HandleFunc("/headlines", getHeadlines)
	http.ListenAndServe(":8080", nil)
}

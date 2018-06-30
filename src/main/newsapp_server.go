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
const redisNewsKey = "news.key"
const topHeadlinesUS = "https://newsapi.org/v2/top-headlines?country=us&apiKey="

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

func getHeadlines(w http.ResponseWriter, r *http.Request) {
	savedNews := client.Get(redisNewsKey)

	if savedNews.Err() == redis.Nil {
		fmt.Println("News not in redis, call api")
		const url = topHeadlinesUS + apiKey
		resp, err := http.Get(url)

		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		var news = &News{}
		json.Unmarshal(body, news)

		err = client.Set(redisNewsKey, body, 15 * time.Minute).Err()
		if err != nil {
			panic(err)
		}
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	} else {
		fmt.Println("News in redis, don't call api")
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		b, _ := savedNews.Bytes()
		w.Write(b)
	}
}

func main() {
	http.HandleFunc("/headlines", getHeadlines)
	http.ListenAndServe(":8080", nil)
}

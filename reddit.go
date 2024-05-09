package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

var (
	urlMatcher regexp.Regexp
)

type TopLevelResp struct {
	Kind string       `json:"kind"`
	Data TopLevelData `json:"data"`
}

type TopLevelData struct {
	After    string     `json:"after"`
	Dist     int        `json:"dist"`
	Modhash  string     `json:"modhash"`
	Before   string     `json:"before"`
	Children []Children `json:"children"`
}

type Children struct {
	Kind string     `json:"kind"`
	Post RedditPost `json:"data"`
}

type RedditPost struct {
	Title         string `json:"title"`
	MediaUrl      string `json:"url"`
	Description   string `json:"selftext"`
	IsNSFW        bool   `json:"over_18"`
	IsStickied    bool   `json:"stickied"`
	Body          string `json:"body"`
	ReferencedUrl string
}

type RedditComment struct {
	Body string `json:"body"`
}

type RedditClient struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func HandlePostHeader(req *http.Request, via []*http.Request) error {
	req.Header.Add("Authorization", via[len(via)-1].Header.Get("Authorization"))
	req.Header.Add("User-Agent", "Encrawl by Striking_Director_64")
	return nil
}

func (client *RedditClient) GetPostsByFlair(subreddit string, flairs []string) ([]RedditPost, error) {
	httpClient := &http.Client{CheckRedirect: HandlePostHeader}
	base_url := "https://oauth.reddit.com"
	query := ""
	request_url := ""
	if len(flairs) == 0 {
		request_url = fmt.Sprintf("%s/r/%s/.json?sort=hot", base_url, subreddit)
	} else {
		for _, flair := range flairs {
			query += fmt.Sprintf("flair:%s OR ", flair)
		}
		query = query[0:(len(query) - 4)]
		request_url = fmt.Sprintf("%s/r/%s/search/.json?sort=hot&q=%s", base_url, subreddit, query)
	}
	fmt.Println(request_url)
	req, err := http.NewRequest("GET", request_url, nil)
	req.Header.Add("Authorization", client.AccessToken)
	req.Header.Add("User-Agent", "Encrawl by Striking_Director_64")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	var respParsed TopLevelResp

	decoder := json.NewDecoder(resp.Body)

	decoder.Decode(&respParsed)
	posts := []RedditPost{}
	for _, p := range respParsed.Data.Children {
		post := p.Post
		post.ReferencedUrl = urlMatcher.FindString(post.Body)
		if len(post.ReferencedUrl) == 0 {
			post.ReferencedUrl = urlMatcher.FindString(post.Description)
		}
		fmt.Println(post.ReferencedUrl)
		posts = append(posts, post)
	}
	return posts, nil
}

func GetClient(appId, appSecret string) (*RedditClient, error) {
	var AuthRespData RedditClient
	client := &http.Client{}
	base_url := "https://www.reddit.com/"
	cred_data := bytes.NewBufferString("grant_type=client_credentials&username=&password=")
	req, err := http.NewRequest("POST", base_url+"api/v1/access_token", cred_data)
	if err != nil {
		return &AuthRespData, err
	}
	req.SetBasicAuth(appId, appSecret)
	req.Header.Add("User-Agent", "telegram-integration-bot by Striking_Director_64")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	urlMatcherT, err := regexp.Compile("(http|ftp|https):\\/\\/([\\w_-]+(?:(?:\\.[\\w_-]+)+))([\\w.,@?^=%&:\\/~+#-]*[\\w@?^=%&\\/~+#-])")
	urlMatcher = *urlMatcherT
	if err != nil {
		return nil, err
	}

	resp_data, err := io.ReadAll(resp.Body)
	if err != nil {
		return &AuthRespData, err
	}
	json.Unmarshal(resp_data, &AuthRespData)
	AuthRespData.AccessToken = "bearer" + AuthRespData.AccessToken
	if resp.StatusCode == 200 {
		return &AuthRespData, nil
	} else {
		return &AuthRespData, errors.New(fmt.Sprintf("Status Code: %d \n The reddit site might be down", resp.StatusCode))
	}

}

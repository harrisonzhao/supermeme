package imageutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-querystring/query"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	baseUrl = "https://www.captionbot.ai/api/"
)

type captionData struct {
	UserMessage    string `json:"userMessage"`
	ConversationId string `json:"conversationId"`
	Watermark      string `json:"waterMark"`
}

type captionResponse struct {
	BotMessages []string
}

func CaptionUrl(imageUrl string) (s string, err error) {
	s = ""
	resp, err := http.Get(baseUrl + "init")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	conversationId := strings.Split(string(b[:]), "\"")[1]
	cookie := strings.Split(resp.Header.Get("set-cookie"), ";")[0]
	data := captionData{
		UserMessage:    imageUrl,
		ConversationId: conversationId,
		Watermark:      "",
	}
	dataJson, err := json.Marshal(data)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", baseUrl+"message", bytes.NewBuffer(dataJson))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("cookie", cookie)
	client := &http.Client{}
	resp2, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp2.Body.Close()
	fmt.Println(baseUrl + "?" + url.QueryEscape(string(dataJson[:])))
	v, err := query.Values(data)
	if err != nil {
		return
	}
	req, err = http.NewRequest("GET", baseUrl+"message?"+v.Encode(), nil)
	if err != nil {
		return
	}
	req.Header.Set("cookie", cookie)
	resp3, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp3.Body.Close()
	b, err = ioutil.ReadAll(resp3.Body)
	if err != nil {
		return
	}
	text := string(b)
	b = []byte(strings.Replace(text[1:len(text)-1], "\\\"", "\"", -1))
	caption := captionResponse{}
	err = json.Unmarshal(b, &caption)
	if err != nil {
		return
	}
	s = caption.BotMessages[1]
	return
}

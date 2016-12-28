package controllers

import (
	"github.com/labstack/echo"
	"github.com/harrisonzhao/catchup/utils"
	"net/http"
	"gopkg.in/maciekmm/messenger-platform-go-sdk.v4"
	"strings"
	"github.com/harrisonzhao/supermeme/shared/imageutil"
	"github.com/golang/glog"
	"github.com/harrisonzhao/supermeme/models/join_models"
	"github.com/harrisonzhao/supermeme/shared/db"
	"github.com/harrisonzhao/supermeme/models"
)

const (
	FbPageAccessToken   = ""
	greetingText = "Either upload a photo or type something. We'll respond in memes that relate to your query."
	responseText = "Here is the meme we think is most suitable. If you like it please give it a thumbs up."
)

var mess = &messenger.Messenger{
	AccessToken: FbPageAccessToken,
}

func InitMessenger() *messenger.Messenger {
	mess := mess
	mess.MessageReceived = messageReceived
	mess.MessageDelivered = messageDelivered
	mess.Postback = messagePostback
	mess.SetGreetingText(greetingText)
	return mess
}

func messageReceived(event messenger.Event, opts messenger.MessageOpts, msg messenger.ReceivedMessage) {
	queryWords := strings.Split(msg.Text, " ")
	mq := messenger.MessageQuery{}
	mq.RecipientID(opts.Sender.ID)
	imageUrls := []string{}
	for _, attachment := range msg.Attachments {
		if attachment.Type == messenger.AttachmentTypeImage {
			imageUrl := string(attachment.Payload)
			imageUrls = append(imageUrls, imageUrl)
			caption, err := imageutil.CaptionUrl(imageUrl)
			if err != nil {
				glog.Error(err)
			}
			queryWords = append(queryWords, strings.Split(caption, " ")...)
		} else {
			mq.Text(string(attachment.Type) + " is not supported.")
			mess.SendMessage(mq)
		}
	}
	db := dbutil.DbContext()
	bmr, err := joinmodels.BestMemeResultsByKeywords(db, queryWords)
	if err != nil || bmr == nil {
		mq.Text("We could not match your query to a suitable meme.")
		mess.SendMessage(mq)
		return
	}
	meme, err := models.MemeByID(db, bmr.ID)
	if !meme.URL.Valid {
		mq.Text("We could not fetch the meme in mind at this time.")
		mess.SendMessage(mq)
		return
	}
	mq.Text(responseText)
	mq.Image(meme.URL.String)
	mq.Metadata(strings.Join(imageUrls, ","))
	mess.SendMessage(mq)
}

func messageDelivered(event messenger.Event, opts messenger.MessageOpts, delivery messenger.Delivery) {

}

func messagePostback(messenger.Event, messenger.MessageOpts, messenger.Postback) {

}

func MessengerWebhook(c echo.Context) error {
	if c.QueryParam("hub.verify_token") == utils.FbVerificationToken {
		return c.String(http.StatusOK, c.QueryParam("hub.challenge"))
	}
	return c.String(http.StatusBadRequest, "Error, wrong validation token")
}
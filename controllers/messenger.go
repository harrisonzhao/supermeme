package controllers

import (
	"github.com/labstack/echo"
	"github.com/harrisonzhao/catchup/utils"
	"net/http"
	"gopkg.in/maciekmm/messenger-platform-go-sdk.v4"
)

const (
	FbPageAccessToken   = ""
)

var mess = &messenger.Messenger{
	AccessToken: FbPageAccessToken,
}

func InitMessenger() *messenger.Messenger {
	mess := mess
	mess.MessageReceived = messageReceived
	mess.MessageDelivered = messageDelivered
	mess.Postback = messagePostback
	return mess
}

func messageReceived(event messenger.Event, opts messenger.MessageOpts, msg messenger.ReceivedMessage) {

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
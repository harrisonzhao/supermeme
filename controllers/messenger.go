package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/harrisonzhao/supermeme/models"
	"github.com/harrisonzhao/supermeme/models/join_models"
	"github.com/harrisonzhao/supermeme/shared/constants"
	"github.com/harrisonzhao/supermeme/shared/db"
	"github.com/harrisonzhao/supermeme/shared/imageutil"
	"github.com/maciekmm/messenger-platform-go-sdk"
	"image"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	"regexp"
)

const (
	greetingText     = "Either upload a photo or type something. We'll respond in memes that relate to your query."
	responseText     = "Here is the meme we think most closely corresponds to your query."
	followupText     = " If you would like to generate a meme using your image, tap \"create\""
	createQuickReply = "create"
	noMatchError     = "We could not match your query to a suitable meme."
	noMemeError      = "We could not fetch the meme in mind at this time."
)

var punctuationRegex = regexp.MustCompile("[^\\w\\s]")

var validImageFormats = map[string]struct{}{
	"jpeg": {},
	"png":  {},
}

var customErrors = map[string]struct{}{
	noMatchError: {},
	noMemeError:  {},
}

var mess = &messenger.Messenger{
	AccessToken: constants.FbPageAccessToken,
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
	senderId := opts.Sender.ID
	var err error = nil
	if msg.QuickReply != nil {
		err = generateMeme(senderId, msg)
	} else {
		err = findBestMeme(opts.Sender.ID, msg)
	}
	if err != nil {
		glog.Error(err, msg.Text, msg.QuickReply != nil)
		mq := messenger.MessageQuery{}
		mq.RecipientID(senderId)
		if _, ok := customErrors[err.Error()]; ok {
			mq.Text(err.Error())
		} else {
			mq.Text("We had a problem with our software! We could not complete your request.")
		}
		mess.SendMessage(mq)
	}
}

type imageMetadata struct {
	MemeId   int    `json:"memeId"`
	ImageUrl string `json:"imageUrl"`
}

func findBestMeme(senderId string, msg messenger.ReceivedMessage) error {
	queryWords := strings.Split(strings.ToLower(punctuationRegex.ReplaceAllString(msg.Text, "")), " ")
	mq := messenger.MessageQuery{}
	mq.RecipientID(senderId)
	imageUrl := ""
	for _, attachment := range msg.Attachments {
		if attachment.Type == messenger.AttachmentTypeImage {
			imageUrl = attachment.Payload.(*messenger.Resource).URL
			caption, err := imageutil.CaptionUrl(imageUrl)
			if err != nil {
				return err
			}
			queryWords = append(queryWords, strings.Split(caption, " ")...)
			break
		} else {
			mq.Text(string(attachment.Type) + " is not supported.")
			mess.SendMessage(mq)
		}
	}
	db := dbutil.DbContext()
	bmr, err := joinmodels.BestMemeResultsByKeywords(db, queryWords)
	if err != nil {
		return err
	}
	if bmr == nil {
		return errors.New(noMatchError)
	}
	meme, err := models.MemeByID(db, bmr.ID)
	if meme == nil || !meme.URL.Valid {
		return errors.New(noMemeError)
	}
	response := responseText
	if len(imageUrl) != 0 {
		response += followupText
		metadata, err := json.Marshal(imageMetadata{
			MemeId:   meme.ID,
			ImageUrl: imageUrl,
		})
		if err != nil {
			return err
		}
		mq.QuickReply(messenger.QuickReply{
			Title:   createQuickReply,
			Payload: string(metadata[:]),
		})
	}
	if _, err = mess.SendSimpleMessage(senderId, response); err != nil {
		return err
	}
	mq.Image(meme.URL.String)
	if _, err = mess.SendMessage(mq); err != nil {
		return err
	}
	return nil
}

func generateMeme(senderId string, msg messenger.ReceivedMessage) error {
	var metadata imageMetadata
	mq := messenger.MessageQuery{}
	mq.RecipientID(senderId)
	if msg.QuickReply != nil {
		err := json.Unmarshal([]byte(msg.QuickReply.Payload[:]), &metadata)
		if err != nil {
			return err
		}
	} else {
		return errors.New("There is no quickreply payload")
	}
	db := dbutil.DbContext()
	meme, err := models.MemeByID(db, metadata.MemeId)
	if err != nil {
		return err
	}
	resp, err := http.Get(metadata.ImageUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	img, format, err := image.Decode(resp.Body)
	if _, ok := validImageFormats[format]; !ok {
		return errors.New(format + " is not a valid file format")
	}
	newMeme := imageutil.CreateMemeFromImage(*meme, img)
	file, err := ioutil.TempFile(constants.PublicImageDir, "tempimg")
	if err != nil {
		return err
	}
	defer file.Close()
	if err = png.Encode(file, newMeme); err != nil {
		return err
	}
	file.Sync()
	mq.Image(constants.Address + "/" + file.Name())
	msgResp, err := mess.SendMessage(mq)
	if err != nil {
		return err
	}
	now := time.Now()
	tmpFileInfo := models.TempFile{
		MessageID:   msgResp.MessageID,
		FileName:    file.Name(),
		TimeCreated: &now,
	}
	if err = tmpFileInfo.Insert(db); err != nil {
		return err
	}
	return nil
}

func messageDelivered(event messenger.Event, opts messenger.MessageOpts, delivery messenger.Delivery) {
	db := dbutil.DbContext()
	for _, messageId := range delivery.MessageIDS {
		tmpFiles, _ := models.TempFilesByMessageID(db, messageId)
		if len(tmpFiles) == 0 {
			continue
		}
		for _, tmpFile := range tmpFiles {
			err := os.Remove(tmpFile.FileName)
			if err != nil {
				glog.Error(err)
			}
		}
	}
}

func messagePostback(messenger.Event, messenger.MessageOpts, messenger.Postback) {

}

func MessengerWebhook(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("hub.verify_token") == constants.FbVerificationToken {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, r.URL.Query().Get("hub.challenge"))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "Error, wrong validation token")
}

package main

import "bytes"
import "encoding/json"
import "net/http"
import "bitbucket.org/liamstask/go-imgur/imgur"
import "fmt"
import "io/ioutil"

// Request statics
var PAGE int = 1
var MEME_REQUEST_LIMIT int = 5

// URL statics
var CLIENT_ID string = "f1d6c6bea6968c6"
var CLIENT_SECRET string = "70366d1e06a6fb2e634d33cb3bfd90fe42d4e1af"
var BASE_URL string = "https://api.imgur.com/3"
var MEME_SEARCH_URL string = BASE_URL + "/image"
var AUTHORIZATION_HEADER string = fmt.Sprintf("Client-ID %s", CLIENT_ID)

// Structs to extract meme metadata
type MemeMetadataStruct struct {
  MemeName     string             `json:"meme_name,omitempty"`
  TopText      string             `json:"top_text,omitempty"`
  BottomText   string             `json:"bottom_text,omitempty"`
  BgImage      string             `json:"bg_image,omitempty"`
}

type MemeMetadataResultData struct {
  Title        string             `json:"title,omitempty"`
  MemeMetadata MemeMetadataStruct `json:"meme_metadata,omitempty"`
}

type MemeMetadataResult struct {
  Data         MemeMetadataResultData
  Status       int
  Success      bool
}

// Fields in this struct should mirror the columns in the Memes table
type MemeRow struct {
  Id           int
  Url          string
  TopText      string
  BottomText   string
  NetUps       int
  Views        int
  NumKeywords  int
  MemeName     string
  ImgurBgImage string
}

func getMemeMetadata(client *http.Client, id string) (*MemeMetadataStruct, error) {
  // Get request url
  url := MEME_SEARCH_URL + "/" + id
  buf := new(bytes.Buffer)
  
  // Create request
  req, err := http.NewRequest("GET", url, buf)
  if err != nil {
    return nil, err
  }
  req.Header.Add("Authorization", AUTHORIZATION_HEADER)

  // Execute request
  resp, err := client.Do(req)
  if err != nil {
    return nil, err
  }

  // Parse response
  defer resp.Body.Close()
  response := &MemeMetadataResult{}
  //blah, _ := ioutil.ReadAll(resp.Body)
  //fmt.Println(url, string(blah[:]))
  err = json.NewDecoder(resp.Body).Decode(response)
  /*//fmt.Println(url, req.Header, response.Status, response.Success)
  fmt.Println(response)*/
  return &response.Data.MemeMetadata, err
}

// Main function
func main() {
  // Do request
  httpClient := http.DefaultClient
  imgurClient := imgur.NewClient(httpClient, CLIENT_ID, CLIENT_SECRET)
  imagesOrAlbums, err := imgurClient.Gallery.Memes("time", "", PAGE)
  
  if (err != nil) {
    fmt.Println(imagesOrAlbums)
    fmt.Println(err)
    return;
  }

  // Process images and albums
  var memes []*MemeRow
  for index, imageOrAlbum := range imagesOrAlbums {
    if imageOrAlbum.IsAlbum {
      album := imageOrAlbum
      for _, image := range album.Images {
        meme := &MemeRow{}
        meme.Url = image.Link
        meme.NetUps = album.Score
        meme.Views = image.Views

        if (index < MEME_REQUEST_LIMIT) {
          // Get meme metadata
          memeMetadata, err := getMemeMetadata(httpClient, image.Id)
          if (err != nil) {
            fmt.Println(err)
          } else {
            meme.TopText = memeMetadata.TopText
            meme.BottomText = memeMetadata.BottomText
            meme.MemeName = memeMetadata.MemeName
            meme.ImgurBgImage = memeMetadata.BgImage
          }
        }

        memes = append(memes, meme)
      }
    } else {
      image := imageOrAlbum
      
      meme := &MemeRow{}
      meme.Url = image.Link
      meme.NetUps = image.Score
      meme.Views = image.Views
      
      if (index < MEME_REQUEST_LIMIT) {
        // Get meme metadata
        memeMetadata, err := getMemeMetadata(httpClient, image.ID)
        if (err != nil) {
          fmt.Println(err)
        } else {
          meme.TopText = memeMetadata.TopText
          meme.BottomText = memeMetadata.BottomText
          meme.MemeName = memeMetadata.MemeName
          meme.ImgurBgImage = memeMetadata.BgImage
        }
      }

      memes = append(memes, meme)
    }
  }

  for _, meme := range memes {
    fmt.Println(meme.Url, meme.NetUps, meme.Views, meme.MemeName, meme.TopText, meme.BottomText, meme.ImgurBgImage)
  }
}
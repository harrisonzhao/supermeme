package main

import (
  //"strconv"
  //"bytes"
//import "encoding/json"
  "net/http"
  "bitbucket.org/liamstask/go-imgur/imgur"
  "fmt"
  //"github.com/harrisonzhao/supermeme/shared/db"
  "../shared/db"
  //"../shared/imageutil"
  //"io"
  //"os"
  //"io/ioutil"
  //"github.com/harrisonzhao/supermeme/models"
  "github.com/golang/glog"
)

// Request statics
const (
  page_start = 0 // Will query memes from, and including, this page
  page_end = 1 // Will query memes up to, but not including, this page
  insert_limit = 1 // Will insert the first insert_limit memes into the database
/*var SORT string = "time"
var WINDOW string = "" // Only relevant if SORT is "top"
var PAGE int = 1
var MEME_REQUEST_LIMIT int = 5*/

// URL statics
  client_id = "f1d6c6bea6968c6"
  client_secret = "70366d1e06a6fb2e634d33cb3bfd90fe42d4e1af"
)
/*var BASE_URL string = "https://api.imgur.com/3"
var TOPICS_MEMES_URL string = "/topics/memes"
var AUTHORIZATION_HEADER string = fmt.Sprintf("Client-ID %s", CLIENT_ID)*/

type Tag struct {
  TotalItems int                       `json:"total_items"`
  Items      []imgur.GalleryImageAlbum `json:"items"`
}

type TagResult struct {
  Data    Tag
  Status  int
  Success bool
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

// Get memes on a certain page
func getMemes(client *imgur.Client, page int) ([]MemeRow, error) {
  // Create request url
  url := "https://api.imgur.com/3/gallery/t/memes/time"
  url = url + "/" + fmt.Sprintf("%d", page)
  
  // Create request
  req, err := client.NewRequest("GET", url, nil)
  if err != nil {
    return nil, err
  }

  // Execute request
  response := &TagResult{} // Response will hold the actual response json
  _, err = client.Do(req, response)
  //body, _ := ioutil.ReadAll(resp.Body)
  //fmt.Println(string(body[:]))
  if err != nil {
    return nil, err
  }
  //fmt.Println(url, resp, response)

  // Process images and albums
  imagesOrAlbums := response.Data.Items
  var memes []MemeRow
  for _, imageOrAlbum := range imagesOrAlbums {
    if imageOrAlbum.IsAlbum {
      // album is imgur.GalleryImageAlbum, image is imgur.Image
      album := imageOrAlbum
      for _, image := range album.Images {
        meme := MemeRow{}
        meme.Url = image.Link
        meme.NetUps = album.Ups - album.Downs
        meme.Views = image.Views

        //fmt.Println("1", image)

        /*memeMetadata := image.MemeMetadata
        meme.TopText = memeMetadata.TopText
        meme.BottomText = memeMetadata.BottomText
        meme.MemeName = memeMetadata.MemeName
        meme.ImgurBgImage = memeMetadata.BgImage*/

        memes = append(memes, meme)
      }
    } else {
      // image is imgur.GalleryImageAlbum
      image := imageOrAlbum
      
      meme := MemeRow{}
      meme.Url = image.Link
      meme.NetUps = image.Ups - image.Downs
      meme.Views = image.Views

      //fmt.Println("2", image)
      
      /*memeMetadata := image.MemeMetadata
      meme.TopText = memeMetadata.TopText
      meme.BottomText = memeMetadata.BottomText
      meme.MemeName = memeMetadata.MemeName
      meme.ImgurBgImage = memeMetadata.BgImage*/

      memes = append(memes, meme)
    }
  }

  return memes, nil
}

// Get all memes on the desired pages
func getAllMemes() ([]MemeRow) {
  httpClient := http.DefaultClient
  var memes []MemeRow

  for page := page_start; page < page_end; page++ {
    imgurClient := imgur.NewClient(httpClient, client_id, client_secret)
    pageMemes, err := getMemes(imgurClient, page)
    if (err != nil) {
      glog.Error(fmt.Sprintf("Could not retrieve memes on page: %d", page), err)
      continue
    }
    memes = append(memes, pageMemes...)
  }

  return memes
}

func main() {
  memes := getAllMemes()

  dbutil.InitDb("alpha")
  db := dbutil.DbContext()

  for index, meme := range memes {
    if (index < insert_limit) {
      /*keyword, err := imageUtil.CaptionUrl(meme.Url)
      if (err != nil) {
        glog.Error(fmt.Sprintf("Could not retrieve keyword for image: %s", meme.Url), err)
        continue
      }*/
      keyword := "Woot"
      fmt.Println(meme.Url, meme.NetUps, meme.Views, keyword)

      _, err := db.Exec("INSERT INTO alpha.meme (url, net_ups, views, num_keywords) VALUES (?, ?, ?, ?)", meme.Url, meme.NetUps, meme.Views, 1)
      if (err != nil) {
        glog.Error(fmt.Sprintf("Could not insert image into database: %s", meme.Url), err)
        continue
      }
      
      rows, err := db.Query("SELECT id FROM alpha.meme WHERE url = ?", meme.Url)
      if (err != nil) {
        glog.Error(fmt.Sprintf("Could not insert keyword into database for image: %s", meme.Url), err)
        continue
      }
      defer rows.Close()
      rows.Next() // Check for multiples or errors
      id := new(int64)
      err = rows.Scan(id)
      if (err != nil) {
        glog.Error(fmt.Sprintf("Could not insert keyword into database for image: %s", meme.Url), err)
        continue
      }

      meme.Id = int(*id)
      fmt.Println(meme.Id)
      _, err = db.Exec("INSERT INTO alpha.meme_keyword (meme_id, keyword) VALUES (?, ?)", meme.Id, keyword)
      if (err != nil) {
        glog.Error(fmt.Sprintf("Could not insert keyword into database for image: %s", meme.Url), err)
        continue
      }
    }
  }
}
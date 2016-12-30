package main

import (
  "net/http"
  "bitbucket.org/liamstask/go-imgur/imgur"
  "fmt"
  "../shared/db"
  "../shared/imageutil"
  "database/sql"
  "github.com/harrisonzhao/supermeme/models"
  "github.com/golang/glog"
)

// Request statics
const (
  // Query parameters
  page_start = 55 // Will query memes from, and including, this page
  page_end = 56 // Will query memes up to, but not including, this page
  insert_limit = 5 // Will insert the first insert_limit memes into the database

  // URL parameters
  client_id = "f1d6c6bea6968c6"
  client_secret = "70366d1e06a6fb2e634d33cb3bfd90fe42d4e1af"

  // URL statics
  meme_url = "https://api.imgur.com/3/gallery/t/memes/time"
)

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
/*type MemeRow struct {
  Id           int
  Url          string
  TopText      string
  BottomText   string
  NetUps       int
  Views        int
  NumKeywords  
  MemeName     string
  ImgurBgImage string
}*/

// Helper functions to convert normal types into sql types
func stringToNullString(s string) (sql.NullString) {
  return sql.NullString{String : s, Valid : s != ""}
}

func intToNullInt64(i int) (sql.NullInt64) {
  return sql.NullInt64{Int64 : int64(i), Valid : true}
}

func nullStringToString(ns sql.NullString) (string) {
  if (!ns.Valid) {
    return ""
  }
  return ns.String
}

func nullInt64ToInt(ni sql.NullInt64) (int) {
  if (!ni.Valid) {
    return 0
  }
  return int(ni.Int64)
}

// Get memes on a certain page
func getMemes(client *imgur.Client, page int) ([]models.Meme, error) {
  // Create request url
  url := meme_url + "/" + fmt.Sprintf("%d", page)
  
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
  var memes []models.Meme
  for _, imageOrAlbum := range imagesOrAlbums {
    if imageOrAlbum.IsAlbum {
      // album is imgur.GalleryImageAlbum, image is imgur.Image
      album := imageOrAlbum
      for _, image := range album.Images {
        meme := models.Meme{}
        meme.URL = stringToNullString(image.Link)
        meme.NetUps = intToNullInt64(album.Ups - album.Downs)
        meme.Views = intToNullInt64(image.Views)

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
      
      meme := models.Meme{}
      meme.URL = stringToNullString(image.Link)
      meme.NetUps = intToNullInt64(image.Ups - image.Downs)
      meme.Views = intToNullInt64(image.Views)

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
func getAllMemes() ([]models.Meme) {
  httpClient := http.DefaultClient
  var memes []models.Meme

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

  // TODO: Get text from memes and filter out ones with garbage text
  //top_text, bottom_text, err := imageutil.GetTextFromMeme()

  dbutil.InitDb("alpha")
  db := dbutil.DbContext()

  fmt.Println(1)

  for index, meme := range memes {
    if (index < insert_limit) {
      // Get caption keywords
      keyword, err := imageutil.CaptionUrl(nullStringToString(meme.URL))
      if (err != nil) {
        glog.Error(fmt.Sprintf("Could not retrieve keyword for image: %s", nullStringToString(meme.URL)), err)
        continue
      }
      fmt.Println(meme.URL, meme.NetUps, meme.Views, keyword)
      meme.NumKeywords = intToNullInt64(1)
      
      // Database queries
      err = meme.Save(db)
      if (err != nil) {
        glog.Error(fmt.Sprintf("Could not insert image into database: %s", nullStringToString(meme.URL)), err)
        continue
      }

      rows, err := db.Query("SELECT id FROM alpha.meme WHERE url = ?", meme.URL)
      if (err != nil) {
        glog.Error(fmt.Sprintf("Could not insert keyword into database for image: %s", nullStringToString(meme.URL)), err)
        continue
      }
      defer rows.Close()
      if (!rows.Next()) {
        glog.Error(fmt.Sprintf("Could not insert keyword into database for image: %s", nullStringToString(meme.URL)), err)
        continue
      }
      id := new(int64)
      err = rows.Scan(id)
      if (err != nil) {
        glog.Error(fmt.Sprintf("Could not insert keyword into database for image: %s", nullStringToString(meme.URL)), err)
        continue
      }

      memeKeyword := models.MemeKeyword{
        MemeID: int(*id),
        Keyword: keyword,
      }
      /*err = memeKeyword.Save(db)*/
      _, err = db.Exec("INSERT INTO alpha.meme_keyword (meme_id, keyword) VALUES (?, ?)", memeKeyword.MemeID, memeKeyword.Keyword)
      if (err != nil) {
        glog.Error(fmt.Sprintf("Could not insert keyword into database for image: %s", nullStringToString(meme.URL)), err)
        continue
      }
    }
  }
}
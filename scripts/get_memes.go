package main

import (
  "regexp"
  "net/http"
  "bitbucket.org/liamstask/go-imgur/imgur"
  "fmt"
  "strings"
  "../shared/db"
  "../shared/imageutil"
  "database/sql"
  "github.com/harrisonzhao/supermeme/models"
  "github.com/golang/glog"
  "image"
  _ "image/png"
  "github.com/Masterminds/squirrel"
)

// Request statics
const (
  // Query parameters
  page_start = 57 // Will query memes from, and including, this page
  page_end = 58 // Will query memes up to, but not including, this page
  insert_limit = 1 // Will insert the first insert_limit memes into the database

  // URL parameters
  client_id = "f1d6c6bea6968c6"
  client_secret = "70366d1e06a6fb2e634d33cb3bfd90fe42d4e1af"

  // URL statics
  meme_url = "https://api.imgur.com/3/gallery/t/memes/time"
)

var reg, _ = regexp.Compile("[^A-Za-z0-9 ]+")

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
  ID           int
  Source       models.Source
  URL          string
  TopText      string
  BottomText   string
  NetUps       int
  Views        int
  Keywords     []string
}

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
func getMemes(client *imgur.Client, page int) ([]MemeRow, error) {
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
        meme.Source = models.SourceImgur
        meme.URL = image.Link
        meme.NetUps = album.Ups - album.Downs
        meme.Views = image.Views

        memes = append(memes, meme)
      }
    } else {
      // image is imgur.GalleryImageAlbum
      image := imageOrAlbum
      
      meme := MemeRow{}
      meme.Source = models.SourceImgur
      meme.URL = image.Link
      meme.NetUps = image.Ups - image.Downs
      meme.Views = image.Views

      memes = append(memes, meme)
    }
  }

  return memes, nil
}

// Get all memes on the desired range of pages
// The MemeRows returned will have Source, URL, NetUps, Views set
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

// Populate the text fields for the memes passed in
// The MemeRow returned will have BottomText, TopText, Keywords set
// If BottomText, TopText, Keywords cannot be processed, then the meme will be dropped from the returned value
func populateTextForMemes(rawMemes []MemeRow) ([]MemeRow) {
  var memes []MemeRow
  for _, meme := range rawMemes {
    if (len(memes) < insert_limit) {
      // Get BottomText and TopText fields
      resp, err := http.Get(meme.URL)
      if err != nil {
        glog.Error(fmt.Sprintf("Could not get text for image: %s", meme.URL), err)
        continue
      }
      
      defer resp.Body.Close()
      img, format, err := image.Decode(resp.Body)
      if (err != nil) || (format != "png") {
        glog.Error(fmt.Sprintf("Could not get text for image: %s", meme.URL), err)
        continue
      }

      topText, bottomText, err := imageutil.GetTextFromMeme(img)
      if err != nil {
        glog.Error(fmt.Sprintf("Could not get text for image: %s", meme.URL), err)
        continue
      }
      //fmt.Println(nullStringToString(meme.URL), topText, bottomText)
      meme.TopText = topText
      meme.BottomText = bottomText

      // Get Keywords field
      rawCaption, err := imageutil.CaptionUrl(meme.URL)
      if (err != nil) {
        glog.Error(fmt.Sprintf("Could not retrieve caption for image: %s", meme.URL), err)
        continue
      }
      //fmt.Println(meme.URL, meme.NetUps, meme.Views, caption)
      
      caption := reg.ReplaceAllString(rawCaption, "")
      rawCaptionWords := strings.Split(caption, " ")
      keywordSet := make(map[string]bool)
      for _, rawCaptionWord := range rawCaptionWords {
        captionWord := strings.ToLower(rawCaptionWord)
        keywordSet[captionWord] = true
      }
      var keywords []string
      for keyword := range keywordSet {
        keywords = append(keywords, keyword)
      }
      meme.Keywords = keywords
      
      memes = append(memes, meme)
    }
  }

  return memes
}

func main() {
  rawMemes := getAllMemes()
  memes := populateTextForMemes(rawMemes)

  //fmt.Println(memes)

  dbutil.InitDb("alpha")
  db := dbutil.DbContext()

  // Get all the memes that are already in the database
  var urls []string
  for _, meme := range memes {
    urls = append(urls, meme.URL)
  }
  
  sql, args, err := squirrel.
    Select("id", "url").
    From("alpha.meme").
    Where("url IN (" + squirrel.Placeholders(len(urls)) + ")", urls...).
    ToSql()
  if (err != nil) {
    glog.Fatal(err)
  }

  fmt.Println(sql, args)
  
  rows, err := db.Query(sql, args)
  if (err != nil) {
    glog.Fatal(err)
  }

  idTemp := new(int)
  urlTemp := new(string)
  urlToIdMap := make(map[string]int)
  for rows.Next() {
    err = rows.Scan(idTemp, urlTemp)
    urlToIdMap[*urlTemp] = *idTemp
  }

  var oldMemes []MemeRow
  var newMemes []MemeRow
  for _, meme := range memes {
    id, exists := urlToIdMap[meme.URL]
    if (exists) {
      meme.ID = id
      oldMemes = append(oldMemes, meme)
    } else {
      newMemes = append(newMemes, meme)
    }
  }

  fmt.Println("oldmemes", oldMemes)
  fmt.Println("newmemes", newMemes)
/*
   
      rows, err := db.Query("SELECT id FROM alpha.meme WHERE url = ?", meme.URL)
      if (err != nil) {
        glog.Error(fmt.Sprintf("Could not insert keyword into database for image: %s", nullStringToString(meme.URL)), err)
        continue
      }
      
      // Database queries
      sqlMeme := models.Meme{
        Source: meme.Source,
        URL: meme.URL,
        TopText: meme.TopText,
        BottomText: meme.BottomText,
        NetUps: meme.NetUps,
        Views: meme.Views,
        NumKeywords
      }
      
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

      
      _, err = db.Exec("INSERT INTO alpha.meme_keyword (meme_id, keyword) VALUES (?, ?)", memeKeyword.MemeID, memeKeyword.Keyword)
      if (err != nil) {
        glog.Error(fmt.Sprintf("Could not insert keyword into database for image: %s", nullStringToString(meme.URL)), err)
        continue
      }
    }
  }*/
}
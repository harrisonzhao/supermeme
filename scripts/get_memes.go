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
  page_start = 1 // Will query memes from, and including, this page
  page_end = 2 // Will query memes up to, but not including, this page
  insert_limit = 1 // Will insert the first insert_limit memes into the database

  // URL parameters
  client_id = "f1d6c6bea6968c6"
  client_secret = "70366d1e06a6fb2e634d33cb3bfd90fe42d4e1af"

  // URL statics
  meme_url = "https://api.imgur.com/3/gallery/t/memes/time"
)

var regNonLetters, _ = regexp.Compile("[^A-Za-z0-9 ]+")
var regStopWords, _ = regexp.Compile(
  "^(i|am|im|not|really|confident|but|i|think|it|is|its|a|in|on|and|of|the|he|seems)$")

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
      
      caption := regNonLetters.ReplaceAllString(rawCaption, "")
      rawCaptionWords := strings.Split(caption, " ")
      keywordSet := make(map[string]bool)
      for _, rawCaptionWord := range rawCaptionWords {
        captionWord := strings.ToLower(rawCaptionWord)
        if (!regStopWords.MatchString(captionWord)) {
          keywordSet[captionWord] = true
        }
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

// Get all the memes that are already in the database
func divideIntoOldAndNewMemes(db models.XODB, memes []MemeRow) ([]MemeRow, []MemeRow, error) {
  // Query for existing memes based on url
  urls := make([]interface{}, len(memes))
  for i, meme := range memes {
    urls[i] = meme.URL
  }
  sql, args, err := squirrel.
    Select("id", "url").
    From("alpha.meme").
    Where("url IN (" + squirrel.Placeholders(len(urls)) + ")", urls...).
    ToSql()
  if (err != nil) {
    return nil, nil, err
  }
  rows, err := db.Query(sql, args...)
  if (err != nil) {
    return nil, nil, err
  }

  idTemp := new(int)
  urlTemp := new(string)
  urlToIdMap := make(map[string]int)
  for rows.Next() {
    err = rows.Scan(idTemp, urlTemp)
    urlToIdMap[*urlTemp] = *idTemp
  }

  // Divide up memes into old memes and new memes
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

  return oldMemes, newMemes, nil
}

func updateMemes(db models.XODB, memes []MemeRow) (error) {
  // TODO: Modify so we don't delete all keywords every time
  // Delete meme keywords
  ids := make([]interface{}, len(memes))
  for i, meme := range memes {
    ids[i] = meme.ID
  }
  sql, args, err := squirrel.
    Delete("alpha.meme_keyword").
    Where("meme_id IN (" + squirrel.Placeholders(len(ids)) + ")", ids...).
    ToSql()
  if (err != nil) {
    return err
  }
  fmt.Println(sql, args)
  _, err = db.Exec(sql, args...)
  if (err != nil) {
    return err
  }

  // Update memes
  var memeRowValues []interface{}
  for _, meme := range memes {
    memeRowValues = append(memeRowValues, []interface{}{
      meme.ID,
      meme.Source,
      meme.URL,
      meme.TopText,
      meme.BottomText,
      meme.NetUps,
      meme.Views,
      len(meme.Keywords),
    }...)
  }
  sql = "INSERT INTO alpha.meme (id, source, url, top_text, bottom_text, net_ups, views, num_keywords) VALUES "
  for i, _ := range memes {
    sql = sql + "(?,?,?,?,?,?,?,?)"
    if (i < len(memes) - 1) {
      sql = sql + ", "
    }
  }
  sql = sql + " ON DUPLICATE KEY UPDATE net_ups = VALUES(net_ups), views = VALUES(views), num_keywords = VALUES(num_keywords)"

  fmt.Println(sql, memeRowValues)
  _, err = db.Exec(sql, memeRowValues...)
  if (err != nil) {
    return err
  }

  // Get total number of keywords
  totalKeywords := 0
  for _, meme := range memes {
    totalKeywords = totalKeywords + len(meme.Keywords)
  }
  if totalKeywords <= 0 {
    glog.Infoln("No keywords associated to old memes")
    return nil
  }

  // Insert keywords
  var memeKeywordRowValues [][]interface{}
  for _, meme := range memes {
    for _, keyword := range meme.Keywords {
      memeKeywordRowValues = append(memeKeywordRowValues, []interface{}{
        meme.ID,
        keyword,
      })
    }
  }
  builder := squirrel.
    Insert("alpha.meme_keyword").
    Columns("meme_id", "keyword")
  for _, memeKeywordRowValue := range memeKeywordRowValues {
    builder = builder.Values(memeKeywordRowValue...)
  }
  sql, args, err = builder.ToSql()
  if (err != nil) {
    return err
  }
  fmt.Println(sql, args)
  _, err = db.Exec(sql, args...)
  if (err != nil) {
    return err
  }

  return nil
}

func insertMemes(db models.XODB, memes [] MemeRow) (error) {
  // Insert memes
  memeRowValues := make([][]interface{}, len(memes))
  for i, meme := range memes {
    memeRowValues[i] = []interface{}{
      meme.ID,
      meme.Source,
      meme.URL,
      meme.TopText,
      meme.BottomText,
      meme.NetUps,
      meme.Views,
      len(meme.Keywords),
    }
  }
  builder := squirrel.
    Insert("alpha.meme").
    Columns("id", "source", "url", "top_text", "bottom_text", "net_ups", "views", "num_keywords")
  for _, memeRowValue := range memeRowValues {
    builder = builder.Values(memeRowValue...)
  }
  sql, args, err := builder.ToSql()
  if (err != nil) {
    return err
  }
  fmt.Println(sql, args)
  _, err = db.Exec(sql, args...)
  if (err != nil) {
    return err
  }

  // Get total number of keywords
  totalKeywords := 0
  for _, meme := range memes {
    totalKeywords = totalKeywords + len(meme.Keywords)
  }
  if totalKeywords <= 0 {
    glog.Infoln("No keywords associated to new memes")
    return nil
  }

  // Query for ids of new memes
  urls := make([]interface{}, len(memes))
  for i, meme := range memes {
    urls[i] = meme.URL
  }
  sql, args, err = squirrel.
    Select("id", "url").
    From("alpha.meme").
    Where("url IN (" + squirrel.Placeholders(len(urls)) + ")", urls...).
    ToSql()
  if (err != nil) {
    return err
  }
  rows, err := db.Query(sql, args...)
  if (err != nil) {
    return err
  }

  idTemp := new(int)
  urlTemp := new(string)
  urlToIdMap := make(map[string]int)
  for rows.Next() {
    err = rows.Scan(idTemp, urlTemp)
    urlToIdMap[*urlTemp] = *idTemp
  }

  // Insert keywords
  var memeKeywordRowValues [][]interface{}
  for _, meme := range memes {
    for _, keyword := range meme.Keywords {
      memeKeywordRowValues = append(memeKeywordRowValues, []interface{}{
        urlToIdMap[meme.URL],
        keyword,
      })
    }
  }
  builder = squirrel.
    Insert("alpha.meme_keyword").
    Columns("meme_id", "keyword")
  for _, memeKeywordRowValue := range memeKeywordRowValues {
    builder = builder.Values(memeKeywordRowValue...)
  }
  sql, args, err = builder.ToSql()
  if (err != nil) {
    return err
  }
  fmt.Println(sql, args)
  _, err = db.Exec(sql, args...)
  if (err != nil) {
    return err
  }

  return nil
}

func main() {
  dbutil.InitDb("alpha")
  db := dbutil.DbContext()

  rawMemes := getAllMemes()
  memes := populateTextForMemes(rawMemes)
  if (len(memes) == 0) {
    return
  }

  //fmt.Println(memes)

  oldMemes, newMemes, err := divideIntoOldAndNewMemes(db, memes)
  if (err != nil) {
    glog.Fatal("Could not divide memes into old and new memes", err)
  }

  fmt.Println(oldMemes)
  fmt.Println(newMemes)

  if (len(oldMemes) > 0) {
    err = updateMemes(db, oldMemes)
    if (err != nil) {
      glog.Fatal("Could not update old memes in the database", err)
    }
  }

  if (len(newMemes) > 0) {
    err = insertMemes(db, newMemes)
    if (err != nil) {
      glog.Fatal("Could not insert new memes into database", err)
    }
  }
}
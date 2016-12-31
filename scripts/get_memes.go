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

// Script parameters (Change ONLY these!)
var page_mode = true // True: Load memes by page, False: Load individual memes

var page_start = 1 // Will load memes from, and including, this page
var page_end = 5 // Will load memes up to, but not including, this page
var insert_limit = 100 // Will insert the first insert_limit memes with all fields from each page into the database

var meme_id_list = []int{} // Will load memes whose ids are in this list

// Final statics
const (
  // Request parameters
  client_id = "f1d6c6bea6968c6"
  client_secret = "70366d1e06a6fb2e634d33cb3bfd90fe42d4e1af"

  // URL
  page_url = "https://api.imgur.com/3/gallery/t/memes/time"
  id_url = "https://api.imgur.com/3/gallery/t/memes"
)

// Regexp
var regNonLetters, _ = regexp.Compile("[^A-Za-z0-9 ]+")
var regStopWords, _ = regexp.Compile(
  "^(i|am|im|not|really|confident|but|i|think|it|is|its|a|in|on|and|of|the|he|seems|to|next)$")

// Non-final statics
var db models.XODB
var imgurClient *imgur.Client

// Tag struct to help in parsing image tag api call responses
type tag struct {
  TotalItems int                       `json:"total_items"`
  Items      []imgur.GalleryImageAlbum `json:"items"`
}

type tagResult struct {
  Data    tag
  Status  int
  Success bool
}

type galleryImageAlbumResult struct {
  Data    imgur.GalleryImageAlbum
  Status  int
  Success bool
}

// Fields in this struct should mirror the columns in the Memes table
type memeRow struct {
  ID           int
  Source       models.Source
  URL          string
  TopText      string
  BottomText   string
  NetUps       int
  Views        int
  Keywords     []string
}

/*
 * TODO: REFACTOR REDUNDANT CODE!
 * TODO: REFACTOR REDUNDANT CODE!
 * TODO: REFACTOR REDUNDANT CODE!
 * TODO: REFACTOR REDUNDANT CODE!
 * TODO: REFACTOR REDUNDANT CODE!
 */

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

// Get imgur.GalleryImageAlbums tagged as memes on a certain page
func getImagesOrAlbumsOnPage(page int) ([]imgur.GalleryImageAlbum, error) {
  // Create request url
  url := page_url + "/" + fmt.Sprintf("%d", page)
  
  // Create request
  req, err := imgurClient.NewRequest("GET", url, nil)
  if err != nil {
    return nil, err
  }

  // Execute request
  response := &tagResult{} // Response will hold the actual response json
  _, err = imgurClient.Do(req, response)
  if err != nil {
    return nil, err
  }
  //fmt.Println(url, resp, response)
  return response.Data.Items, nil
}

// Get all imgur.GalleryImageAlbums with given ids
func getImagesOrAlbumsWithIds() ([]imgur.GalleryImageAlbum, error) {
  // Get all the urls for the memes with given ids
  ids := make([]interface{}, len(meme_id_list))
  for i, id := range meme_id_list {
    ids[i] = id
  }
  sql, args, err := squirrel.
    Select("id", "url").
    From("alpha.meme").
    Where("id IN (" + squirrel.Placeholders(len(ids)) + ")", ids...).
    ToSql()
  if (err != nil) {
    return nil, err
  }
  fmt.Println(sql, args)
  rows, err := db.Query(sql, args...)
  if (err != nil) {
    return nil, err
  }

  idTemp := new(int)
  urlTemp := new(string)
  idToUrlMap := make(map[int]string)
  for rows.Next() {
    err = rows.Scan(idTemp, urlTemp)
    idToUrlMap[*idTemp] = (*urlTemp)[19:26] // TODO: Hacky way to get imgur ID, just save it in the database, perhaps?
  }

  // Get the imgur.GalleryImageAlbum for each url
  var imagesOrAlbums []imgur.GalleryImageAlbum

  for _, id := range meme_id_list {
    // Create request url
    url := id_url + "/" + idToUrlMap[id]
    
    // Create request
    req, err := imgurClient.NewRequest("GET", url, nil)
    if err != nil {
      return nil, err
    }

    // Execute request
    response := &galleryImageAlbumResult{} // Response will hold the actual response json
    _, err = imgurClient.Do(req, response)
    if err != nil {
      return nil, err
    }
    //fmt.Println(url, resp, response)
    imagesOrAlbums = append(imagesOrAlbums, response.Data)
  }

  return imagesOrAlbums, nil
}

/*// Get all memes on the desired range of pages
// The memeRows returned will have Source, URL, NetUps, Views set
func getAllMemes() ([]memeRow) {
  httpClient := http.DefaultClient
  var memes []memeRow

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
}*/

// Convert the given imgur.GalleryImageAlbums into memeRows
// The memeRow returned will have Source, URL, NetUps, Views, BottomText, TopText, Keywords set
// If BottomText, TopText, Keywords cannot be processed, then the meme will be dropped from the returned value
func convertImagesOrAlbumsToMemes(imagesOrAlbums []imgur.GalleryImageAlbum) ([]memeRow) {
  // Process images and albums
  var rawMemes []memeRow

  for _, imageOrAlbum := range imagesOrAlbums {
    if imageOrAlbum.IsAlbum {
      // album is imgur.GalleryImageAlbum, image is imgur.Image
      album := imageOrAlbum
      for _, image := range album.Images {
        meme := memeRow{}
        meme.Source = models.SourceImgur
        meme.URL = image.Link
        meme.NetUps = album.Ups - album.Downs
        meme.Views = image.Views

        rawMemes = append(rawMemes, meme)
      }
    } else {
      // image is imgur.GalleryImageAlbum
      image := imageOrAlbum
      
      meme := memeRow{}
      meme.Source = models.SourceImgur
      meme.URL = image.Link
      meme.NetUps = image.Ups - image.Downs
      meme.Views = image.Views

      rawMemes = append(rawMemes, meme)
    }
  }

  // Add in the remaining fields for memes
  var memes []memeRow

  for _, meme := range rawMemes {
    if (len(memes) < insert_limit) {
      // Get BottomText and TopText fields
      resp, err := http.Get(meme.URL)
      if err != nil {
        glog.Info(fmt.Sprintf("Unable to get text for image: %s", meme.URL), err)
        continue
      }
      
      defer resp.Body.Close()
      img, format, err := image.Decode(resp.Body)
      if (err != nil) || (format != "png") {
        glog.Info(fmt.Sprintf("Unable to get text for image: %s", meme.URL), err)
        continue
      }

      topText, bottomText, err := imageutil.GetTextFromMeme(img)
      if err != nil {
        glog.Info(fmt.Sprintf("Unable to get text for image: %s", meme.URL), err)
        continue
      }
      //fmt.Println(nullStringToString(meme.URL), topText, bottomText)
      meme.TopText = topText
      meme.BottomText = bottomText

      // Get Keywords field
      rawCaption, err := imageutil.CaptionUrl(meme.URL)
      if (err != nil) {
        glog.Info(fmt.Sprintf("Unable to retrieve caption for image: %s", meme.URL), err)
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
func divideIntoOldAndNewMemes(memes []memeRow) ([]memeRow, []memeRow, error) {
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
  var oldMemes []memeRow
  var newMemes []memeRow
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

func updateMemes(memes []memeRow) (error) {
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
    glog.Info("No keywords associated to old memes")
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

func insertMemes(memes [] memeRow) (error) {
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
    glog.Info("No keywords associated to new memes")
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
  fmt.Println(sql, args)
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

func loadImagesOrAlbums(imagesOrAlbums []imgur.GalleryImageAlbum) (error) {
  // Convert the imgur.GalleryImageAlbums into memeRows
  //fmt.Println(0)

  memes := convertImagesOrAlbumsToMemes(imagesOrAlbums)
  if (len(memes) == 0) {
    return nil
  }

  //fmt.Println(1)

  // Decide which memes are old and which are new
  oldMemes, newMemes, err := divideIntoOldAndNewMemes(memes)
  if (err != nil) {
    return err
  }

  fmt.Println(oldMemes)
  fmt.Println(newMemes)

  // Upload the meme data into the database
  if (len(oldMemes) > 0) {
    err = updateMemes(oldMemes)
    if (err != nil) {
      return err
    }
  }

  if (len(newMemes) > 0) {
    err = insertMemes(newMemes)
    if (err != nil) {
      return err
    }
  }

  return nil
}

func loadMemesForPage(page int) (error) {
  // Get all imgur.GalleryImageAlbums tagged as memes on page
  imagesOrAlbums, err := getImagesOrAlbumsOnPage(page)
  if (err != nil) {
    return err
  }
  // Convert and load all imgur.GalleryImageAlbums
  err = loadImagesOrAlbums(imagesOrAlbums)
  if (err != nil) {
    return err
  }
  return nil
}

func loadMemesWithIds() (error) {
  // Get all imgur.GalleryImageAlbums with given ids
  imagesOrAlbums, err := getImagesOrAlbumsWithIds()
  if (err != nil) {
    return err
  }
  // Convert and load all imgur.GalleryImageAlbums
  // TODO: All memes passed through here exist in the database, so the later sql call to get meme_ids is redundant. Fix.
  err = loadImagesOrAlbums(imagesOrAlbums)
  if (err != nil) {
    return err
  }
  return nil
}

func main() {
  // Initialize database context and imgur client
  dbutil.InitDb("alpha")
  db = dbutil.DbContext()
  imgurClient = imgur.NewClient(http.DefaultClient, client_id, client_secret)

  if (page_mode) {
    // Upsert memes for each page in range
    for page := page_start; page < page_end; page++ {
      err := loadMemesForPage(page);
      if (err != nil) {
        glog.Error(fmt.Sprintf("Unable to upsert memes into database for page: %d", page), err)
      } else {
        glog.Info(fmt.Sprintf("Successfully upserted memes into database for page: %d", page))
      }
    }
  } else {
    // Upsert all memes with given ids
    err := loadMemesWithIds();
    if (err != nil) {
      glog.Error("Unable to upsert all memes into database with given ids", err)
    } else {
      glog.Info("Successfully upserted all memes into database with given ids")
    }
  }
}
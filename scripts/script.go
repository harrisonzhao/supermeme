package main

import "strconv"
import "bytes"
import "encoding/json"
import "net/http"
//import "bitbucket.org/liamstask/go-imgur/imgur"
import "fmt"
//import "io/ioutil"

// Request statics
var SORT string = "time"
var WINDOW string = "" // Only relevant if SORT is "top"
var PAGE int = 1
var MEME_REQUEST_LIMIT int = 5

// URL statics
var CLIENT_ID string = "f1d6c6bea6968c6"
var CLIENT_SECRET string = "70366d1e06a6fb2e634d33cb3bfd90fe42d4e1af"
var BASE_URL string = "https://api.imgur.com/3"
var MEME_SEARCH_URL string = BASE_URL + "/g/memes"
var AUTHORIZATION_HEADER string = fmt.Sprintf("Client-ID %s", CLIENT_ID)

// Structs to extract meme metadata
type MemeMetadataStruct struct {
  MemeName     string             `json:"meme_name,omitempty"`
  TopText      string             `json:"top_text,omitempty"`
  BottomText   string             `json:"bottom_text,omitempty"`
  BgImage      string             `json:"bg_image,omitempty"`
}

/*type MemeMetadataResultData struct {
  Title        string             `json:"title,omitempty"`
  MemeMetadata MemeMetadataStruct `json:"meme_metadata,omitempty"`
}

type MemeMetadataResult struct {
  Data         MemeMetadataResultData
  Status       int
  Success      bool
}*/

type Image struct {
  // Common fields
  Id           string `json:"id,omitempty"`
  Title        string `json:"title,omitempty"`
  Description  string `json:"description,omitempty"`
  Datetime     int    `json:"datetime,omitempty"`
  MimeType     string `json:"type,omitempty"`
  Animated     bool   `json:"animated,omitempty"`
  Width        int    `json:"width,omitempty"`
  Height       int    `json:"height,omitempty"`
  Size         int    `json:"size,omitempty"`
  Views        int    `json:"views,omitempty"`
  Bandwidth    int    `json:"bandwidth,omitempty"`
  DeleteHash   string `json:"deletehash,omitempty"`
  Name         string `json:"name,omitempty"`
  Section      string `json:"section,omitempty"`
  Link         string `json:"link,omitempty"`
  Gifv         string `json:"gifv,omitempty"`
  Mp4          string `json:"mp4,omitempty"`
  Mp4Size      int    `json:"mp4_size,omitempty"`
  Looping      bool   `json:"looping,omitempty"`
  Favorite     bool   `json:"favorite,omitempty"`
  Nsfw         bool   `json:"nsfw,omitempty"`
  Vote         string `json:"vote,omitempty"`
  InGallery    bool   `json:"in_gallery,omitempty"`
  MemeMetadata MemeMetadataStruct `json:"meme_metadata,omitempty"`
}

type GalleryImageOrAlbum struct {
  // Common fields
  Id           string `json:"id,omitempty"`
  Title        string `json:"title,omitempty"`
  Description  string `json:"description,omitempty"`
  Datetime     int    `json:"datetime,omitempty"`
  Views        int    `json:"views,omitempty"`
  Link         string `json:"link,omitempty"`
  Vote         string `json:"vote,omitempty"`
  Favorite     bool   `json:"favorite,omitempty"`
  Nsfw         bool   `json:"nsfw,omitempty"`
  CommentCount int    `json:"comment_count,omitempty"`
  Topic        string `json:"topic,omitempty"`
  TopicId      int    `json:"topic_id,omitempty"`
  AccountUrl   string `json:"account_url,omitempty"`
  AccountId    int    `json:"account_id,omitempty"`
  Ups          int    `json:"ups,omitempty"`
  Downs        int    `json:"downs,omitempty"`
  Points       int    `json:"points,omitempty"`
  Score        int    `json:"score,omitempty"`
  IsAlbum      bool   `json:"is_album,omitempty"`

  // Gallery image only fields
  MimeType     string             `json:"type,omitempty"`
  Animated     bool               `json:"animated,omitempty"`
  Width        int                `json:"width,omitempty"`
  Height       int                `json:"height,omitempty"`
  Size         int                `json:"size,omitempty"`
  Bandwidth    int                `json:"bandwidth,omitempty"`
  DeleteHash   string             `json:"deletehash,omitempty"`
  Gifv         string             `json:"gifv,omitempty"`
  Mp4          string             `json:"mp4,omitempty"`
  Mp4Size      int                `json:"mp4_size,omitempty"`
  Looping      bool               `json:"looping,omitempty"`
  Section      string             `json:"section,omitempty"`
  MemeMetadata MemeMetadataStruct `json:"meme_metadata,omitempty"`

  // Gallery album only fields
  Cover       string  `json:"cover,omitempty"`
  CoverWidth  int     `json:"cover_width,omitempty"`
  CoverHeight int     `json:"cover_height,omitempty"`
  Privacy     string  `json:"privacy,omitempty"`
  Layout      string  `json:"layout,omitempty"`
  ImagesCount int     `json:"images_count,omitempty"`
  Images      []Image `json:"images,omitempty"`
}

type galleryImageOrAlbumResult struct {
  Data    []GalleryImageOrAlbum
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

func getMemes(client *http.Client, sort string, window string, page int) ([]GalleryImageOrAlbum, error) {
  // Get request url
  url := ""
  if sort == "top" {
    url = MEME_SEARCH_URL + "/" + sort + "/" + window + "/" + strconv.Itoa(page)
  } else {
    url = MEME_SEARCH_URL + "/" + sort + "/" + strconv.Itoa(page)
  }
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
  response := &galleryImageOrAlbumResult{}
  //blah, _ := ioutil.ReadAll(resp.Body)
  //fmt.Println(url, string(blah[:]))
  err = json.NewDecoder(resp.Body).Decode(response)
  //fmt.Println(url, req.Header, response.Status, response.Success)
  //fmt.Println(response.Data)
  return response.Data, err
}

// Main function
func main() {
  // Do request
  httpClient := http.DefaultClient
  imagesOrAlbums, err := getMemes(httpClient, SORT, WINDOW, PAGE)
  
  if (err != nil) {
    fmt.Println(imagesOrAlbums)
    fmt.Println(err)
    return;
  }

  // Process images and albums
  var memes []*MemeRow
  for _, imageOrAlbum := range imagesOrAlbums {
    if imageOrAlbum.IsAlbum {
      album := imageOrAlbum
      for _, image := range album.Images {
        meme := &MemeRow{}
        meme.Url = image.Link
        meme.NetUps = album.Points
        meme.Views = image.Views

        //fmt.Println("1", image)

        memeMetadata := image.MemeMetadata
        meme.TopText = memeMetadata.TopText
        meme.BottomText = memeMetadata.BottomText
        meme.MemeName = memeMetadata.MemeName
        meme.ImgurBgImage = memeMetadata.BgImage

        memes = append(memes, meme)
      }
    } else {
      image := imageOrAlbum
      
      meme := &MemeRow{}
      meme.Url = image.Link
      meme.NetUps = image.Points
      meme.Views = image.Views

      //fmt.Println("2", image)
      
      memeMetadata := image.MemeMetadata
      meme.TopText = memeMetadata.TopText
      meme.BottomText = memeMetadata.BottomText
      meme.MemeName = memeMetadata.MemeName
      meme.ImgurBgImage = memeMetadata.BgImage

      memes = append(memes, meme)
    }
  }

  for _, meme := range memes {
    fmt.Println(meme.Url, meme.NetUps, meme.Views, meme.MemeName, meme.TopText, meme.BottomText, meme.ImgurBgImage)
  }
}
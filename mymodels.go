package main

import (
	"encoding/json"
	"time"

	"github.com/tidusant/chadmin-repo/models"
	"gopkg.in/mgo.v2/bson"
)

type Template struct {
	ID           bson.ObjectId `bson:"_id,omitempty"`
	Code         string        `bson:"code"`
	UserID       string        `bson:"userid"`
	Status       int           `bson:"status"` //-2: delete, -1: reject, 1: approved and publish, 2: pending, 3: approved but not publish
	Title        string        `bson:"title"`
	Description  string        `bson:"description"`
	Viewed       int           `bson:"viewed"`
	InstalledIDs []string      `bson:"installedid"`
	ActiveIDs    []string      `bson:"activedid"`
	Configs      string        `bson:"configs"`
	Avatar       string        `bson:"avatar"`
	Created      time.Time     `bson:"created"`
	Modified     time.Time     `bson:"modified"`
	Content      string        `bson:"content"`
	Pages        string        `bson:"pages"`
	CSS          string        `bson:"css"`
	Script       string        `bson:"script"`
	Images       string        `bson:"images"`
	Fonts        string        `bson:"fonts"`
	Screenshot   string        `bson:"screenshot"`
	Langs        string        `bson:"langs"`
}

type RequestResult struct {
	Status  string          `json:"status"`
	Error   string          `json:"error"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}
type ViewData struct {
	PageName     string
	Siteurl      string
	Data         map[string]json.RawMessage
	TemplatePath string
	Templateurl  string
	Imageurl     string
	Pages        map[string]string
	Resources    map[string]string
	Configs      map[string]string
}

type NewsCat struct {
	Code        string `bson:"code"`
	Slug        string `bson:"slug"`
	Title       string `bson:"title"`
	Description string `bson:"description"`
	Content     string `bson:"content"`
}

//NewsLang ...
type News struct {
	CatId       string `bson:"catid"`
	Title       string `bson:"title"`
	Slug        string `bson:"slug"`
	Content     string `bson:"content"`
	Description string `bson:"description"`
	Avatar      string `bson:"avatar"`
	Viewed      int    `bson:"viewed"`
}
type ProdCat struct {
	Code        string `bson:"code"`
	Slug        string `bson:"slug"`
	Title       string `bson:"name"`
	Description string `bson:"description"`
	Content     string `bson:"conent"`
}
type Product struct {
	CatId           string `bson:"catid"`
	Name            string `bson:"name"`
	Slug            string `bson:"slug"`
	MaxPrice        int
	MinPrice        int
	BasePrice       int      `bson:"baseprice"`
	DiscountPrice   int      `bson:"discountprice"`
	PercentDiscount int      `bson:"percentdiscount"`
	Description     string   `bson:"description"`
	Content         string   `bson:"content"`
	Avatar          string   `bson:"avatar"`
	Images          []string `bson:"images"`
	Viewed          int      `bson:"viewed"`
	CatTitle        string
	CatSlug         string
	Properties      []models.ProductProperty
	NumInCart       int
}

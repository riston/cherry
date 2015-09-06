package cherry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	//BaseURL for main site
	BaseURL = "https://cherry.ee"

	// Categories is an url and category map, move this map to config file
	Categories = map[string]string{
		"Perfume":        fmt.Sprintf("%s%s", BaseURL, "/kaubad/parfuumid"),
		"BeautyProducts": fmt.Sprintf("%s%s", BaseURL, "/pakkumised/kaubamaja/ilutooted"),
		"SportsHealth":   fmt.Sprintf("%s%s", BaseURL, "/kaubad/spordi-ja-tervisetooted"),
		"HomeGarden":     fmt.Sprintf("%s%s", BaseURL, "/kaubad/kodu-ja-aed"),
		"FoodConsumer":   fmt.Sprintf("%s%s", BaseURL, "/kaubad/toidu-ja-tarbekaubad"),
		"Fashion":        fmt.Sprintf("%s%s", BaseURL, "/kaubad/moekaubad"),
		"Children":       fmt.Sprintf("%s%s", BaseURL, "/kaubad/lastekaubad-ja-manguasjad"),
	}
)

// Offer is single cherry structure with all the possible
// metadata available
type Offer struct {
	Title      string
	URL        string
	Price      float64
	PromoPrice float64
	Bought     int
	Limit      int
	Time       time.Duration
}

// ParseResult is high level container for the offers
type ParseResult struct {
	Category string
	Created  time.Time
	Offers   []*Offer
}

func (o Offer) String() string {

	return fmt.Sprintf("[%s - %0.2f - %0.2f bought %d limit '%d' time %d]",
		o.Title, o.Price, o.PromoPrice, o.Bought, o.Limit, o.Time)
}

// The timeleft is not in each page at the same location
// that is the reason why there is need to go through the elements
//
// It returns the string of javascript object
func getTimeCache(d *goquery.Document) string {

	filterScript := func(index int, s *goquery.Selection) bool {

		return strings.Contains(s.Text(), "timeleft_cache")
	}

	// Handle case when there is no element
	return d.Find("script").FilterFunction(filterScript).First().Text()
}

// Times are cached in string so to parse
// first needs a string clearing and then parse as a JSON
func parseTimeCache(cache string) map[string]int {

	m := map[string]int{}

	// First trim space
	cache = strings.TrimSpace(cache)

	// Trim the variable instance
	cache = strings.TrimLeft(cache, "timeleft_cache = ")

	err := json.Unmarshal([]byte(cache), &m)

	if err != nil {
		log.Println("Error on parsing timecache")
		return m
	}

	return m
}

func getTimerValue(s *goquery.Selection) string {

	timeVal := s.Find("span.timer").AttrOr("class", "")

	timeVal = strings.TrimSuffix(timeVal, " timer")
	timeVal = strings.TrimPrefix(timeVal, "timeleft_")

	return timeVal
}

func getTitle(s *goquery.Selection) string {

	return s.Find("h3 > a").Text()
}

func getURL(s *goquery.Selection) string {

	return s.Find("h3 > a").AttrOr("href", "")
}

func getPrice(s *goquery.Selection) (result float64) {

	fmt.Sscan(s.Find(".price > .actual").Text(), &result)
	return
}

func getPromotionPrice(s *goquery.Selection) (result float64) {

	fmt.Sscan(s.Find(".price > .promotion").Text(), &result)
	return
}

func getBoughtAmount(s *goquery.Selection) (bought int) {

	fmt.Sscan(s.Find(".amount > strong").Text(), &bought)
	return
}

func getLimit(s *goquery.Selection) (limit int) {

	limitText := s.Find(".limit").Text()

	num, err := strconv.Atoi(strings.TrimPrefix(limitText, "/ "))
	if err != nil {
		num = 0
	}
	return num
}

// Parse based on the goquery.Document object and return result
func Parse(category string, d *goquery.Document) *ParseResult {

	offers := []*Offer{}

	// The script tag holds object map of time
	timeCacheText := getTimeCache(d)

	timeIDMap := parseTimeCache(timeCacheText)

	d.Find(".box-green").Each(func(i int, s *goquery.Selection) {

		duration := time.Duration(int64(timeIDMap[getTimerValue(s)]))

		offers = append(offers, &Offer{
			getTitle(s),
			getURL(s),
			getPrice(s),
			getPromotionPrice(s),
			getBoughtAmount(s),
			getLimit(s),
			duration})
	})

	return &ParseResult{
		category,
		time.Now(),
		offers,
	}
}

// ParseFromReader parse result from io.Reader, could be example from file
func ParseFromReader(category string, r io.Reader) (*ParseResult, error) {

	doc, err := goquery.NewDocumentFromReader(r)

	if err != nil {
		return nil, errors.New("Failed to parse from io reader")
	}

	return Parse(category, doc), nil
}

// ParseFromResponse parse the result from HTTP response
func ParseFromResponse(category string, r *http.Response) (*ParseResult, error) {

	doc, err := goquery.NewDocumentFromResponse(r)

	if err != nil {
		return nil, errors.New("Failed to parse from HTTP response")
	}

	return Parse(category, doc), nil
}

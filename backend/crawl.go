package backend

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type Crawl struct {
	Url        string
	StatusCode int
	Type       string
	Size       int
	Age        string
	Redirect   string
	Body       string
	ExtraData  CrawlExtraData
}

type CrawlExtraData struct {
	RedirectLinks []string
	RedirectCount int
	Headers       map[string]string
}

var crawlResult []Crawl

func NewCrawl() *Crawl {
	return &Crawl{
		Url:        "",
		StatusCode: 0,
		Type:       "",
		Size:       0,
		Age:        "",
		Redirect:   "",
		ExtraData: CrawlExtraData{
			RedirectLinks: []string{},
			RedirectCount: 0,
			Headers:       make(map[string]string),
		},
	}
}

// NewCrawl creates a new Crawl application struct
func (c *Crawl) Crawl() string {

	return "Crawl"
}

func getUserAgent(userAgent string) string {
	if userAgent == "bot-mobile" {
		return "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html) (headofmastercemo)"
	} else if userAgent == "bot-desktop" {
		return "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Googlebot/2.1; +http://www.google.com/bot.html) Chrome/126.0.0.0 Safari/537.36 (headofmastercemo)"
	} else if userAgent == "mobile" {
		return "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Mobile Safari/537.36 (headofmastercemo)"
	} else if userAgent == "desktop" {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.85 Safari/537.36 (headofmastercemo)"

	}
	return "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html) (headofmastercemo)"
}

func (c *Crawl) StartCrawl(urls string, userAgent string) {
	client := resty.New().SetDebug(true)
	urlList := strings.Split(urls, "\n")
	clearCrawlResult()
	wp := NewWorkerPool(250)
	wp.Run()
	userAgent = getUserAgent(userAgent)
	for i := 0; i < len(urlList); i++ {
		url := urlList[i]
		wp.AddTask(func() {
			log.Printf("[main] Starting task %s", url)
			time.Sleep(500 * time.Millisecond)
			crawlData := Crawl{
				Url:       url,
				ExtraData: CrawlExtraData{},
			}
			redirects := []string{}

			client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
				fmt.Println("Making request to:", req.URL)
				if len(redirects) > 0 {
					fmt.Println("Redirected from:", redirects[len(redirects)-1])
				}
				return nil
			})

			client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
				if resp.StatusCode() >= 300 && resp.StatusCode() < 400 {
					location := resp.Header().Get("Location")
					if location != "" {
						redirects = append(redirects, location)
					}
				}
				return nil
			})

			resp, err := client.SetHeader("User-Agent", userAgent).R().Get(url)
			if err != nil {
				fmt.Println("Error:", err)
			}

			crawlData.Type = resp.Header().Get("Content-Type")
			crawlData.StatusCode = resp.StatusCode()
			crawlData.Size = len(resp.Body())
			crawlData.Age = resp.Header().Get("Age")

			crawlData.ExtraData.Headers = make(map[string]string)

			for key, v := range resp.Header() {
				crawlData.ExtraData.Headers[key] = v[0]
			}

			crawlData.Body = string(resp.Body())
			if resp.RawResponse.Request.Response != nil {
				redirects = append(redirects, resp.RawResponse.Request.URL.String())
				crawlData.ExtraData.RedirectCount = len(redirects)
				crawlData.ExtraData.RedirectLinks = redirects
				crawlData.Redirect = resp.RawResponse.Request.URL.String()
			}

			setCrawlResult(crawlData)
		})
	}

}

func setCrawlResult(crawl Crawl) {
	crawlResult = append(crawlResult, crawl)
}

func (c *Crawl) GetCrawlResults() []Crawl {
	return crawlResult
}

func clearCrawlResult() {
	crawlResult = []Crawl{}
}

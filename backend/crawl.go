package backend

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Crawl struct {
	Url        string
	StatusCode int
	Type       string
	Size       int
	Age        string
	Redirect   string
	Body       string          `csv:"-"`
	ExtraData  CrawlExtraData  `csv:"-"`
	ctx        context.Context `csv:"-"`
}

type CrawlExtraData struct {
	RedirectLinks []string
	RedirectCount int
	Headers       map[string]string
}

func NewCrawl(ctx context.Context) *Crawl {
	return &Crawl{
		Url:        "",
		StatusCode: 0,
		Type:       "",
		Size:       0,
		Age:        "",
		Redirect:   "",
		ctx:        ctx,
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

func fetchUrl(url string, userAgent string) *Crawl {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()
	var client = resty.New().SetRetryCount(3).NewRequest().SetContext(context.Background()).SetHeader("User-Agent", userAgent).SetHeader("Accept-Encoding", "gzip, deflate, br")
	// defer wg.Done()
	// sem.Lock()         // Acquire semaphore
	// defer sem.Unlock() // Release semaphore

	// log.Printf("[main] Starting task %s", url)
	crawlData := Crawl{
		Url:       url,
		ExtraData: CrawlExtraData{},
	}
	redirects := []string{}
	// log.Println("Crawling URL:", url)
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
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

	return &crawlData
}

func worker(urls <-chan string, results chan<- *Crawl, userAgent string, wg *sync.WaitGroup, delay time.Duration, batchSize int) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()
	for url := range urls {
		result := fetchUrl(url, userAgent)
		if result == nil {
			continue
		}

		results <- result
		time.Sleep(delay)
	}

}
func processURLs(urls []string, numWorkers int, requestDelay time.Duration, userAgent string, batchSize int) (<-chan *Crawl, <-chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()

	var wg sync.WaitGroup
	urlChan := make(chan string, len(urls))
	results := make(chan *Crawl, len(urls)/batchSize+1)
	done := make(chan struct{})
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(urlChan, results, userAgent, &wg, requestDelay, batchSize)
	}

	go func() {
		for _, url := range urls {
			urlChan <- url
		}
		close(urlChan)
	}()

	go func() {
		wg.Wait()
		close(done)
		close(results)
	}()

	return results, done
}

func StartCrawl(ctx context.Context, urls string, userAgent string) string {
	runtime.EventsOffAll(ctx)
	urlList := strings.Split(urls, "\n")
	userAgent = getUserAgent(userAgent)
	const batchSize = 1
	log.Println("URL List:", len(urlList))
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()

	const numWorkers = 50
	const requestDelay = 0 * time.Millisecond

	results, done := processURLs(urlList, numWorkers, requestDelay, userAgent, batchSize)
	id := uuid.New()

	go func() {
		for result := range results {
			runtime.EventsEmit(ctx, id.String(), result)
		}
	}()

	go func() {
		<-done
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Title:   "Crawl Completed",
			Message: "Crawl completed successfully",
		})
	}()

	return id.String()
}

func CancelFetch(ctx context.Context) {
	runtime.EventsOffAll(ctx)
}

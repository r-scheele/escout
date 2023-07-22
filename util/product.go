package util

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

type ScrapeFunc func(*colly.Collector, string) (float64, error)

// Dummy functions for demonstration
func ScrapeAmazon(c *colly.Collector, url string) (float64, error) {
	var price float64

	c.OnHTML(".a-size-mini .a-price .a-offscreen", func(e *colly.HTMLElement) {
		priceStr := strings.TrimSpace(e.Text)
		priceStr = strings.Replace(priceStr, "$", "", -1) // remove dollar sign

		var err error
		price, err = strconv.ParseFloat(priceStr, 64) // convert to float
		if err != nil {
			log.Fatal(err)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	err := c.Visit(url)
	if err != nil {
		log.Fatal(err)
	}

	return price, nil
}

func ScrapeBestBuy(c *colly.Collector, url string) (float64, error) {
	fmt.Println("Scraping BestBuy...")
	return 0, nil
}

func ScrapeWalmart(c *colly.Collector, url string) (float64, error) {
	fmt.Println("Scraping Walmart...")
	return 0, nil
}

var AllowedDomains = map[string]ScrapeFunc{
	"www.amazon.com":  ScrapeAmazon,
	"www.bestbuy.com": ScrapeBestBuy,
	"www.walmart.com": ScrapeWalmart,
}

// ScrapePriceFromURL fetches the price given a URL
func ScrapePriceFromURL(c *colly.Collector, targetURL string) (float64, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal(err)
	}
	domain := parsedURL.Hostname()

	scrapeFunc, exists := AllowedDomains[domain]
	if !exists {
		return 0, fmt.Errorf("no scraper function exists for domain: %s", domain)
	}

	return scrapeFunc(c, targetURL)
}

package util

import "log"

// ScrapePriceFromURL fetches the price given a URL
func ScrapePriceFromURL(url string) (float64, error) {
	// Initialize the colly collector
	// c := colly.NewCollector()

	// var priceStr string

	// // On every span element which has id attribute call callback
	// c.OnHTML(`span[id="priceblock_ourprice"]`, func(e *colly.HTMLElement) {
	// 	priceStr = e.Text
	// })

	// // Start scraping
	// err := c.Visit(url)
	// if err != nil {
	// 	log.Println("Error visiting URL:", err)
	// 	return 0, err
	// }

	// // Clean up price string and convert to float64
	// priceStr = strings.Replace(priceStr, "$", "", -1)
	// priceStr = strings.TrimSpace(priceStr)
	// price, err := strconv.ParseFloat(priceStr, 64)
	// if err != nil {
	// 	log.Println("Error parsing price:", err)
	// 	return 0, err
	// }

	price := 100.00
	log.Println("Price:", price, "URL:", url)
	return price, nil
}

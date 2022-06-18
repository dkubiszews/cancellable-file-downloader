package links_extractor

import (
	"github.com/gocolly/colly/v2"
)

func From_url(url_string string) []string {
	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	var result []string
	c.OnRequest(func(r *colly.Request) {
		request_url_string := r.URL.String()
		if request_url_string != url_string {
			result = append(result, request_url_string)
		}
	})

	c.Visit(url_string)
	return result
}

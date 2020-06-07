package crawler

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
	"net/http"
)

func FetchCities(url string) (map[string]string, error){
	// get the raw html data
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Could not get resp from url", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		fmt.Println("Bad status code:", resp.StatusCode)
	}

	// encode to utf-8 to deal with Chinese
	utfBody, err := iconv.NewReader(resp.Body, "gb2312", "utf-8")

	// parse this html to document
	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		fmt.Println("Could not parse response:", err)
		return nil, err
	}

	// selector to locate the <tr> that contains city & province
	selector := "tbody > tr > td:nth-child(3) > table > tbody > " +
		"tr > td:nth-child(1) > table > tbody > tr > td > table > tbody > tr"
	count := 0
	cityMap := make(map[string]string)

	doc.Find(selector).Each(func(i int, selection *goquery.Selection) {
		// select the city and province in each <tr>
		city := selection.Find("td:nth-child(2) > a").Text()
		province := selection.Find("td:nth-child(3) > a").Text()

		if city != "" {
			// save it to map
			count += 1
			cityMap[city] = province
		}
	})

	fmt.Println("Total city items collected:", count)

	return cityMap, err
}
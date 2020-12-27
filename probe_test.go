package probe

import (
	"fmt"
	"net/http"
	"testing"
)

func TestProbeFetchHtml(t *testing.T) {
	p := NewProbe()
	p.GenURL = func(urlChan chan string) {
		urlChan <- "https://www.google.com"
	}
	p.OnHTML = func(doc *Document) {
		doc.Find("a").Each(func(i int, s *Selection) {
			txt := s.Text()
			url, _ := s.Attr("href")
			fmt.Println(txt, url)
		})
	}
	p.Run()
}
func TestProbeFetchJson(t *testing.T) {
	p := NewProbe()
	p.GenURL = func(urlChan chan string) {
		urlChan <- "http://api.bilibili.com/x/relation/stat?vmid=1850091"
	}
	p.OnJSON = func(json map[string]interface{}) {
		fmt.Println(json)
	}
	p.Run()
}
func TestHeader(t *testing.T) {
	p := NewProbe()
	p.settings.header.Set("User-Agent", "test")
	p.GenURL = func(urlChan chan string) {
		urlChan <- "https://www.baidu.com"
	}
	p.OnRes = func(res http.Response) {
		ua := res.Request.Header.Get("User-Agent")
		fmt.Println(ua)
		if ua != "test" {
			t.Fail()
		}
	}
	p.Run()
}

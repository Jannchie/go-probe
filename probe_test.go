package probe

import (
	"fmt"
	"testing"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func TestProbeFetchHtml(t *testing.T) {
	p := NewProbe()
	p.GenURL = func(urlChan chan string) {
		urlChan <- "https://www.google.com"
	}
	p.OnHTML = func(doc *html.Node) {
		nodes := htmlquery.Find(doc, "//a/@href")
		for _, node := range nodes {
			fmt.Println(htmlquery.SelectAttr(node, "href"))
		}
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

package main

import (
	"fmt"
	"net/http"

	"github.com/jannchie/go-probe/probe"
)

func main() {
	p := probe.NewProbe()
	p.GenURL = func(urlChan chan string) {
		for i := 0; i < 1; i++ {
			urlChan <- "https://www.google.com"
		}
	}
	p.OnRes = func(res http.Response) {
	}
	p.OnJSON = func(json interface{}) {
		fmt.Println(json)
	}
	p.Run()
}

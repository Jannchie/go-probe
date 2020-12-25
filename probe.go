package probe

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type stat struct {
	urlSucceedCount int
	urlFailedCount  int
	endTime         time.Time
	startTime       time.Time
}

type settings struct {
	header http.Header
}

// Probe probe
type Probe struct {
	settings settings
	GenURL   func(urlChan chan string)
	OnRes    func(res http.Response)
	OnJSON   func(json map[string]interface{})
	OnHTML   func(html *html.Node)
	guard    chan struct{}
	done     chan struct{}
	urlChan  chan string
	resChan  chan http.Response
	client   *http.Client
	stat     stat
}

func (probe *Probe) runGenURLTask() {
	probe.GenURL(probe.urlChan)
	close(probe.urlChan)
}

func (probe *Probe) runDownloadTask() {
	wg := sync.WaitGroup{}
	for url := range probe.urlChan {
		probe.guard <- struct{}{}
		wg.Add(1)
		go probe.downloadTask(url, &wg)
	}
	wg.Wait()
	close(probe.resChan)
}

func (probe *Probe) downloadTask(url string, wg *sync.WaitGroup) {
	defer func() {
		<-probe.guard
		wg.Done()
	}()
	res, err := probe.getRes(url)
	if err != nil {
		log.Println(err)
		probe.stat.urlFailedCount++
		return
	}
	probe.resChan <- *res
	probe.stat.urlSucceedCount++
}

func (probe *Probe) getRes(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = probe.settings.header
	res, err := probe.client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (probe *Probe) runSaveDataTask() {
	for res := range probe.resChan {
		probe.OnRes(res)
		contentType := res.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			var j map[string]interface{}
			json.NewDecoder(res.Body).Decode(&j)
			probe.OnJSON(j)
		} else if strings.Contains(contentType, "text/html") {
			doc, err := html.Parse(res.Body)
			if err != nil {
				continue
			}
			probe.OnHTML(doc)
		}
		_ = res.Body.Close()
	}
	close(probe.done)
}

func (probe *Probe) printFinal() {
	deltaTime := time.Now().Sub(probe.stat.startTime)
	speed := float64(probe.stat.urlSucceedCount) / deltaTime.Minutes()
	fmt.Printf("Fetched URL: %d, Failed: %d [ %.2f%% ]\n", probe.stat.urlSucceedCount, probe.stat.urlFailedCount, probe.rate())
	fmt.Printf("Speed: %.2f req/min, URL: %d\n", speed, probe.stat.urlSucceedCount+probe.stat.urlFailedCount)
}

func (probe *Probe) runLoggingTask() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			fmt.Printf("Fetched URL: %d, Failed: %d [ %.2f%% ]\n", probe.stat.urlSucceedCount, probe.stat.urlFailedCount, probe.rate())
		case <-probe.done:
			return
		}
	}
}

func (probe *Probe) rate() float64 {
	rate := 0.0
	if probe.stat.urlSucceedCount+probe.stat.urlFailedCount != 0 {
		rate = float64(probe.stat.urlSucceedCount) / float64(probe.stat.urlSucceedCount+probe.stat.urlFailedCount) * 100
	}
	return rate
}

// Run run the probe
func (probe *Probe) Run() {
	probe.stat.startTime = time.Now()
	go probe.runGenURLTask()
	go probe.runDownloadTask()
	go probe.runSaveDataTask()
	probe.runLoggingTask()
	probe.printFinal()
}

// NewProbe generates new Probe
func NewProbe() *Probe {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	h := http.Header{}
	h.Set("User-Agent", "Probe ver.0.0.1")
	return &Probe{
		client:  &http.Client{},
		urlChan: make(chan string),
		resChan: make(chan http.Response),
		guard:   make(chan struct{}, 128),
		done:    make(chan struct{}),
		GenURL: func(urlChan chan string) {
			log.Fatal("Please implement the function: GenURL")
		},
		OnRes:  func(res http.Response) {},
		OnJSON: func(json map[string]interface{}) {},
		OnHTML: func(html *html.Node) {},
		settings: settings{
			header: h,
		},
	}
}

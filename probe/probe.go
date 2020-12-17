package probe

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type stat struct {
	urlSucceedCount int
	urlFailedCount  int
	endTime         time.Time
	startTime       time.Time
}

// Probe probe
type Probe struct {
	probeGroup    *sync.WaitGroup
	downloadGroup *sync.WaitGroup
	urlChannel    chan string
	resChannel    chan http.Response
	concurrency   int
	stat          stat
	temp          int
}

// runGenURLTask generates url
func (probe *Probe) runGenURLTask() {
	defer probe.probeGroup.Done()
	defer func() {
		for {
			if len(probe.urlChannel) == 0 {
				close(probe.urlChannel)
				break
			}
		}
	}()
	probe.GenURL()
}

// GenURL generates urls
func (probe *Probe) GenURL() {
	for i := 0; i < 500; i++ {
		probe.urlChannel <- "https://www.google.com"
	}
}

func (probe *Probe) runDownloadTask() {
	defer probe.probeGroup.Done()
	defer close(probe.resChannel)
	defer probe.downloadGroup.Wait()
	probe.downloadGroup.Add(probe.concurrency)
	for i := 0; i < probe.concurrency; i++ {
		go probe.downloadTask(i)
	}
}

func (probe *Probe) downloadTask(i int) {
	defer probe.downloadGroup.Done()
	for url := range probe.urlChannel {
		res, err := getRes(url)
		if err != nil {
			log.Println(err)
			probe.stat.urlFailedCount++
			continue
		}
		probe.resChannel <- *res
		probe.stat.urlSucceedCount++
	}
}

func getRes(url string) (*http.Response, error) {

	client := http.Client{}
	header := http.Header{}
	header.Set("User-Agent", "probe 0.0.1")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{
		Name:       "",
		Value:      "",
		Path:       "",
		Domain:     "",
		Expires:    time.Time{},
		RawExpires: "",
		MaxAge:     0,
		Secure:     false,
		HttpOnly:   false,
		SameSite:   0,
		Raw:        "",
		Unparsed:   []string{},
	})
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (probe *Probe) runSaveDataTask() {

	defer probe.probeGroup.Done()
	for res := range probe.resChannel {
		defer res.Body.Close()
		probe.Save(res)
	}
}

// Save saves data
func (probe *Probe) Save(res http.Response) {
	// fmt.Println(res.Status)
	if probe.stat.urlSucceedCount == 1 {
		_, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return
		}
	}
}

func (probe *Probe) printFinal() {
	deltaTime := time.Now().Sub(probe.stat.startTime)
	speed := float64(probe.stat.urlSucceedCount) / deltaTime.Minutes()
	fmt.Printf("Speed: %.2f req/min, URL: %d\n", speed, probe.stat.urlSucceedCount+probe.stat.urlFailedCount)
}

func (probe *Probe) runLoggingTask() {
	for {
		var rate float64
		if probe.stat.urlSucceedCount+probe.stat.urlFailedCount != 0 {
			rate = float64(probe.stat.urlSucceedCount) / float64(probe.stat.urlSucceedCount+probe.stat.urlFailedCount) * 100
		} else {
			rate = 0
		}
		fmt.Printf("Fetched URL: %d, Failed: %d [ %.2f%% ]\n", probe.stat.urlSucceedCount, probe.stat.urlFailedCount, rate)
		time.Sleep(1 * time.Second)
	}
}

// Run run the probe
func (probe *Probe) Run() {
	probe.stat.startTime = time.Now()
	probe.probeGroup.Add(3)
	defer probe.printFinal()
	defer probe.probeGroup.Wait()
	go probe.runGenURLTask()
	go probe.runDownloadTask()
	go probe.runSaveDataTask()
	go probe.runLoggingTask()
}

// NewProbe generates new Probe
func NewProbe() *Probe {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return &Probe{
		urlChannel:    make(chan string, 16),
		resChannel:    make(chan http.Response, 16),
		probeGroup:    &sync.WaitGroup{},
		downloadGroup: &sync.WaitGroup{},
		concurrency:   128,
	}
}

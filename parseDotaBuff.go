package main

import (
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

type vsData struct {
	Adv, WR float64
}

type heroData struct {
	Hero int
	Data []vsData
}

func hasAttr(t html.Token, name, value string) bool {
	for _, v := range t.Attr {
		if v.Key == name && v.Val == value {
			return true
		}
	}
	return false
}

func GetMainPage() (HeroList map[string]int, heroesURLs map[string]string) {
	dotaBuffURL, err := url.Parse("https://www.dotabuff.com/heroes/")
	if err != nil {
		log.Panicln(err)
	}

	resp, err := http.Get(dotaBuffURL.String())
	if err != nil {
		log.Fatal(err)
	}

	HeroList = make(map[string]int)
	heroesURLs = make(map[string]string, 130)

	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)

	heroN := 0
	var startedTable bool = false
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if z.Err() == io.EOF {
				break
			} else {
				log.Panicln(z.Err())
			}
		}

		if tt != html.StartTagToken {
			continue
		}

		t := z.Token()
		if t.Data == "div" {
			if hasAttr(t, "class", "hero-grid") {
				startedTable = true
			}
		}
		if startedTable && t.Data == "a" {
			href := ""
			for _, v := range t.Attr {
				if v.Key == "href" {
					href = v.Val
					break
				}
			}
			if strings.Contains(href, "/heroes/") {
				Link, err := url.Parse(href)
				if err != nil {
					log.Panicln(err)
				}
				heroURL := dotaBuffURL.ResolveReference(Link)
				ind := strings.LastIndex(heroURL.String(), "/")
				heroName := heroURL.String()[ind+1:]
				heroesURLs[heroName] = heroURL.String()
				HeroList[heroName] = heroN
				heroN++
			}
		}

	}
	return HeroList, heroesURLs
}

func ParseHeroPage(HeroList map[string]int, heroID int, URL string, resCh chan<- heroData) {
	resp, err := http.Get(URL + "/matchups")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	res := heroData{Hero: heroID, Data: make([]vsData, len(HeroList))}

	z := html.NewTokenizer(resp.Body)
	var tableStarted bool = false
	curHero := -1
	tddatavalueN := 0
loop:
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if z.Err() == io.EOF {
				break
			} else {
				log.Panicln(z.Err())
			}
		}
		if tt != html.StartTagToken {
			continue
		}

		t := z.Token()
		if t.Data == "article" {
			tableStarted = true
		}
		if tableStarted && t.Data == "tr" {
			herolink := ""
			for _, v := range t.Attr {
				if v.Key == "data-link-to" {
					herolink = v.Val
					break
				}
			}
			ind := strings.LastIndex(herolink, "/")
			if ind != 1 {
				curHero = HeroList[herolink[ind+1:]]
				tddatavalueN = 0
			}
		}
		if tableStarted && curHero != -1 && t.Data == "td" {
			V := 0.0
			found := false
			for _, v := range t.Attr {
				if v.Key == "data-value" {
					V, err = strconv.ParseFloat(v.Val, 64)
					if err != nil {
						continue loop
					}
					found = true
					break
				}
			}
			if found {
				switch tddatavalueN {
				case 0:
					res.Data[curHero].Adv = V
				case 1:
					res.Data[curHero].WR = V
				}
				tddatavalueN++
			}
		}

	}
	resCh <- res
}

func GetDotaBuffData() (HeroList map[string]int, AllData [][]vsData) {
	const parseLimit = 10

	wCh := make(chan bool, parseLimit)
	resCh := make(chan heroData)

	HeroList, heroURLs := GetMainPage()

	AllData = make([][]vsData, len(HeroList))

	alldataread := make(chan bool)
	go func() {
		for HD := range resCh {
			AllData[HD.Hero] = HD.Data
		}
		alldataread <- true
	}()

	var wg sync.WaitGroup
	for Name, Num := range HeroList {
		wg.Add(1)
		wCh <- true
		go func(heroID int, URL string, res chan<- heroData) {
			ParseHeroPage(HeroList, heroID, URL, res)
			<-wCh
			wg.Done()
		}(Num, heroURLs[Name], resCh)
	}
	wg.Wait()
	close(resCh)
	close(wCh)
	<-alldataread
	return HeroList, AllData
}

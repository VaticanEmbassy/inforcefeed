package main

import (
	"os"
	"fmt"
	"path"
	"regexp"
	"strings"
	"net/http"
	"io/ioutil"
	"github.com/VaticanEmbassy/inforcefeed/cfg"
	"github.com/mmcdole/gofeed"
)


type Fetcher struct {
	config *cfg.Config
}


func (f *Fetcher) fetchPaste(id string) bool {
	resp, err := http.Get(fmt.Sprintf(cfg.PASTEBIN_URL, id))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	if _, err := os.Stat(f.config.Outdir); os.IsNotExist(err) {
		os.MkdirAll(f.config.Outdir, 0755)
	}
	destFile := path.Join(f.config.Outdir, id)
	err = ioutil.WriteFile(destFile, body, 0644)
	if err != nil {
		return false
	}
	return true
}


func (f *Fetcher) Run() {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(f.config.Feed)
	reLink := regexp.MustCompile(cfg.PASTEBIN_LINK)
	groupNames := reLink.SubexpNames()
	seen := map[string]bool{}
	for _, item := range feed.Items {
		for _, match := range reLink.FindAllStringSubmatch(item.Content, -1) {
			for groupIdx, group := range match {
				name := groupNames[groupIdx]
				if name != "path" {
					continue
				}
				var id string
				pathPieces := strings.Split(group, "/")
				id = pathPieces[0]
				if len(pathPieces) > 1 && pathPieces[1] != "" {
					id = pathPieces[1]
				}
				if id == "" {
					continue
				}
				if _, ok := seen[id]; ok {
					continue
				}
				url := fmt.Sprintf(cfg.PASTEBIN_URL, id)
				fmt.Printf("%s\n", url)
				seen[id] = true
				if !f.config.Dry {
					f.fetchPaste(id)
				}
			}
		}
	}
}


func NewFetcher(config *cfg.Config) *Fetcher {
	f := new(Fetcher)
	f.config = config
	return f
}


func main() {
	config := cfg.ReadArgs()

	fetcher := NewFetcher(config)
	fetcher.Run()
}

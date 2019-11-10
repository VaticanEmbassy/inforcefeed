package main

import (
	"os"
	"fmt"
	"log"
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
		log.Println(err)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}
	if _, err := os.Stat(f.config.Outdir); os.IsNotExist(err) {
		os.MkdirAll(f.config.Outdir, 0755)
	}
	destFile := path.Join(f.config.Outdir, id)
	err = ioutil.WriteFile(destFile, body, 0644)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}


func (f *Fetcher) writeLastSeen(guid string) bool {
	if _, err := os.Stat(f.config.Outdir); os.IsNotExist(err) {
		os.MkdirAll(f.config.Outdir, 0755)
	}
	lsFile := path.Join(f.config.Outdir, ".last_guid")
	err := ioutil.WriteFile(lsFile, []byte(guid), 0644)
	if err != nil {
		return false
	}
	return true
}


func (f *Fetcher) readLastSeen() string {
	lsFile := path.Join(f.config.Outdir, ".last_guid")
	if _, err := os.Stat(lsFile); os.IsNotExist(err) {
		return ""
	}
	file, err := os.Open(lsFile)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		return ""
	}
	return strings.TrimSpace(string(b))
}


func (f *Fetcher) writeHistory(guid string, title string, url string) bool {
	if _, err := os.Stat(f.config.Outdir); os.IsNotExist(err) {
		os.MkdirAll(f.config.Outdir, 0755)
	}
	hFile := path.Join(f.config.Outdir, ".history")
	fd, err := os.OpenFile(hFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	defer fd.Close()
	line := fmt.Sprintf("%s\t%s\t%s\n", guid, title, url)
	if _, err = fd.WriteString(line); err != nil {
		return false
	}
	return true
}


func (f *Fetcher) Run() int {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(f.config.Feed)
	reLink := regexp.MustCompile(cfg.PASTEBIN_LINK)
	groupNames := reLink.SubexpNames()
	seen := map[string]bool{}
	count := 0
	for idx, item := range feed.Items {
		lastSeenGUID := f.readLastSeen()
		if lastSeenGUID != "" && item.GUID == lastSeenGUID {
			break
		}
		if idx == 0 && !f.config.Dry && item.GUID != "" {
			f.writeLastSeen(item.GUID)
		}
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
					if !f.fetchPaste(id) {
						continue
					}
				}
				f.writeHistory(id, item.Title, url)
				count += 1
			}
		}
	}
	return count
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

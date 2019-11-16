package main

import (
	"os"
	"fmt"
	"log"
	"path"
	"bufio"
	"regexp"
	"strings"
	"net/http"
	"io/ioutil"
	"github.com/VaticanEmbassy/pastebinimport/cfg"
	"github.com/mmcdole/gofeed"
)

var reNewlines = regexp.MustCompile(`[\r\n\t]`)


type Paste struct {
	Id string
	Title string
	Url string
}


func (p *Paste) String() string {
	return fmt.Sprintf("%s\t%s\t%s",
			reNewlines.ReplaceAllString(p.Id, " "),
			reNewlines.ReplaceAllString(p.Title, " "),
			reNewlines.ReplaceAllString(p.Url, " "))
}


func NewPaste(id string, title string, url string) Paste {
	p := Paste{}
	p.Id = id
	p.Title = title
	p.Url = url
	return p
}


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


func (f *Fetcher) writeHistoryLine(paste Paste) bool {
	if _, err := os.Stat(f.config.Outdir); os.IsNotExist(err) {
		os.MkdirAll(f.config.Outdir, 0755)
	}
	hFile := path.Join(f.config.Outdir, ".history")
	fd, err := os.OpenFile(hFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	defer fd.Close()
	line := fmt.Sprintf("%s\n", paste.String())
	if _, err = fd.WriteString(line); err != nil {
		return false
	}
	return true
}


func (f *Fetcher) readHistory() map[string]Paste {
	hist := make(map[string]Paste)
	if _, err := os.Stat(f.config.Outdir); os.IsNotExist(err) {
		return hist
	}
	hFile := path.Join(f.config.Outdir, ".history")
	fd, err := os.OpenFile(hFile, os.O_RDONLY, 0644)
	defer fd.Close()
	if err != nil {
		return hist
	}
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		pieces := strings.SplitN(line, "\t", 3)
		if (len(pieces) != 3) {
			continue
		}
		paste := NewPaste(pieces[0], pieces[1], pieces[2])
		hist[pieces[0]] = paste
	}
	return hist
}


func (f *Fetcher) Run() int {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(f.config.Feed)
	reLink := regexp.MustCompile(cfg.PASTEBIN_LINK)
	groupNames := reLink.SubexpNames()
	count := 0
	history := f.readHistory()
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
				if _, ok := history[id]; ok {
					continue
				}
				url := fmt.Sprintf(cfg.PASTEBIN_URL, id)
				fmt.Printf("%s\n", url)
				paste := NewPaste(id, item.Title, url)
				history[id] = paste
				if !f.config.Dry {
					if !f.fetchPaste(id) {
						continue
					}
				}
				f.writeHistoryLine(paste)
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

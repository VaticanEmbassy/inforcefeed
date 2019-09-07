package main

import (
	"fmt"
	"regexp"
	"github.com/mmcdole/gofeed"
)

var INFORGE_FORUM = "https://www.inforge.net/forum/forums/passwords-dump.1565/index.rss"
var PASTEBIN_LINK = "src=\"//pastebin.com/embed_js/(.+)/noheader\""
var PASTEBIN_URL = "https://pastebin.com/embed_js/%s/noheader"


func main() {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(INFORGE_FORUM)
	fmt.Println(feed.Title)
	reLink := regexp.MustCompile(PASTEBIN_LINK)
	for _, item := range feed.Items {
		//fmt.Println(item.GUID)
		link := reLink.FindAllStringSubmatch(item.Content, -1)
		if link == nil {
			continue
		}
		fmt.Println(fmt.Sprintf(PASTEBIN_URL, link[0][1]))
	}
}

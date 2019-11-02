package cfg

import (
	"flag"
)

var INFORGE_FORUM = "https://www.inforge.net/forum/forums/passwords-dump.1565/index.rss"
var PASTEBIN_LINK = "(?i)pastebin.com/(?P<path>[a-zA-Z0-9_/-]+)"
var PASTEBIN_URL = "https://pastebin.com/raw/%s"

type Config struct {
	Feed string
	Outdir string
	Dry bool
}

func ReadArgs() (*Config) {
	c := Config{}
	flag.StringVar(&c.Feed, "feed", INFORGE_FORUM,
			"URL of RSS or Atom feed to parse")
	flag.StringVar(&c.Outdir, "output-dir", "dumps",
			"Directory used to store the fetched files")
	flag.BoolVar(&c.Dry, "dry-run", false,
			"do not fetch files, just print the URLs")
	flag.Parse()
	return &c
}

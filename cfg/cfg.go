package cfg

import (
	"flag"
)

var INFORGE_FORUM = "https://www.inforge.net/forum/forums/passwords-dump.1565/index.rss"
var PASTEBIN_LINK = "(?i)pastebin.com/(?P<path>[a-zA-Z0-9_/-]+)"
var PASTEBIN_URL = "https://pastebin.com/raw/%s"
var USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36"

type Config struct {
	Feed string
	Outdir string
	UserAgent string
	Dry bool
	Verbose bool
}

func ReadArgs() (*Config) {
	c := Config{}
	flag.StringVar(&c.Feed, "feed", INFORGE_FORUM,
			"URL of RSS or Atom feed to parse")
	flag.StringVar(&c.Outdir, "output-dir", "dumps",
			"Directory used to store the fetched files")
	flag.StringVar(&c.UserAgent, "user-agent", USER_AGENT,
			"set the User-Agent for the HTTP requests")
	flag.BoolVar(&c.Dry, "dry-run", false,
			"do not fetch files, just print the URLs")
	flag.BoolVar(&c.Verbose, "verbose", false,
			"be more verbose")
	flag.Parse()
	return &c
}

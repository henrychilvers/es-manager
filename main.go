package main

// From: http://olivere.github.io/elastic/

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
)

var (
	version       = flag.Int("v", 6, "Version of ES at target URL (currently only 6 or 2 are supported)")
	url           = flag.String("url", "https://your.elasticsearch.com:9200", "ES URL (e.g. 'http://192.168.2.10:9200')")
	ips           = flag.String("ip", "kpi-,data-,synthetic-", "comma-separated list of Index name prefixes (e.g. 'kpi-,data-'")
	lessThanDate  = flag.String("lt", "", "Look for indexes before this cut-off date (e.g. 2018-10-26), or the number of days to go back (e.g. 14)")
	performDelete = flag.Bool("d", false, "Delete? (if not provided, just check for Empty and Future indexes)")

	indexPrefixes    = make([]string, 0)
	indexesToProcess = make([]string, 0)

	futureIndexCount = 0
	emptyIndexCount  = 0
	oldIndexCount    = 0
)

func main() {
	setFlags()

	var strategy VersionedESHandlingStrategy

	if *version == 6 {
		strategy = V6VersionedESHandlingStrategy{}
	} else {
		strategy = V2VersionedESHandlingStrategy{}
	}

	fmt.Printf("Inspecting %s\n", *url)

	strategy.Process(*url)
}

func setFlags() {
	flag.Parse()

	if *url == "" {
		log.Fatal("no url specified")
	}

	if *ips == "" {
		log.Fatal("no Index prefix(es) specified")
	}
	indexPrefixes = strings.SplitN(*ips, ",", -1)
	sort.Strings(indexPrefixes)
}

func ShouldEvaluateIndex(index string, indexesToProcess []string) bool {
	for _, v := range indexesToProcess {
		if strings.HasPrefix(index, v) {
			return true
		}
	}

	return false
}

// func GetRunningMode() string {
// 	if *performDelete {
// 		return "Deleting"
// 	} else {
// 		return "Analyzing"
// 	}
// }

func GetCompletedAction() string {
	if *performDelete {
		return "Deleted"
	} else {
		return "Analyzed"
	}
}

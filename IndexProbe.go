package main

import (
	"errors"
	"fmt"
	"gopkg.in/olivere/elastic.v2"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func NewIndexProbe(url string, delete bool) (*IndexProbe, error) {
	return &IndexProbe{
		nodes:               []string{url},
		performDelete:       delete,
		errorlogfile:        "error_log.txt",
		infologfile:         "info_log.txt",
		tracelogfile:        "trace_log.txt",
		healthcheck:         true,
		healthcheckInterval: elastic.DefaultHealthcheckInterval,
		sniff:               true,
		snifferInterval:     elastic.DefaultSnifferInterval,
		runCh:               make(chan RunInfo),
	}, nil
}

type IndexProbe struct {
	nodes               []string
	client              *elastic.Client
	performDelete       bool
	runs                int64
	failures            int64
	runCh               chan RunInfo
	index               string
	indexes             []string
	errorlogfile        string
	infologfile         string
	tracelogfile        string
	maxRetries          int
	healthcheck         bool
	healthcheckInterval time.Duration
	sniff               bool
	snifferInterval     time.Duration
}

func (t *IndexProbe) SetIndex(name string) {
	t.index = name
}

func (t *IndexProbe) SetPerformDelete(aORd bool) {
	t.performDelete = aORd
}

func (t *IndexProbe) SetErrorLogFile(name string) {
	t.errorlogfile = name
}

func (t *IndexProbe) SetInfoLogFile(name string) {
	t.infologfile = name
}

func (t *IndexProbe) SetTraceLogFile(name string) {
	t.tracelogfile = name
}

func (t *IndexProbe) SetSniff(enabled bool) {
	t.sniff = enabled
}

func (t *IndexProbe) SetSnifferInterval(d time.Duration) {
	t.snifferInterval = d
}

func (t *IndexProbe) SetHealthcheck(enabled bool) {
	t.healthcheck = enabled
}

func (t *IndexProbe) SetHealthcheckInterval(d time.Duration) {
	t.healthcheckInterval = d
}

func (t *IndexProbe) setup() error {
	var errorlogger *log.Logger
	if t.errorlogfile != "" {
		f, err := os.OpenFile(t.errorlogfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)

		if err != nil {
			return err
		}

		errorlogger = log.New(f, "", log.Ltime|log.Lmicroseconds|log.Lshortfile)
	}

	var infologger *log.Logger
	if t.infologfile != "" {
		f, err := os.OpenFile(t.infologfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)

		if err != nil {
			return err
		}

		infologger = log.New(f, "", log.LstdFlags)
	}

	// Trace request and response details like this
	var tracelogger *log.Logger
	if t.tracelogfile != "" {
		f, err := os.OpenFile(t.tracelogfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)

		if err != nil {
			return err
		}

		tracelogger = log.New(f, "", log.LstdFlags)
	}

	client, err := elastic.NewClient(
		elastic.SetURL(t.nodes...),
		elastic.SetErrorLog(errorlogger),
		elastic.SetInfoLog(infologger),
		elastic.SetTraceLog(tracelogger))

	if err != nil {
		// Handle error
		return err
	}
	t.client = client

	indexes, err := client.IndexNames()
	if err != nil {
		return err
	}
	t.indexes = indexes

	return nil
}

func (t *IndexProbe) Run() error {
	fmt.Printf("Inspecting index: %s...\n", t.index)

	if *lessThanDate == "" {
		if err := t.processEmptyIndexes(); err != nil {
			return err
		}

		if err := t.processFutureIndexes(); err != nil {
			return err
		}
	} else {
		if err := t.processOldIndexes(); err != nil {
			return err
		}
	}

	return nil
}

func (t *IndexProbe) processEmptyIndexes() error {
	// Use the IndexExists service to check if a specified index exists.
	exists, err := t.client.IndexExists(t.index).Do()
	if err != nil {
		return err
	}

	if exists {
		stats, err := t.client.IndexStats(t.index).Do()
		if err != nil {
			return err
		}

		if stats.All.Total.Docs.Count == 0 {
			if t.performDelete {
				deleteIndex, err := t.client.DeleteIndex(t.index).Do()
				if err != nil {
					return err
				}

				if !deleteIndex.Acknowledged {
					return errors.New("delete index not acknowledged")
				}
			}

			fmt.Printf("%s empty index: %s\n", GetCompletedAction(), t.index)
			emptyIndexCount++
		}
	}

	return nil
}

func (t *IndexProbe) processFutureIndexes() error {
	// Use the IndexExists service to check if a specified index exists.
	exists, err := t.client.IndexExists(t.index).Do()
	if err != nil {
		return err
	}

	if exists {
		datePart := strings.Index(t.index, "-")
		indexDate, _ := time.Parse(time.RFC3339, t.index[datePart+1:len(t.index)]+"T00:00:00Z")
		fmt.Printf("%s\n%s", indexDate, time.Now().Add(24*time.Hour))

		// If this index in in the future...
		if indexDate.After(time.Now()) {
			if t.performDelete {
				deleteIndex, err := t.client.DeleteIndex(t.index).Do()
				if err != nil {
					return err
				}

				if !deleteIndex.Acknowledged {
					return errors.New("delete index not acknowledged")
				}
			}

			fmt.Printf("%s future index: %s\n", GetCompletedAction(), t.index)
			futureIndexCount++
		}
	}

	return nil
}

func (t *IndexProbe) processOldIndexes() error {
	exists, err := t.client.IndexExists(t.index).Do()
	if err != nil {
		return err
	}

	if exists {
		datePart := strings.Index(t.index, "-")
		indexDate, _ := time.Parse(time.RFC3339, t.index[datePart+1:len(t.index)]+"T00:00:00Z")

		ltd, parse_error := time.Parse(time.RFC3339, *lessThanDate+"T00:00:00Z")

		if parse_error != nil {
			days, _ := strconv.Atoi(*lessThanDate)

			ltd = time.Now().Add(time.Duration(-days) * time.Hour).Truncate(24 * time.Hour)
		}

		fmt.Printf("using cutoff date of %s\n", ltd)

		// If this index in before the cut-off date...
		if indexDate.Before(ltd) {
			if t.performDelete {
				deleteIndex, err := t.client.DeleteIndex(t.index).Do()
				if err != nil {
					return err
				}

				if !deleteIndex.Acknowledged {
					return errors.New("delete index not acknowledged")
				}
			}

			fmt.Printf("%s old index: %s\n", GetCompletedAction(), t.index)
			oldIndexCount++
		}
	}

	return nil
}

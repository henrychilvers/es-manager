package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/olivere/elastic"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ctx = context.Background()
)

type V6VersionedESHandlingStrategy struct{}

func (V6VersionedESHandlingStrategy) Process(url string) {
	elasticClient, _ := elastic.NewClient(elastic.SetURL(url), elastic.SetSniff(false))

	info, code, err := elasticClient.Ping(url).Do(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)
	fmt.Println()

	// Getting the ES version number is quite common, so there's a shortcut
	// esversion, err := elasticClient.ElasticsearchVersion(url)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("Elasticsearch version %s\n", esversion)

	names, err := elasticClient.IndexNames()
	if err != nil {
		panic(err)
	}
	sort.Strings(names)

	fmt.Printf("Found %d Indexes...\n", len(names))

	for _, name := range names {
		if ShouldEvaluateIndex(name, indexPrefixes) {
			fmt.Printf("%s\n", name)

			indexesToProcess = append(indexesToProcess, name)
		}
	}

	ltd, parseError := time.Parse(time.RFC3339, *lessThanDate+"T00:00:00Z")

	if parseError != nil {
		days, _ := strconv.Atoi(*lessThanDate)

		ltd = time.Now().Add(time.Duration(-days) * time.Hour * 24).Truncate(24 * time.Hour)
	}

	fmt.Printf("using cutoff date of %s\n", ltd)

	for _, index := range indexesToProcess {
		err := process(elasticClient, index, ltd)

		if err != nil {
			panic(err)
		}
	}

	PrintResults()
}

func process(client *elastic.Client, index string, ltd time.Time) error {
	fmt.Printf("Inspecting index: %s... ", index)

	if *lessThanDate == "" {
		if err := processEmptyIndexes(client, index); err != nil {
			return err
		}

		if err := processFutureIndexes(client, index); err != nil {
			return err
		}
	} else {
		if err := processOldIndexes(client, index, ltd); err != nil {
			return err
		}
	}

	return nil
}

func processEmptyIndexes(client *elastic.Client, index string) error {
	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		return err
	}

	if exists {
		stats, err := client.IndexStats(index).Do(ctx)
		if err != nil {
			return err
		}

		if stats.All.Total.Docs.Count == 0 {
			if *performDelete {
				deleteIndex, err := client.DeleteIndex(index).Do(ctx)
				if err != nil {
					return err
				}

				if !deleteIndex.Acknowledged {
					return errors.New("delete index not acknowledged")
				}
			}

			fmt.Printf("%s\n", GetCompletedAction())
			emptyIndexCount++
		} else {
			fmt.Printf("Not Empty\n")
		}
	}

	return nil
}

func processFutureIndexes(client *elastic.Client, index string) error {
	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		return err
	}

	if exists {
		datePart := strings.Index(index, "-")
		indexDate, _ := time.Parse(time.RFC3339, index[datePart+1:]+"T00:00:00Z")

		// If this index in in the future...
		if indexDate.After(time.Now()) {
			if *performDelete {
				deleteIndex, err := client.DeleteIndex(index).Do(ctx)
				if err != nil {
					return err
				}

				if !deleteIndex.Acknowledged {
					return errors.New("delete index not acknowledged")
				}
			}

			fmt.Printf("%s\n", GetCompletedAction())
			futureIndexCount++
		} else {
			fmt.Printf("Not Future\n")
		}
	}

	return nil
}

func processOldIndexes(client *elastic.Client, index string, ltd time.Time) error {
	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		return err
	}

	if exists {
		datePart := strings.Index(index, "-")
		indexDate, _ := time.Parse(time.RFC3339, index[datePart+1:]+"T00:00:00Z")

		// If this index in before the cut-off date...
		if indexDate.Before(ltd) {
			if *performDelete {
				deleteIndex, err := client.DeleteIndex(index).Do(ctx)
				if err != nil {
					return err
				}

				if !deleteIndex.Acknowledged {
					return errors.New("delete index not acknowledged")
				}
			}

			fmt.Printf("%s\n", GetCompletedAction())
			oldIndexCount++
		} else {
			fmt.Printf("Not Old\n")
		}
	}

	return nil
}

package main

// From: http://olivere.github.io/elastic/
import (
	"log"
)

type V2VersionedESHandlingStrategy struct{}

func (V2VersionedESHandlingStrategy) Process(url string) {
	testcase, err := NewIndexProbe(url, *performDelete)
	if err != nil {
		log.Fatal(err)
	}

	if err := testcase.setup(); err != nil {
		log.Fatal(err)
	}

	for _, i := range testcase.indexes {
		if ShouldEvaluateIndex(i, indexesToProcess) {
			testcase.index = i

			if err := testcase.Run(); err != nil {
				log.Fatal(err)
			}
		}
	}

	PrintResults()
}

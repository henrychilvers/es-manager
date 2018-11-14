package main

type VersionedESHandlingStrategy interface {
	Process(url string)
}

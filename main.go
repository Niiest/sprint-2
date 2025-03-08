package main

import (
	"example/kafka/blocker"
	"example/kafka/emitter"
	"example/kafka/filter"
	"example/kafka/user"

	"github.com/lovoo/goka"
)

var brokers = []string{"127.0.0.1:9094"}

var (
	Emitter2FilterTopic       goka.Stream = "emitter2filter-stream"
	Filter2UserProcessorTopic goka.Stream = "filter2userprocessor-stream"
	BlockerTopic              goka.Stream = "blocker-stream"
)

func main() {
	go blocker.RunBlocker(brokers, BlockerTopic)
	go emitter.RunEmitter(brokers, Emitter2FilterTopic)
	go filter.RunFilter(brokers, Emitter2FilterTopic, Filter2UserProcessorTopic)
	go user.RunUserProcessor(brokers, Filter2UserProcessorTopic)
	user.RunUserView(brokers)
}

package main

import (
	"git.pnhub.ru/core/components/pharvester"
	"git.pnhub.ru/core/libs/base"
	"git.pnhub.ru/core/libs/log"
)

func main() {
	app := base.NewApplication([]base.ProviderFunction{
		log.DefaultLogger,
		pharvester.NewCrawler,
		pharvester.NewParser,
		pharvester.NewStorage,
		pharvester.NewValidator,
		pharvester.NewHarvester,
		pharvester.NewConfig,
	})
	app.Start(invoker)
}
func invoker(harvester *pharvester.Harvester) error {
	return harvester.Harvest()
}

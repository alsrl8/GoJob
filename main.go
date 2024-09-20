package main

import (
	"GoJob/web"
	"GoJob/xlog"
)

func main() {
	logger := xlog.NewXLogger("./app.log")
	defer func() {
		logger.Info("GoJob Stopped")
		logger.Close()
	}()

	xlog.Logger.Info("GoJob Started")

	web.CrawlJumpit()
}

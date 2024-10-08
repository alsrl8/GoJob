package main

import (
	"GoJob/view"
	"GoJob/xlog"
)

func main() {
	logger := xlog.NewXLogger()
	defer func() {
		logger.Info("GoJob Stopped")
		logger.Close()
	}()

	xlog.Logger.Info("GoJob Started")

	view.Init()
}

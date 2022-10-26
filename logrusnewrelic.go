package gqlserver

import (
	"log"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type LogrusNewrelicHook struct {
	nrApp *newrelic.Application
}

func NewLogrusNewrelicHook(nrApp *newrelic.Application) *LogrusNewrelicHook {
	return &LogrusNewrelicHook{
		nrApp: nrApp,
	}
}

func (h *LogrusNewrelicHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		log.Printf("NrHook failed to fire. Unable to read entry, %v", err)
		return err
	}

	h.nrApp.RecordLog(newrelic.LogData{
		Severity:  entry.Level.String(),
		Message:   line,
		Timestamp: entry.Time.UnixMilli(),
	})

	return nil
}

func (h *LogrusNewrelicHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

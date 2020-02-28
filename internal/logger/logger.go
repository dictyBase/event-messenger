package logger

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func NewLogger(c *cli.Context) (*logrus.Entry, error) {
	lfmt, err := getLogFmt(c)
	if err != nil {
		return &logrus.Entry{}, err
	}
	level, err := getLogLevel(c)
	if err != nil {
		return &logrus.Entry{}, err
	}
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetFormatter(lfmt)
	logger.SetLevel(level)
	return logrus.NewEntry(logger), nil
}

func getLogLevel(c *cli.Context) (logrus.Level, error) {
	var level logrus.Level
	switch c.String("log-level") {
	case "debug":
		level = logrus.DebugLevel
	case "info":
		level = logrus.InfoLevel
	case "warn":
		level = logrus.WarnLevel
	case "error":
		level = logrus.ErrorLevel
	case "fatal":
		level = logrus.FatalLevel
	case "panic":
		level = logrus.PanicLevel
	default:
		return level, fmt.Errorf(
			"%s log level is not supported",
			level,
		)
	}
	return level, nil
}

func getLogFmt(c *cli.Context) (logrus.Formatter, error) {
	var lfmt logrus.Formatter
	switch c.String("log-format") {
	case "text":
		lfmt = &logrus.TextFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	case "json":
		lfmt = &logrus.JSONFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	default:
		return lfmt, fmt.Errorf(
			"only json and text are supported %s log format is not supported",
			c.String("log-format"),
		)
	}
	return lfmt, nil
}

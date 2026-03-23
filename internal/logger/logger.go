package logger

import "log/slog"

type LogLevel string

var levelMap = map[LogLevel]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

var programLevel = &slog.LevelVar{}

func Init(cfgLevel LogLevel) {
	if level, ok := levelMap[cfgLevel]; ok {
		programLevel.Set(level)
	} else {
		programLevel.Set(levelMap["info"])
	}
}

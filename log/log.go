package log

import (
	"log"
	"os"

	"github.com/byuoitav/common/nerr"
	"github.com/fatih/color"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// L is our SugaredLogger
var L *zap.SugaredLogger
var cfg zap.Config
var atom zap.AtomicLevel

func init() {
	atom = zap.NewAtomicLevelAt(zapcore.WarnLevel)

	cfg = zap.NewDevelopmentConfig()

	cfg.OutputPaths = append(cfg.OutputPaths)
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.Level = atom

	l, err := cfg.Build()
	if err != nil {
		log.Print(color.HiRedString("Couldn't build config for zap logger: %v", err.Error()))
		panic(err)
	}

	L = l.Sugar()
	L.Info(color.HiYellowString("Zap Logger Started"))

	// look for log level env var
	// +deploy not_required
	level := os.Getenv("LOG_LEVEL")
	if len(level) > 0 {
		SetLevel(level)
		L.Info(color.HiYellowString("Set log level to %s", level))
	}
}

// SetLevel sets the log level
func SetLevel(level string) *nerr.E {
	switch level {
	case "debug":
		atom.SetLevel(zapcore.DebugLevel)
	case "info":
		atom.SetLevel(zapcore.InfoLevel)
	case "warn":
		atom.SetLevel(zapcore.WarnLevel)
	case "error":
		atom.SetLevel(zapcore.ErrorLevel)
	case "dpanic":
		atom.SetLevel(zapcore.DPanicLevel)
	case "panic":
		atom.SetLevel(zapcore.PanicLevel)
	case "fatal":
		atom.SetLevel(zapcore.FatalLevel)
	default:
		return nerr.Create("Invalid level", "invalid_args")
	}

	return nil
}

// GetLevel returns the current log level
func GetLevel() (string, *nerr.E) {
	return atom.Level().String(), nil
}

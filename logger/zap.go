package logger

import (
	"flowcontroller/config"
	"io"
	"slices"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate mockery --filename=mock_logger.go --name=Logger --dir=. --structname MockLogger  --inpackage=true
type Logger interface {
	Debug(args ...any)
	Info(args ...any)
	Error(args ...any)
	Fatal(args ...any)
}

//go:generate mockery --filename=mock_write_syncer.go --name=WriteSyncer --dir=. --structname MockWriteSyncer  --inpackage=true
type WriteSyncer interface {
	io.Writer
	Sync() error
}

// Attaches specific name to zap.AtomicLevel
type LevelWithName struct {
	zap.AtomicLevel
	name string
}

func withName(name string, level zap.AtomicLevel) LevelWithName {
	return LevelWithName{level, name}
}

// R_ONLY name
func (l LevelWithName) Name() string {
	return l.name
}

// []LevelWithName may be used to change specific output destination log levels
// Changing them in runtime is tread safe
func AssembleLogger(config config.Config) (Logger, []LevelWithName, error) {

	levels := make([]LevelWithName, 0, len(config.Logger.Cores))

	// TODO: Add remote dest support
	cores := make([]zapcore.Core, 0, len(config.Logger.Cores))

	// Iterating thorough config cores and creating zapcore.Cores out of them
	for _, core := range config.Logger.Cores {
		logDest, err := assembleDestination(string(core.Path))
		if err != nil {
			if core.MustCreateCore {
				return nil, nil, err
			}
			continue
		}
		encoder, err := setEncoder(core.EncoderLevel)
		if err != nil {
			return nil, nil, err
		}
		namedLevel := withName(core.Name, zap.NewAtomicLevelAt(zapcore.Level(core.Level)))
		levels = append(levels, namedLevel)
		cores = append(cores, zapcore.NewCore(
			encoder,               // production or development
			logDest,               // file or stderr/stdout // TODO Add remote dest support
			levels[len(levels)-1], // last level, every time
		))
	}
	levels = slices.Clip(levels)
	cores = slices.Clip(cores)
	if len(cores) == 0 {
		return nil, nil, ErrNoCoresWasInitialized
	}

	// Creating Sugar Logger from cores
	unifiedcore := zapcore.NewTee(cores...)
	logger := zap.New(unifiedcore).Sugar()

	// First log message
	// That tells us that logger construction succeeded
	logger.Debug("Logger construction succeeded")

	// TODO utilise returning stopFunc
	_ = syncOnTimout(logger, config.Logger.SyncTimeout)

	return logger, levels, nil
}

// ===================================================================================================================

func InitGlobalLogger(config config.Config) (err error) {
	panic("DEPRECATED GLOBAL LOGGER")
}
func Debug(args ...any) {
	panic("DEPRECATED GLOBAL LOGGER")
}
func Info(args ...any) {
	panic("DEPRECATED GLOBAL LOGGER")
}
func Fatal(args ...any) {
	panic("DEPRECATED GLOBAL LOGGER")

}

package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(cfg config.Config) zap.SugaredLogger {
	var logLVL zapcore.Level
	switch {
	case cfg.Logger.Level == "DEBUG":
		logLVL = zap.DebugLevel
	case cfg.Logger.Level == "INFO":
		logLVL = zap.InfoLevel
	}

	config := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(logLVL),
		OutputPaths:      append([]string{"stdout"}, cfg.Logger.Out...),
		ErrorOutputPaths: append([]string{"stderr"}, cfg.Logger.OutErr...),
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,

			TimeKey:    "date",
			EncodeTime: zapcore.TimeEncoderOfLayout("2006 Jan 15:04:05.000"),
		},
	}
	files := make([]io.Writer, 0, 16)
	errFiles := make([]io.Writer, 0, 16)
	for _, name := range cfg.Logger.Out {
		dir := filepath.Dir(name)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatalf("cannot open file: %s", name)
		}
		file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			log.Fatalf("cannot open file: %s", name)
		}
		files = append(files, file)
	}
	for _, name := range cfg.Logger.Out {
		dir := filepath.Dir(name)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatalf("cannot open file: %s", name)
		}
		fileErr, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			log.Fatalf("cannot open file: %s", name)
		}
		errFiles = append(errFiles, fileErr)
	}
	fw := io.MultiWriter(files...)
	ew := io.MultiWriter(errFiles...)

	ws := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fw))
	wsErr := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stderr), zapcore.AddSync(ew))

	logg, err := config.Build(
		zap.WrapCore(
			func(zapcore.Core) zapcore.Core {
				levelEnabler := zap.LevelEnablerFunc(
					func(lvl zapcore.Level) bool {
						return lvl < zap.ErrorLevel && lvl >= logLVL
					},
				)
				return zapcore.NewTee(
					zapcore.NewCore(zapcore.NewConsoleEncoder(config.EncoderConfig), ws, levelEnabler),
					zapcore.NewCore(zapcore.NewConsoleEncoder(config.EncoderConfig), wsErr, zapcore.ErrorLevel),
				)
			}),
	)
	if err != nil {
		log.Fatalf("cannot inititalize logger, error: %v\n", err)
	}
	logger := logg.Sugar()
	return *logger
}

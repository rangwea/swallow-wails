package backend

import (
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger
var Sugar *zap.SugaredLogger

type ZapConfig struct {
	Level        string `mapstructure:"level" json:"level" yaml:"level"`
	Directory    string `mapstructure:"directory" json:"directory"  yaml:"directory"`
	LogInConsole bool   `mapstructure:"log-in-console" json:"log-in-console" yaml:"log-in-console"`
	MaxAge       uint   `mapstructure:"max-age" json:"max-age" yaml:"max-age"`
}

var zapConfig = ZapConfig{
	Level:        "info",
	Directory:    "./logs",
	LogInConsole: true,
	MaxAge:       30,
}

var Zap = _zap{}

type _zap struct{}

func (z *_zap) Initialize() {
	// initialize the rotator
	logFile := filepath.Join(zapConfig.Directory, "app-%Y-%m-%d.log")
	rotator, err := rotatelogs.New(
		logFile,
		rotatelogs.WithRotationCount(zapConfig.MaxAge),
		rotatelogs.WithRotationTime(time.Hour*24))
	if err != nil {
		panic(err)
	}

	// add the encoder config and rotator to create a new zap logger
	var writer zapcore.WriteSyncer
	fileWriter := zapcore.AddSync(rotator)
	if zapConfig.LogInConsole {
		writer = zapcore.NewMultiWriteSyncer(fileWriter, zapcore.AddSync(os.Stdout))
	} else {
		writer = fileWriter
	}

	core := zapcore.NewCore(
		z.GetEncoder(),
		writer,
		zap.InfoLevel)

	Logger = zap.New(core)
	Sugar = Logger.Sugar()

	Logger.Info("Now logging inited")
}

// GetEncoder get zapcore.Encoder
func (z *_zap) GetEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(z.GetEncoderConfig())
}

// GetEncoderConfig get zapcore.EncoderConfig
func (z *_zap) GetEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		EncodeTime:     z.CustomTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
}

// CustomTimeEncoder custom the time format
func (z *_zap) CustomTimeEncoder(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
	encoder.AppendString(t.Format("2006/01/02 - 15:04:05.000"))
}

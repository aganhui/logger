package logger

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

const (
	defaultLogFile = "./logs/app"
)

var (
	confFile = "configs/configs.yaml"
	sugar    = &zap.SugaredLogger{} // skip 1
	sugar0   = &zap.SugaredLogger{} // skip 0
	logTag   = &LogTag{}
)

type LogTag struct {
	TaskID      string
	MachineAddr string
}

type LogConfig struct {
	LogLevel string `yaml:"log_level" json:"log_level"`
	LogFile  string `yaml:"log_file"  json:"log_file"`
}

func Sugar() *zap.SugaredLogger {
	return sugar
}

func init() {
	Init()
}

func main() {
	Init()
	return
}

func Init() {
	// init LogTag
	if cf := os.Getenv("LoggerConfFile"); cf != "" {
		fmt.Printf("change log conf file to : %v\n", cf)
		confFile = cf
	}

	// init log
	// log.Printf("config file: %s\n", confFile)
	file, err := ioutil.ReadFile(confFile)
	if err != nil {
		log.Printf("logger read config file %s err: %s", confFile, err.Error())
	}
	yamlString := string(file)
	// log.Printf("yaml string: %s\n", yamlString)

	lcfg := &LogConfig{}
	err = yaml.Unmarshal([]byte(yamlString), lcfg)
	// log.Printf("log file: %s\n", lcfg.LogFile)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if lcfg.LogLevel == "" {
		lcfg.LogLevel = zap.DebugLevel.String()
	}
	if lcfg.LogFile == "" {
		lcfg.LogFile = defaultLogFile
	}
	timestamp := time.Now().UTC().Format("20060102_150405.000")
	timepostfix := fmt.Sprintf(".%s.log", timestamp)
	lcfg.LogFile += timepostfix
	var l zapcore.Level
	err = l.UnmarshalText([]byte(lcfg.LogLevel))
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   lcfg.LogFile,
		MaxSize:    20, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})

	zconf := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		//		EncodeTime:     LoggerTimeEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zconf),
		w,
		l,
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	logger0 := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))

	sugar = logger.Sugar()
	sugar0 = logger0.Sugar()
}

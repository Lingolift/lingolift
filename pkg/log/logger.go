package log

import (
	_ "fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultLoggerMaxSize = 512

	defaultMaxAge = 10

	defaultMaxBackups = 100

	defaultCompress = false
)

type Options struct {
	// 文件保存地方
	LogFileDir string `yaml:"filename"`

	// 日志文件前缀
	AppName string `yaml:"app_name"`

	ErrorFileName string

	WarnFileName string

	InfoFileName string

	DebugFileName string

	// 日志等级
	Level zapcore.Level

	// 日志文件小大（M）
	MaxSize int `yaml:"max_size"`

	// 最多存在多少个切片文件
	MaxBackups int `yaml:"max_backups"`

	// 保存的最大天数
	MaxAge int `yaml:"max_days"`

	// 是否是开发模式
	Development bool

	zap.Config
}

type ModOptions func(options *Options)

var (
	sp                             = string(filepath.Separator)
	errWS, warnWS, infoWS, debugWS zapcore.WriteSyncer       // IO输出
	debugConsoleWS                 = zapcore.Lock(os.Stdout) // 控制台标准输出
	errorConsoleWS                 = zapcore.Lock(os.Stderr)
)

type Logger struct {
	*zap.Logger
	sync.RWMutex
	Opts      *Options `json:"opts"`
	zapConfig zap.Config
	initd     bool
}

func NewLogger(mod ...ModOptions) *zap.Logger {
	l := &Logger{}

	l.Lock()
	defer l.Unlock()
	if l.initd {
		return nil
	}

	l.Opts = &Options{
		ErrorFileName: "error.log",
		WarnFileName:  "warn.log",
		InfoFileName:  "info.log",
		DebugFileName: "debug.log",
		Level:         zapcore.DebugLevel,
		MaxSize:       defaultLoggerMaxSize,
		MaxBackups:    defaultMaxBackups,
		MaxAge:        defaultMaxAge,
	}

	for _, fn := range mod {
		fn(l.Opts)
	}

	// if l.Opts.LogFileDir == "" {
	// 	l.Opts.LogFileDir, _ = filepath.Abs(filepath.Dir(filepath.Join(".")))
	// 	l.Opts.LogFileDir += sp + "logs" + sp
	// }

	l.zapConfig = zap.NewDevelopmentConfig()

	l.zapConfig.EncoderConfig.TimeKey = "ts"
	l.zapConfig.EncoderConfig.LevelKey = "level"
	l.zapConfig.EncoderConfig.MessageKey = "message"
	l.zapConfig.EncoderConfig.CallerKey = "caller"

	l.zapConfig.DisableCaller = true
	l.zapConfig.DisableStacktrace = true

	l.zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")

	if l.Opts.OutputPaths == nil || len(l.Opts.OutputPaths) == 0 {
		l.zapConfig.OutputPaths = []string{"stdout"}
	}
	if l.Opts.ErrorOutputPaths == nil || len(l.Opts.ErrorOutputPaths) == 0 {
		l.zapConfig.OutputPaths = []string{"stderr"}
	}

	l.zapConfig.Level.SetLevel(l.Opts.Level)
	l.init()
	l.initd = true

	return l.Logger
}

func (l *Logger) init() {
	l.setSyncers()
	var err error
	l.Logger, err = l.zapConfig.Build(l.cores())
	if err != nil {
		panic(err)
	}
	defer l.Logger.Sync()
}

func (l *Logger) setSyncers() {
	f := func(fN string) zapcore.WriteSyncer {
		return zapcore.AddSync(&lumberjack.Logger{
			Filename:   l.Opts.LogFileDir + sp + l.Opts.AppName + "-" + fN,
			MaxSize:    l.Opts.MaxSize,
			MaxBackups: l.Opts.MaxBackups,
			MaxAge:     l.Opts.MaxAge,
			Compress:   false,
			LocalTime:  true,
		})
	}
	errWS = f(l.Opts.ErrorFileName)
	warnWS = f(l.Opts.WarnFileName)
	infoWS = f(l.Opts.InfoFileName)
	debugWS = f(l.Opts.DebugFileName)
	return
}

// SetMaxSize
func SetMaxSize(MaxSize int) ModOptions {
	return func(option *Options) {
		option.MaxSize = MaxSize
	}
}

// SetMaxBackups
func SetMaxBackups(MaxBackups int) ModOptions {
	return func(option *Options) {
		option.MaxBackups = MaxBackups
	}
}

// SetMaxAge
func SetMaxAge(MaxAge int) ModOptions {
	return func(option *Options) {
		option.MaxAge = MaxAge
	}
}

// SetLogFileDir
func SetLogFileDir(LogFileDir string) ModOptions {
	return func(option *Options) {
		option.LogFileDir = LogFileDir
	}
}

// SetAppName
func SetAppName(AppName string) ModOptions {
	return func(option *Options) {
		option.AppName = AppName
	}
}

func SetLevel(Level zapcore.Level) ModOptions {
	return func(option *Options) {
		option.Level = Level
	}
}
func SetErrorFileName(ErrorFileName string) ModOptions {
	return func(option *Options) {
		option.ErrorFileName = ErrorFileName
	}
}
func SetWarnFileName(WarnFileName string) ModOptions {
	return func(option *Options) {
		option.WarnFileName = WarnFileName
	}
}

func SetInfoFileName(InfoFileName string) ModOptions {
	return func(option *Options) {
		option.InfoFileName = InfoFileName
	}
}
func SetDebugFileName(DebugFileName string) ModOptions {
	return func(option *Options) {
		option.DebugFileName = DebugFileName
	}
}
func SetDevelopment(Development bool) ModOptions {
	return func(option *Options) {
		option.Development = Development
	}
}
func (l *Logger) cores() zap.Option {
	fileEncoder := zapcore.NewJSONEncoder(l.zapConfig.EncoderConfig)
	//consoleEncoder := zapcore.NewConsoleEncoder(l.zapConfig.EncoderConfig)
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = timeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	errPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.ErrorLevel && zapcore.ErrorLevel-l.zapConfig.Level.Level() > -1
	})
	warnPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.WarnLevel && zapcore.WarnLevel-l.zapConfig.Level.Level() > -1
	})
	infoPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.InfoLevel && zapcore.InfoLevel-l.zapConfig.Level.Level() > -1
	})
	debugPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.DebugLevel && zapcore.DebugLevel-l.zapConfig.Level.Level() > -1
	})
	cores := []zapcore.Core{
		zapcore.NewCore(fileEncoder, errWS, errPriority),
		zapcore.NewCore(fileEncoder, warnWS, warnPriority),
		zapcore.NewCore(fileEncoder, infoWS, infoPriority),
		zapcore.NewCore(fileEncoder, debugWS, debugPriority),
	}
	if l.Opts.Development {
		cores = append(cores, []zapcore.Core{
			zapcore.NewCore(consoleEncoder, errorConsoleWS, errPriority),
			zapcore.NewCore(consoleEncoder, debugConsoleWS, warnPriority),
			zapcore.NewCore(consoleEncoder, debugConsoleWS, infoPriority),
			zapcore.NewCore(consoleEncoder, debugConsoleWS, debugPriority),
		}...)
	}
	return zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(cores...)
	})
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func timeUnixNano(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(t.Unix() / 1e6)
}

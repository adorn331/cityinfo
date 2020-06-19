package logger

import (
	"cityinfo/configs"
	"cityinfo/utils/emailutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"sync"
	"time"
)

var (
	// Log is global logger
	Log *zap.Logger

	// onceInit guarantee initialize logger only once
	onceInit sync.Once
)

// Init initializes log by input parameters
// lvl - global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
// timeFormat - custom time format for logger of empty string to use default
func init() {
	onceInit.Do(func() {
		lvl := configs.LOG_LEVEL
		// First, define our level-handling logic.
		globalLevel := zapcore.Level(lvl)

		// High-priority output should also go to standard error, and low-priority
		// output should also go to standard out.
		// It is usefull for Kubernetes deployment.
		// Kubernetes interprets os.Stdout log items as INFO and os.Stderr log items
		// as ERROR by default.
		highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		})
		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= globalLevel && lvl < zapcore.ErrorLevel
		})

		file, err := os.OpenFile(configs.LOG_FILE, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		localFileErrors := zapcore.Lock(file)
		EmailErrors := zapcore.AddSync(emailutil.ErrNotifier)
		consoleInfos := zapcore.Lock(os.Stdout)
		consoleErrors := zapcore.Lock(os.Stderr)

		// Configure logger output format.
		config := zap.NewProductionEncoderConfig()
		config.LevelKey = "level"
		config.TimeKey = "time"
		config.CallerKey = "caller"
		config.MessageKey = "message"
		config.EncodeTime = func (t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(configs.LOG_TIME_FORMAT))
		}
		JsonEncoder := zapcore.NewJSONEncoder(config)

		// Join the outputs, encoders, and level-handling functions into
		// zapcore.
		core := zapcore.NewTee(
			// for highPriority, need to log to console, file and send email notifier
			zapcore.NewCore(JsonEncoder, consoleErrors, highPriority),
			zapcore.NewCore(JsonEncoder, localFileErrors, highPriority),
			zapcore.NewCore(JsonEncoder, EmailErrors, highPriority), // Error info could send Email!

			zapcore.NewCore(JsonEncoder, consoleInfos, lowPriority),
		)

		// From a zapcore.Core, it's easy to construct a Logger.
		Log = zap.New(core)
		zap.RedirectStdLog(Log)
	})
}
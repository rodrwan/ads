package logger

import (
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Fields es un alias para logrus.Fields para mantener la compatibilidad
type Fields = logrus.Fields

// Entry es un alias para logrus.Entry para mantener la compatibilidad
type Entry = logrus.Entry

var log *logrus.Logger

const (
	ENV_DEVELOPMENT = "development"
	ENV_STAGING     = "staging"
	ENV_PRODUCTION  = "production"
)

type LogConfig struct {
	Level      string
	Format     string
	OutputFile string
}

// ServiceHook es un hook personalizado que agrega el nombre del servicio a cada log
type ServiceHook struct {
	ServiceName string
}

// Fire implementa la interfaz logrus.Hook
func (hook *ServiceHook) Fire(entry *logrus.Entry) error {
	if entry.Data == nil {
		entry.Data = make(logrus.Fields)
	}
	entry.Data["service"] = hook.ServiceName
	return nil
}

// Levels implementa la interfaz logrus.Hook
func (hook *ServiceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Initialize inicializa el logger con un nombre de servicio
func Initialize(serviceName string) {
	environment := getEnvOrDefault("ENVIRONMENT", "development")
	config := getLogConfigForEnvironment(environment)
	setupLogger(config)

	// Agregar el hook con el nombre del servicio
	log.AddHook(&ServiceHook{ServiceName: serviceName})
}

func init() {
	log = logrus.New()
	environment := getEnvOrDefault("ENVIRONMENT", "development")
	config := getLogConfigForEnvironment(environment)
	setupLogger(config)
}

func getEnvOrDefault(env, defaultValue string) string {
	if value := os.Getenv(env); value != "" {
		return value
	}
	return defaultValue
}

func getLogConfigForEnvironment(env string) LogConfig {
	switch strings.ToLower(env) {
	case ENV_PRODUCTION:
		return LogConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),
			Format: "json",
		}
	case ENV_STAGING:
		return LogConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "debug"),
			Format: "json",
		}
	default: // development
		return LogConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "debug"),
			Format: "text",
		}
	}
}

func setupLogger(config LogConfig) {
	// Configurar el formato
	if config.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "severity",
				logrus.FieldKeyMsg:   "message",
			},
			TimestampFormat: time.RFC3339Nano,
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        time.RFC3339Nano,
			ForceColors:            true,
			DisableLevelTruncation: true,
		})
	}

	// Configurar el nivel de log
	if level, err := logrus.ParseLevel(strings.ToLower(config.Level)); err == nil {
		log.SetLevel(level)
	}

	// Configurar output
	log.SetOutput(os.Stdout)
}

// Info logs a message at level Info.
func Info(args ...interface{}) {
	log.Info(args...)
}

// Infof logs a formatted message at level Info.
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Error logs a message at level Error.
func Error(args ...interface{}) {
	log.Error(args...)
}

// Errorf logs a formatted message at level Error.
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Warn logs a message at level Warn.
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Warnf logs a formatted message at level Warn.
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Debug logs a message at level Debug.
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Debugf logs a formatted message at level Debug.
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Fatal logs a message at level Fatal then the process will exit with status set to 1.
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Fatalf logs a formatted message at level Fatal then the process will exit with status set to 1.
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// SignalInfo adds a signal field to the log entry.
func SignalInfo(sig interface{}) *logrus.Entry {
	return log.WithField("signal", sig)
}

// WithFields adds multiple fields to the log entry.
func WithFields(fields Fields) *logrus.Entry {
	return log.WithFields(fields)
}

// WithError adds an error field to the log entry.
func WithError(err error) *logrus.Entry {
	return log.WithError(err)
}

// WithField adds a single field to the log entry.
func WithField(key string, value interface{}) *logrus.Entry {
	return log.WithField(key, value)
}

// GetLogger returns the underlying logrus logger instance.
func GetLogger() *logrus.Logger {
	return log
}

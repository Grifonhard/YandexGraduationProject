package logger

import (
	"fmt"
	"reflect"
)

var log *Logger

func InitLogger(level string, logFilePath string) error {
	lognew, err := new(level, logFilePath)
	if err != nil {
		return err
	}
	log = lognew
	return nil
}

func Info(msg string, a ...any) {
	log.info(fmt.Sprintf(msg, a...))
}

func Debug(msg string, a ...any) {
	if len(a) == 1 {
		arg := a[0]
		typ := reflect.TypeOf(arg)
		if typ.Kind() == reflect.Ptr && typ.Elem().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			err := arg.(*error) 
			if err != nil && *err != nil {
				log.debug(fmt.Sprintf(msg, (*err).Error()))
				return
			}
			return
		}
	}
	log.debug(fmt.Sprintf(msg, a...))
}

func Warn(msg string, a ...any) {
	log.warn(fmt.Sprintf(msg, a...))
}

func Error(msg string, a ...any) {
	log.error(fmt.Sprintf(msg, a...))
}

func Fatal(msg string, a ...any) {
	log.fatal(fmt.Sprintf(msg, a...))
}

func Sync() {
	log.sync()
}

// Package logs provides primitive logger with filtering level
package logs

import (
	"io"
	"log"
	"strings"
)

type Logger struct {
	log   *log.Logger
	level Level
}

func New(out io.Writer, lv Level, prop int) *Logger {
	return &Logger{log: log.New(out, "", prop), level: lv}
}

func (self *Logger) Printf(fmt string, v ...interface{}) {
	self.log.Printf(fmt, v...)
}

func (self *Logger) writef(lv Level, fmt string, v ...interface{}) {
	if lv >= self.level {
		self.log.Printf(fmt, v...)
	}
}

func (self *Logger) Tracef(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[TRACE]", fmt}, " ")
	self.writef(Trace, fmt, v...)
}

func (self *Logger) Debugf(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[DEBUG]", fmt}, " ")
	self.writef(Debug, fmt, v...)
}

func (self *Logger) Infof(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[INFO]", fmt}, " ")
	self.writef(Info, fmt, v...)
}

func (self *Logger) Noticef(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[NOTICE]", fmt}, " ")
	self.writef(Notice, fmt, v...)
}

func (self *Logger) Warnf(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[WARN]", fmt}, " ")
	self.writef(Warning, fmt, v...)
}

func (self *Logger) Errorf(fmt string, v ...interface{}) {
	fmt = strings.Join([]string{"[ERROR]", fmt}, " ")
	self.writef(Error, fmt, v...)
}

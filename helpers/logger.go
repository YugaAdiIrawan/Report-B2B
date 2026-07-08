package helpers

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type LastLog struct {
	mu   sync.RWMutex
	last LastEntry
}

type LastEntry struct {
	Timestamp time.Time
	Level     zerolog.Level
	Message   string
}

func NewLastLog() *LastLog {
	return &LastLog{}
}

func (log *LastLog) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level == zerolog.NoLevel || !e.Enabled() {
		return
	}
	log.mu.Lock()
	log.last = LastEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
	}
	log.mu.Unlock()
}

func (log *LastLog) Last() (LastEntry, bool) {
	log.mu.RLock()
	defer log.mu.RUnlock()
	if log.last.Timestamp.IsZero() {
		return LastEntry{}, false
	}
	return log.last, true
}

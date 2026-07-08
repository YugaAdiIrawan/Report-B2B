package schaduler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	report2 "github.com/YugaAdiIrawan/model/report"
	"github.com/YugaAdiIrawan/module/report"
	"github.com/rs/zerolog/log"
)

type AutoUWScheduler struct {
	usecase  report.ReportUsecase
	config   report2.AutoUWSchedulerConfig
	stopChan chan struct{}
	doneChan chan struct{}
	stopOnce sync.Once
	paused   atomic.Bool
}

func NewAutoUWSchaduers(uc report.ReportUsecase, cfg report2.AutoUWSchedulerConfig) (*AutoUWScheduler, error) {
	_, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, fmt.Errorf("scheduler: load timezone: %s %w", cfg.Timezone, err)
	}

	return &AutoUWScheduler{
		usecase:  uc,
		config:   cfg,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}, nil
}

func (s *AutoUWScheduler) Start() {
	go s.loop()
}

func (s *AutoUWScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
		<-s.doneChan
	})
}

func (s *AutoUWScheduler) Pause() {
	s.paused.Store(true)
	log.Info().Msg("auto_uw_scheduler: paused")
}

func (s *AutoUWScheduler) Resume() {
	s.paused.Store(false)
	log.Info().Msg("auto_uw_scheduler: resumed")
}

func (s *AutoUWScheduler) IsPaused() bool {
	return s.paused.Load()
}

func (s *AutoUWScheduler) loop() {
	defer close(s.doneChan)

	for {
		next := s.nextRunTime()
		waitDuration := time.Until(next)
		timer := time.NewTimer(waitDuration)

		select {
		case <-s.stopChan:

			timer.Stop()
			return

		case <-timer.C:
			s.runDailyReport()
		}
	}
}

func (s *AutoUWScheduler) nextRunTime() time.Time {
	loc, _ := time.LoadLocation(s.config.Timezone)
	now := time.Now().In(loc)

	next := time.Date(now.Year(), now.Month(), now.Day(), s.config.RunHour, s.config.RunMinute, 0, 0, loc)

	if !next.After(now) {
		next = time.Date(now.Year(), now.Month(), now.Day()+1, s.config.RunHour, s.config.RunMinute, 0, 0, loc)
	}

	//interval := time.Duration(s.config.IntervalMin) * time.Minute
	//next := now.Add(interval)

	log.Debug().
		Str("now", now.Format("2006-01-02 15:04:05 MST")).
		Str("next_run", next.Format("2006-01-02 15:04:05 MST")).
		Dur("wait_duration", time.Until(next)).
		Msgf("auto_uw_scheduler: next run scheduled %s", next)
	return next
}

func (s *AutoUWScheduler) runDailyReport() {
	if s.paused.Load() {
		log.Info().Msg("auto_uw_scheduler: skipped, scheduler is paused")
		return
	}

	loc, _ := time.LoadLocation(s.config.Timezone)
	yesterday := time.Now().In(loc).AddDate(0, 0, -1)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := s.usecase.SendDailyReport(ctx, yesterday, s.config.Recipients); err != nil {
		log.Error().
			Err(err).
			Str("target_date", yesterday.Format("2006-01-02")).
			Str("timezone", s.config.Timezone).
			Msg("auto_uw_scheduler: failed to send daily report")
		return
	}

	log.Info().
		Str("target_date", yesterday.Format("2006-01-02")).
		Msg("auto_uw_scheduler: daily report sent successfully")

}

package dblayer

import (
	"sync"
	"time"
)

// MetricsCollector собирает метрики выполнения запросов
type MetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string]*QueryMetrics
}

type QueryMetrics struct {
	Count        int64
	TotalTime    time.Duration
	AverageTime  time.Duration
	MaxTime      time.Duration
	ErrorCount   int64
	LastExecuted time.Time
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*QueryMetrics),
	}
}

func (mc *MetricsCollector) Track(query string, duration time.Duration, err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	m, exists := mc.metrics[query]
	if !exists {
		m = &QueryMetrics{}
		mc.metrics[query] = m
	}

	m.Count++
	m.TotalTime += duration
	m.AverageTime = m.TotalTime / time.Duration(m.Count)
	if duration > m.MaxTime {
		m.MaxTime = duration
	}
	if err != nil {
		m.ErrorCount++
	}
	m.LastExecuted = time.Now()
}

// WithMetrics добавляет сбор метрик
func (qb *QueryBuilder) WithMetrics(collector *MetricsCollector) *QueryBuilder {
	qb.On(BeforeCreate, func(data interface{}) error {
		start := time.Now()
		collector.Track("CREATE", time.Since(start), nil)
		return nil
	})

	qb.On(BeforeUpdate, func(data interface{}) error {
		start := time.Now()
		collector.Track("UPDATE", time.Since(start), nil)
		return nil
	})

	return qb
}

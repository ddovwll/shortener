package services

import (
	"context"
	"encoding/json"
	"shortener/src/internal/application/contracts"
	"shortener/src/internal/domain/visit"
)

type VisitService struct {
	visitRepository visit.VisitRepository
	producer        contracts.MessageProducer
}

func NewVisitService(visitRepository visit.VisitRepository, producer contracts.MessageProducer) *VisitService {
	return &VisitService{
		visitRepository: visitRepository,
		producer:        producer,
	}
}

func (s *VisitService) Register(ctx context.Context, visit visit.Visit) error {
	bytes, err := json.Marshal(visit)
	if err != nil {
		return err
	}

	return s.producer.Produce(ctx, []byte{}, bytes)
}

func (s *VisitService) CreateBatch(ctx context.Context, visits []visit.Visit) {
	s.visitRepository.CreateBatch(ctx, visits)
}

func (s *VisitService) ByDayAnalytics(ctx context.Context, shortURL string) ([]visit.PeriodCount, error) {
	return s.visitRepository.AnalyticsAggregatedByDay(ctx, shortURL)
}

func (s *VisitService) ByMonthAnalytics(ctx context.Context, shortURL string) ([]visit.PeriodCount, error) {
	return s.visitRepository.AnalyticsAggregatedByMonth(ctx, shortURL)
}

func (s *VisitService) ByUserAgentAnalytics(ctx context.Context, shortURL string) ([]visit.UserAgentCount, error) {
	return s.visitRepository.AnalyticsAggregatedByUserAgent(ctx, shortURL)
}

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

	return s.producer.Produce(ctx, bytes)
}

func (s *VisitService) CreateBatch(ctx context.Context, visits []*visit.Visit) error {
	return s.visitRepository.CreateBatch(ctx, visits)
}

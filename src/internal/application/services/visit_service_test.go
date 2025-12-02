package services_test

import (
	"context"
	"encoding/json"
	"testing"

	"shortener/src/internal/application/services"
	"shortener/src/internal/application/services/mocks"
	"shortener/src/internal/domain/visit"

	"go.uber.org/mock/gomock"
)

func TestVisitService_Register_CallsProducer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockVisitRepository(ctrl)
	mockProducer := mocks.NewMockMessageProducer(ctrl)

	svc := services.NewVisitService(mockRepo, mockProducer)

	v := visit.Visit{}

	var captured []byte
	mockProducer.
		EXPECT().
		Produce(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key, value []byte) error {
			captured = make([]byte, len(value))
			copy(captured, value)
			return nil
		})

	if err := svc.Register(context.Background(), v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded visit.Visit
	if err := json.Unmarshal(captured, &decoded); err != nil {
		t.Fatalf("produced value is not valid JSON of visit: %v", err)
	}
}

func TestVisitService_CreateBatch_CallsRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockVisitRepository(ctrl)
	mockProducer := mocks.NewMockMessageProducer(ctrl)

	svc := services.NewVisitService(mockRepo, mockProducer)

	visits := []visit.Visit{{}, {}}

	mockRepo.EXPECT().CreateBatch(gomock.Any(), gomock.Eq(visits)).Times(1)

	svc.CreateBatch(context.Background(), visits)
}

func TestVisitService_AnalyticsDelegation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockVisitRepository(ctrl)
	mockProducer := mocks.NewMockMessageProducer(ctrl)

	byDay := []visit.PeriodCount{{}}
	byMonth := []visit.PeriodCount{{}}
	byUA := []visit.UserAgentCount{{}}

	mockRepo.EXPECT().AnalyticsAggregatedByDay(gomock.Any(), gomock.Eq("k")).Return(byDay, nil)
	mockRepo.EXPECT().AnalyticsAggregatedByMonth(gomock.Any(), gomock.Eq("k")).Return(byMonth, nil)
	mockRepo.EXPECT().AnalyticsAggregatedByUserAgent(gomock.Any(), gomock.Eq("k")).Return(byUA, nil)

	svc := services.NewVisitService(mockRepo, mockProducer)

	if _, err := svc.ByDayAnalytics(context.Background(), "k"); err != nil {
		t.Fatalf("ByDayAnalytics error: %v", err)
	}
	if _, err := svc.ByMonthAnalytics(context.Background(), "k"); err != nil {
		t.Fatalf("ByMonthAnalytics error: %v", err)
	}
	if _, err := svc.ByUserAgentAnalytics(context.Background(), "k"); err != nil {
		t.Fatalf("ByUserAgentAnalytics error: %v", err)
	}
}

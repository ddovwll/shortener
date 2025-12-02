package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"shortener/src/internal/application/services/mocks"
	"testing"

	"shortener/src/internal/application/services"
	shortlink "shortener/src/internal/domain/short_link"

	"go.uber.org/mock/gomock"
)

type seqGenerator struct {
	vals []string
	i    int
	err  error
}

func (g *seqGenerator) Generate() (string, error) {
	if g.err != nil {
		return "", g.err
	}
	if g.i >= len(g.vals) {
		return g.vals[len(g.vals)-1], nil
	}
	v := g.vals[g.i]
	g.i++
	return v, nil
}

func TestShortLinkService_Create_SuccessGenerated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockShortLinkRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)

	gen := &seqGenerator{vals: []string{"gen1"}}

	mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Eq("gen1"), gomock.Eq("https://example.com")).
		Return(&shortlink.ShortLink{}, nil)

	svc := services.NewShortLinkService(mockRepo, gen, mockCache)

	ctx := context.Background()
	link, err := svc.Create(ctx, "", "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link == nil {
		t.Fatalf("expected link, got nil")
	}
}

func TestShortLinkService_Create_CustomConflict_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockShortLinkRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)

	gen := &seqGenerator{vals: []string{"unused"}}

	mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Eq("custom"), gomock.Eq("https://ex")).
		Return(nil, shortlink.ErrShortLinkAlreadyExists)

	svc := services.NewShortLinkService(mockRepo, gen, mockCache)
	_, err := svc.Create(context.Background(), "custom", "https://ex")
	if !errors.Is(err, shortlink.ErrShortLinkAlreadyExists) {
		t.Fatalf("expected ErrShortLinkAlreadyExists, got: %v", err)
	}
}

func TestShortLinkService_Create_RetryThenSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockShortLinkRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)

	gen := &seqGenerator{vals: []string{"a", "b"}}

	first := mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Eq("a"), gomock.Eq("o")).
		Return(nil, shortlink.ErrShortLinkAlreadyExists)
	second := mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Eq("b"), gomock.Eq("o")).
		Return(&shortlink.ShortLink{}, nil)
	gomock.InOrder(first, second)

	svc := services.NewShortLinkService(mockRepo, gen, mockCache)
	link, err := svc.Create(context.Background(), "", "o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link == nil {
		t.Fatalf("expected non-nil link")
	}
}

func TestShortLinkService_Create_FailedAfterAttempts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockShortLinkRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)

	gen := &seqGenerator{vals: []string{"x"}}
	mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, shortlink.ErrShortLinkAlreadyExists).
		AnyTimes()

	svc := services.NewShortLinkService(mockRepo, gen, mockCache)
	_, err := svc.Create(context.Background(), "", "orig")
	if err == nil {
		t.Fatalf("expected error after attempts, got nil")
	}
}

func TestShortLinkService_Get_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockShortLinkRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)

	gen := &seqGenerator{vals: []string{"unused"}}

	model := &shortlink.ShortLink{}
	//nolint: errcheck // empty model
	bytes, _ := json.Marshal(model)

	mockCache.EXPECT().Get(gomock.Any(), gomock.Eq("k")).Return(string(bytes), nil)

	svc := services.NewShortLinkService(mockRepo, gen, mockCache)
	got, err := svc.Get(context.Background(), "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected model from cache, got nil")
	}
}

func TestShortLinkService_Get_CacheMiss_ReadsDBAndSetsCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockShortLinkRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)

	gen := &seqGenerator{vals: []string{"unused"}}

	mockCache.EXPECT().Get(gomock.Any(), gomock.Eq("k")).Return("", errors.New("not found"))

	mockRepo.EXPECT().Get(gomock.Any(), gomock.Eq("k")).Return(&shortlink.ShortLink{}, nil)

	mockCache.EXPECT().Set(gomock.Any(), gomock.Eq("k"), gomock.Any(), gomock.Any()).Return(nil)

	svc := services.NewShortLinkService(mockRepo, gen, mockCache)
	got, err := svc.Get(context.Background(), "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected model from db, got nil")
	}
}

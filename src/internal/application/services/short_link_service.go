package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"shortener/src/internal/application/contracts"
	shortlink "shortener/src/internal/domain/short_link"
	"shortener/src/pkg/logger"
	"time"
)

type ShortLinkService struct {
	shortLinkRepository shortlink.ShortLinkRepository
	generator           shortlink.ShortLinkGenerator
	cache               contracts.Cache
}

const maxCreateAttempts = 5
const cachedURLTTL = time.Hour * 5

func NewShortLinkService(
	shortLinkRepo shortlink.ShortLinkRepository,
	generator shortlink.ShortLinkGenerator,
	cache contracts.Cache,
) *ShortLinkService {
	return &ShortLinkService{
		shortLinkRepository: shortLinkRepo,
		generator:           generator,
		cache:               cache,
	}
}

func (s *ShortLinkService) Create(
	ctx context.Context,
	shortURL, originalURL string,
) (*shortlink.ShortLink, error) {
	customURL := shortURL != ""
	for attempt := 0; attempt < maxCreateAttempts; attempt++ {
		if shortURL == "" {
			generated, err := s.generator.Generate()
			if err != nil {
				return nil, err
			}
			shortURL = generated
		}

		link, err := s.shortLinkRepository.Create(ctx, shortURL, originalURL)
		if err == nil {
			return link, nil
		}

		if errors.Is(err, shortlink.ErrShortLinkAlreadyExists) && customURL {
			return nil, err
		}

		shortURL = ""
	}

	return nil, fmt.Errorf(
		"failed to create short link after %d attempts: %w",
		maxCreateAttempts, shortlink.ErrShortLinkAlreadyExists,
	)
}

func (s *ShortLinkService) Get(ctx context.Context, shortURL string) (*shortlink.ShortLink, error) {
	value, err := s.cache.Get(ctx, shortURL)
	if err != nil {
		dbValue, err := s.shortLinkRepository.Get(ctx, shortURL)
		if err != nil {
			return nil, err
		}

		bytes, err := json.Marshal(dbValue)
		if err != nil {
			logger.Error("failed to marshal short link", "err", err)
		} else {
			if err := s.cache.Set(ctx, shortURL, bytes, cachedURLTTL); err != nil {
				logger.Error("failed to set URL into cache", "err", err)
			}
		}

		return dbValue, nil
	}

	var model shortlink.ShortLink
	if err := json.Unmarshal([]byte(value), &model); err != nil {
		return nil, err
	}

	return &model, nil
}

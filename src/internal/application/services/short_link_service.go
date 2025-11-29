package services

import (
	"context"
	"errors"
	"fmt"
	shortlink "shortener/src/internal/domain/short_link"
)

type ShortLinkService struct {
	shortLinkRepository shortlink.ShortLinkRepository
	generator           shortlink.ShortLinkGenerator
}

const maxCreateAttempts = 5

func NewShortLinkService(
	shortLinkRepo shortlink.ShortLinkRepository,
	generator shortlink.ShortLinkGenerator,
) *ShortLinkService {
	return &ShortLinkService{
		shortLinkRepository: shortLinkRepo,
		generator:           generator,
	}
}

func (s *ShortLinkService) Create(
	ctx context.Context,
	shortURL, originalURL string,
) (*shortlink.ShortLink, error) {
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

		if !errors.Is(err, shortlink.ErrShortLinkAlreadyExists) {
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
	return s.shortLinkRepository.Get(ctx, shortURL)
}

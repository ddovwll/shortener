package services

import (
	"context"
	"errors"
	"fmt"
	"shortener/src/internal/domain/short_link"
)

type ShortLinkService struct {
	shortLinkRepository shortlink.ShortLinkRepository
	generator           shortlink.ShortLinkGenerator
}

const maxCreateAttempts = 5

func NewShortLinkService(shortLinkRepo shortlink.ShortLinkRepository, generator shortlink.ShortLinkGenerator) ShortLinkService {
	return ShortLinkService{
		shortLinkRepository: shortLinkRepo,
		generator:           generator,
	}
}

func (s *ShortLinkService) Create(
	ctx context.Context,
	shortUrl, originalUrl string,
) (*shortlink.ShortLink, error) {
	for attempt := 0; attempt < maxCreateAttempts; attempt++ {
		if shortUrl == "" {
			generated, err := s.generator.Generate()
			if err != nil {
				return nil, err
			}
			shortUrl = generated
		}

		link, err := s.shortLinkRepository.Create(ctx, shortUrl, originalUrl)
		if err == nil {
			return link, nil
		}

		if !errors.Is(err, shortlink.ErrShortURLAlreadyExists) {
			return nil, err
		}

		shortUrl = ""
	}

	return nil, fmt.Errorf(
		"failed to create short link after %d attempts: %w",
		maxCreateAttempts, shortlink.ErrShortURLAlreadyExists,
	)
}

func (s *ShortLinkService) Get(ctx context.Context, shortUrl string) (*shortlink.ShortLink, error) {
	return s.shortLinkRepository.Get(ctx, shortUrl)
}

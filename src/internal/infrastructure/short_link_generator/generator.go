package generator

import (
	"crypto/rand"
	"fmt"
	shortlink "shortener/src/internal/domain/short_link"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

var (
	alphabetLen  = byte(len(alphabet))
	maxRandValue = byte(255 - (256 % len(alphabet)))
)

type RandomShortLinkGenerator struct{}

func NewRandomShortCodeGenerator() *RandomShortLinkGenerator {
	return &RandomShortLinkGenerator{}
}

func (g *RandomShortLinkGenerator) Generate() (string, error) {
	b := make([]byte, shortlink.ShortLinkLength)

	for i := 0; i < shortlink.ShortLinkLength; {
		var rb [1]byte

		if _, err := rand.Read(rb[:]); err != nil {
			return "", fmt.Errorf("read random: %w", err)
		}

		if rb[0] > maxRandValue {
			continue
		}

		b[i] = alphabet[rb[0]%alphabetLen]
		i++
	}

	code := string(b)

	return code, nil
}

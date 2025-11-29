package shortlink

type ShortLinkGenerator interface {
	Generate() (string, error)
}

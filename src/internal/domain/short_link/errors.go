package shortlink

import "errors"

var ErrShortLinkAlreadyExists = errors.New("short link already exists")
var ErrShortLinkNotFound = errors.New("short link not found")

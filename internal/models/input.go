package models

import (
	"errors"
	"fmt"
	"net/url"
)

type Input []string

var (
	ErrTooManyLinks = errors.New("too many links in a request")
	ErrNoLinks      = errors.New("no links provided")
)

func (i *Input) Validate(maxLinksPerIn uint32) error {
	linksCount := len(*i)
	if linksCount < 1 {
		return ErrNoLinks
	}

	if uint32(linksCount) > maxLinksPerIn {
		return ErrTooManyLinks
	}

	for _, link := range *i {
		_, err := url.ParseRequestURI(link)
		if err != nil {
			return fmt.Errorf("failed to parse a link: %w", err)
		}
	}

	return nil
}

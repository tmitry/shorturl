package models

import netUrl "net/url"

type URL string

func (u URL) IsValid() bool {
	_, err := netUrl.ParseRequestURI(u.String())

	return err == nil
}

func (u URL) String() string {
	return string(u)
}

package DataBase

import "errors"

var (
	ErrUserExists          = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrAppNotFound         = errors.New("app not found")
	ErrCatalogItemNotFound = errors.New("catalog item not found")
	ErrNoStyles            = errors.New("catalog item must contain at least one style")
	ErrNoColors            = errors.New("catalog item must contain at least one color")
	ErrOutfitNotFound      = errors.New("outfit item not found")
)

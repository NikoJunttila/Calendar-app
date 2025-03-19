package handlers

import (
	"gothstack/app/types"
	"gothstack/app/views/errors"

	"github.com/anthdm/superkit/kit"
)

func HandleAuthentication(kit *kit.Kit) (kit.Auth, error) {
	return types.AuthUser{}, nil
}
func HandleUnauthorized(kit *kit.Kit) error {
	return kit.Render(errors.Unauthorized())
}

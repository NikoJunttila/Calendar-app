package handlers

import (
	"gothstack/app/views/landing"

	"github.com/anthdm/superkit/kit"
)

func HandleLandingIndex(kit *kit.Kit) error {
	return kit.Render(landing.Index())
}

func HandleLandingTest(title, description string) func(*kit.Kit) error {
	return func(kit *kit.Kit) error {
		return kit.Render(landing.Test(
			title,
			description,
		))
	}
}

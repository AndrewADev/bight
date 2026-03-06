package strategy

import (
	"fmt"

	"github.com/AndrewADev/bight/internal/config"
)

type Context struct {
	Branch  string
	Project string
}

func Apply(s string, ctx Context, cfg *config.Config) (string, error) {
	switch s {
	case "template":
		return applyTemplate(ctx, cfg.Defaults.BranchTemplate)
	case "random":
		return applyRandom()
	default:
		return "", fmt.Errorf("unknown strategy: %s", s)
	}
}

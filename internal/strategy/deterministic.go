package strategy

import (
	"crypto/sha256"
	"encoding/hex"
)

func applyDeterministic(ctx Context) (string, error) {
	input := ctx.Project + "_" + ctx.Branch
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:]), nil
}

package authn_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/internal/storage/authn/memory"
)

func TestBootstrap(t *testing.T) {
	store := memory.NewStore()
	s, err := authn.Bootstrap(context.TODO(), store, authn.WithExpiration(time.Minute), authn.WithToken("this-is-a-token"))
	assert.NoError(t, err)
	assert.NotEmpty(t, s)
}

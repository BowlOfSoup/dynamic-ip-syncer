package transip

import (
	"github.com/transip/gotransip/v6/jwt"
	"sync"
)

type TokenCache struct {
	tokens map[string]jwt.Token
	mutex  sync.Mutex
}

func NewTokenCache() *TokenCache {
	return &TokenCache{
		tokens: make(map[string]jwt.Token),
	}
}

// Get retrieves the token from memory for the given key.
func (c *TokenCache) Get(key string) (jwt.Token, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	token, exists := c.tokens[key]
	if !exists {
		return jwt.Token{}, nil
	}
	return token, nil
}

// Set stores the token in memory for the given key.
func (c *TokenCache) Set(key string, token jwt.Token) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.tokens[key] = token
	return nil
}

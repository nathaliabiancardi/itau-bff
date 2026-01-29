package clients

import (
	"context"
	"time"
)

type CustomerProfile struct {
	ID         string
	Name       string
	Preference string
}

type ProfileClient struct {
}

func NewProfileClient() *ProfileClient {
	return &ProfileClient{}
}

func (c *ProfileClient) GetProfile(ctx context.Context, customerID string) (*CustomerProfile, error) {
	select {
	case <-time.After(150 * time.Millisecond):
		return &CustomerProfile{
			ID:         customerID,
			Name:       "Cliente Exemplo",
			Preference: "digital",
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

package access

import "context"

// Service is the single entitlement boundary for published trainer content.
// Phase 19 allows authenticated users; subscriptions can replace this policy.
type Service interface {
	CanAccessContent(context.Context, string, *string) (bool, error)
}
type PublishedContent struct{}

func (PublishedContent) CanAccessContent(_ context.Context, user string, _ *string) (bool, error) {
	return user != "", nil
}

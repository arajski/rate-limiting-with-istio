package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/arajski/custom-rate-limiting-cloudcon-2025/ratelimiting-service/internal/domain"
	"github.com/arajski/custom-rate-limiting-cloudcon-2025/ratelimiting-service/pkg/repository"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	envoy "github.com/envoyproxy/go-control-plane/ratelimit/config/ratelimit/v3"
	"github.com/google/uuid"
)

type RateLimit struct {
	RateLimitsRepository repository.RateLimits
	UsersRepository      repository.Users
	Cache                cache.SnapshotCache
	DB                   *sql.DB
}

func (s *RateLimit) RefreshCache() error {
	rateLimits, err := s.RateLimitsRepository.GetAllRateLimits()
	if err != nil {
		return fmt.Errorf("could not retrieve rate limits:%w", err)
	}
	envoyConfigs, err := s.buildEnvoyConfigs("demo", "ratelimit", rateLimits)
	if err != nil {
		return fmt.Errorf("could not build envoy config: %v", err)
	}
	err = s.updateCacheWithRateLimitConfigs(envoyConfigs)
	if err != nil {
		return fmt.Errorf("could not update the cache: %w", err)
	}
	return nil
}

func (s *RateLimit) buildEnvoyConfigs(name string, domain string, rateLimits []*domain.RateLimit) ([]*envoy.RateLimitConfig, error) {
	allDescriptors := make([]*envoy.RateLimitDescriptor, 0, len(rateLimits))
	authTokenDescriptorsByPath := make(map[string][]*envoy.RateLimitDescriptor)

	for _, rateLimit := range rateLimits {
		unit, err := toEnvoyInterval(rateLimit.LimitInterval)
		if err != nil {
			return nil, fmt.Errorf("could not parse rate limit interval %v", rateLimit.LimitInterval)
		}
		apiToken, err := s.UsersRepository.GetAPITokenByID(rateLimit.UserID)
		if err != nil {
			fmt.Printf("could not get an api token for userId %d", rateLimit.UserID)
			continue
		}

		descriptor := &envoy.RateLimitDescriptor{
			Key:   "auth5",
			Value: apiToken[len(apiToken)-5:],
			RateLimit: &envoy.RateLimitPolicy{
				Unit:            unit,
				RequestsPerUnit: uint32(rateLimit.LimitUnit),
			},
		}

		if rateLimit.Endpoint == "" {
			allDescriptors = append(allDescriptors, descriptor)
		} else {
			authTokenDescriptorsByPath[rateLimit.Endpoint] = append(authTokenDescriptorsByPath[rateLimit.Endpoint], descriptor)
		}
	}

	for path, descriptors := range authTokenDescriptorsByPath {
		descriptor := &envoy.RateLimitDescriptor{
			Key:         "path",
			Value:       path,
			Descriptors: descriptors,
		}
		allDescriptors = append(allDescriptors, descriptor)
	}

	rateLimitConfigs := []*envoy.RateLimitConfig{
		{
			Name:        name,
			Domain:      domain,
			Descriptors: allDescriptors,
		},
	}

	return rateLimitConfigs, nil
}

func (s *RateLimit) updateCacheWithRateLimitConfigs(envoyConfigs []*envoy.RateLimitConfig) error {
	rateLimitResources := make([]types.Resource, 0, len(envoyConfigs))
	for _, config := range envoyConfigs {
		rateLimitResources = append(rateLimitResources, &envoy.RateLimitConfig{
			Name:        config.Name,
			Domain:      config.Domain,
			Descriptors: config.Descriptors,
		})
	}

	snapshot, _ := cache.NewSnapshot(uuid.NewString(),
		map[resource.Type][]types.Resource{
			resource.RateLimitConfigType: rateLimitResources,
		},
	)

	if err := snapshot.Consistent(); err != nil {
		return fmt.Errorf("snapshot is inconsistent: %+v\n%+v", snapshot, err)
	}

	if err := s.Cache.SetSnapshot(context.Background(), "test-node", snapshot); err != nil {
		return fmt.Errorf("snapshot error %q for %+v", err, snapshot)
	}
	return nil
}

func toEnvoyInterval(u string) (envoy.RateLimitUnit, error) {
	switch u {
	case "second":
		return envoy.RateLimitUnit_SECOND, nil
	case "minute":
		return envoy.RateLimitUnit_MINUTE, nil
	case "hour":
		return envoy.RateLimitUnit_HOUR, nil
	case "day":
		return envoy.RateLimitUnit_DAY, nil
	default:
		return envoy.RateLimitUnit_UNKNOWN, fmt.Errorf("unit %v not supported", u)
	}
}

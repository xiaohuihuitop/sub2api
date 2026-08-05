package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type platformRepositoryStub struct {
	allRules  []PlatformModelRule
	platforms []Platform
	created   *Platform
}

func (r *platformRepositoryStub) Create(_ context.Context, platform *Platform) error {
	r.created = platform
	return nil
}

func (r *platformRepositoryStub) ListModelRules(context.Context) ([]PlatformModelRule, error) {
	return append([]PlatformModelRule(nil), r.allRules...), nil
}

func (r *platformRepositoryStub) List(context.Context) ([]Platform, error) {
	return append([]Platform(nil), r.platforms...), nil
}

type platformManagementRepositoryStub struct {
	platformRepositoryStub
	platform *Platform
	updated  *Platform
}

func (r *platformManagementRepositoryStub) GetByID(context.Context, int64) (*Platform, error) {
	return r.platform, nil
}

func (r *platformManagementRepositoryStub) Update(_ context.Context, platform *Platform) error {
	r.updated = platform
	return nil
}

func TestPlatformServiceCreateRejectsCrossPlatformModelOverlap(t *testing.T) {
	repo := &platformRepositoryStub{
		allRules: []PlatformModelRule{{
			PlatformID:   1,
			PlatformCode: PlatformOpenAI,
			ModelPattern: "gpt-*",
			Enabled:      true,
		}},
	}
	svc := NewPlatformService(repo)

	_, err := svc.Create(context.Background(), CreatePlatformInput{
		Code:            "grok",
		Name:            "Grok",
		AccountPlatform: PlatformOpenAI,
		ModelRules: []PlatformModelRule{{
			ModelPattern: "gpt-4o",
			Enabled:      true,
		}},
	})

	require.ErrorContains(t, err, "overlaps")
	require.Nil(t, repo.created)
}

func TestPlatformServiceResolveModelUsesActiveRepositoryRules(t *testing.T) {
	repo := &platformRepositoryStub{
		allRules: []PlatformModelRule{{
			ID:            22,
			PlatformID:    2,
			PlatformCode:  PlatformOpenAI,
			ModelPattern:  "gpt-4o",
			UpstreamModel: "gpt-4o-2024-08-06",
			Enabled:       true,
		}},
	}
	svc := NewPlatformService(repo)

	resolved, err := svc.ResolveModel(context.Background(), "gpt-4o")

	require.NoError(t, err)
	require.Equal(t, int64(2), resolved.PlatformID)
	require.Equal(t, PlatformOpenAI, resolved.PlatformCode)
	require.Equal(t, "gpt-4o-2024-08-06", resolved.UpstreamModel)
}

func TestPlatformServiceUpdateReplacesOwnRuleWithoutSelfConflict(t *testing.T) {
	repo := &platformManagementRepositoryStub{
		platformRepositoryStub: platformRepositoryStub{
			allRules: []PlatformModelRule{{
				PlatformID:   1,
				PlatformCode: PlatformOpenAI,
				ModelPattern: "gpt-*",
				Enabled:      true,
			}},
		},
		platform: &Platform{
			ID:     1,
			Code:   PlatformOpenAI,
			Name:   "OpenAI",
			Status: PlatformStatusActive,
			ModelRules: []PlatformModelRule{{
				PlatformID:   1,
				PlatformCode: PlatformOpenAI,
				ModelPattern: "gpt-*",
				Enabled:      true,
			}},
		},
	}
	svc := NewPlatformService(repo)
	rules := []PlatformModelRule{{ModelPattern: "gpt-*", Enabled: true}}

	updated, err := svc.Update(context.Background(), 1, UpdatePlatformInput{ModelRules: &rules})

	require.NoError(t, err)
	require.Equal(t, int64(1), updated.ID)
	require.NotNil(t, repo.updated)
	require.Equal(t, int64(1), repo.updated.ModelRules[0].PlatformID)
	require.Equal(t, PlatformOpenAI, repo.updated.ModelRules[0].PlatformCode)
}

func TestPlatformServiceListReturnsIndependentPlatformCopies(t *testing.T) {
	repo := &platformRepositoryStub{platforms: []Platform{{
		ID:              7,
		Code:            "gpt",
		Name:            "GPT",
		AccountPlatform: PlatformOpenAI,
		Status:          PlatformStatusActive,
		ModelRules: []PlatformModelRule{{
			ID:           11,
			ModelPattern: "gpt-4o",
			Enabled:      true,
		}},
	}}}

	platforms, err := NewPlatformService(repo).List(context.Background())

	require.NoError(t, err)
	require.Len(t, platforms, 1)
	platforms[0].Name = "mutated"
	platforms[0].ModelRules[0].ModelPattern = "mutated"
	require.Equal(t, "GPT", repo.platforms[0].Name)
	require.Equal(t, "gpt-4o", repo.platforms[0].ModelRules[0].ModelPattern)
}

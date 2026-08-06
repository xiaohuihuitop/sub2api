package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildAccountForCreatePreservesPlatformPool(t *testing.T) {
	platformID := int64(42)
	account, err := buildAccountForCreate(&CreateAccountInput{
		Name:       "gpt-account",
		Platform:   PlatformOpenAI,
		PlatformID: &platformID,
		Type:       AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "test-key",
		},
		Concurrency: 1,
	}, map[string]any{})

	require.NoError(t, err)
	require.Equal(t, &platformID, account.PlatformID)
}

type platformBindingAccountRepoStub struct {
	*upstreamBillingProbeAccountRepo
	binding PlatformAccountBinding
}

func (r *platformBindingAccountRepoStub) ResolvePlatformAccountBinding(context.Context, int64) (PlatformAccountBinding, error) {
	return r.binding, nil
}

func (r *platformBindingAccountRepoStub) ValidatePlatformAccountBinding(_ context.Context, _ int64, accountPlatform string) error {
	if accountPlatform != r.binding.AccountPlatform {
		return ErrPlatformInvalid
	}
	return nil
}

func TestCreateAccountDerivesAdapterFromSelectedPlatform(t *testing.T) {
	repo := &platformBindingAccountRepoStub{
		upstreamBillingProbeAccountRepo: &upstreamBillingProbeAccountRepo{},
		binding:                         PlatformAccountBinding{ID: 42, Code: "openai-main", AccountPlatform: PlatformOpenAI, Status: StatusActive},
	}

	created, err := (&adminServiceImpl{accountRepo: repo}).CreateAccount(context.Background(), &CreateAccountInput{
		Name:                 "platform-account",
		Platform:             PlatformGemini,
		PlatformID:           int64Pointer(42),
		Type:                 AccountTypeAPIKey,
		Credentials:          map[string]any{"api_key": "test"},
		SkipDefaultGroupBind: true,
	})

	require.NoError(t, err)
	require.Equal(t, PlatformOpenAI, created.Platform)
	require.Equal(t, int64(42), *created.PlatformID)
}

func TestValidatePlatformAccountAdapterChangeRejectsCrossAdapterMove(t *testing.T) {
	err := validatePlatformAccountAdapterChange(PlatformOpenAI, PlatformAnthropic)

	require.ErrorIs(t, err, ErrPlatformInvalid)
}

func TestValidatePlatformAccountAdapterChangeAllowsSameAdapterPoolMove(t *testing.T) {
	require.NoError(t, validatePlatformAccountAdapterChange(PlatformOpenAI, " OPENAI "))
}

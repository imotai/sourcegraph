package providers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

type MockAuthProvider struct {
	MockConfigID            ConfigID
	MockConfig              schema.AuthProviderCommon
	MockAuthProvidersConfig *schema.AuthProviders
	MockPublicAccountData   *extsvc.PublicAccountData
	MockType                ProviderType
}

func (m MockAuthProvider) ConfigID() ConfigID {
	return m.MockConfigID
}

func (m MockAuthProvider) Type() ProviderType {
	return m.MockType
}

func (m MockAuthProvider) Config() schema.AuthProviders {
	if m.MockAuthProvidersConfig != nil {
		return *m.MockAuthProvidersConfig
	}

	return schema.AuthProviders{
		Github: &schema.GitHubAuthProvider{
			Type:          m.MockConfigID.Type,
			DisplayName:   m.MockConfig.DisplayName,
			DisplayPrefix: m.MockConfig.DisplayPrefix,
			Hidden:        m.MockConfig.Hidden,
			Order:         m.MockConfig.Order,
		},
	}
}

func (m MockAuthProvider) CachedInfo() *Info {
	return &Info{
		DisplayName: m.MockConfigID.Type,
	}
}

func (m MockAuthProvider) ExternalAccountInfo(_ context.Context, _ extsvc.Account) (*extsvc.PublicAccountData, error) {
	return m.MockPublicAccountData, nil
}

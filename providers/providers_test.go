package providers

import "testing"

func TestAllProvidersImplementInterface(t *testing.T) {
	_ = ProviderRegistry{
		"nzbsorg":      NewNzbsOrg(""),
		"nyaatorrents": NewNyaaTorrents(),
	}
}

package issuetracker

type ClientFactory interface {
	NewClient(provider ProviderType, credentials map[string]string) (Client, error)
}

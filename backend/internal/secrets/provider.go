package secrets

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type Provider interface {
	GetSecret(name string) (string, error)
	ListSecrets() ([]string, error)
}

type EnvSecretProvider struct{}

func NewEnvSecretProvider() *EnvSecretProvider {
	return &EnvSecretProvider{}
}

func (p *EnvSecretProvider) GetSecret(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("secret name is required")
	}
	value, ok := os.LookupEnv(name)
	if !ok {
		return "", fmt.Errorf("secret not found: %s", name)
	}
	return value, nil
}

func (p *EnvSecretProvider) ListSecrets() ([]string, error) {
	env := os.Environ()
	names := make([]string, 0, len(env))
	for _, entry := range env {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
			continue
		}
		names = append(names, parts[0])
	}
	sort.Strings(names)
	return names, nil
}

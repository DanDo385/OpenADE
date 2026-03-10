package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"openade/internal/model"
)

type ProviderService struct {
	DB *sql.DB
}

func NewProviderService(database *sql.DB) *ProviderService {
	return &ProviderService{DB: database}
}

// providerSettings is the JSON structure stored in the config column.
type providerSettings struct {
	APIKey       string `json:"api_key"`
	BaseURL      string `json:"base_url,omitempty"`
	DefaultModel string `json:"default_model,omitempty"`
}

func (s *ProviderService) List(ctx context.Context) ([]model.ProviderConfig, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, provider, config FROM provider_configs`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing providers: %w", err)
	}
	defer rows.Close()

	var configs []model.ProviderConfig
	for rows.Next() {
		var id, provider, configJSON string
		if err := rows.Scan(&id, &provider, &configJSON); err != nil {
			return nil, fmt.Errorf("scanning provider: %w", err)
		}
		cfg := parseProviderConfig(id, provider, configJSON)
		configs = append(configs, cfg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if !hasProvider(configs, "openai") {
		if cfg := envProviderConfig("openai"); cfg != nil {
			configs = append(configs, *cfg)
		}
	}

	return configs, nil
}

func (s *ProviderService) Get(ctx context.Context, provider string) (*model.ProviderConfig, error) {
	var id, configJSON string
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, config FROM provider_configs WHERE provider = ?`, provider,
	).Scan(&id, &configJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting provider: %w", err)
	}
	cfg := parseProviderConfig(id, provider, configJSON)
	if !cfg.Configured {
		if envCfg := envProviderConfig(provider); envCfg != nil {
			return envCfg, nil
		}
	}
	return &cfg, nil
}

func (s *ProviderService) Save(ctx context.Context, provider string, req model.SaveProviderRequest) (*model.ProviderConfig, error) {
	settings := providerSettings{
		APIKey:       req.APIKey,
		BaseURL:      req.BaseURL,
		DefaultModel: req.DefaultModel,
	}
	configJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("marshaling provider config: %w", err)
	}

	// Upsert: try update first, then insert
	res, err := s.DB.ExecContext(ctx,
		`UPDATE provider_configs SET config = ? WHERE provider = ?`,
		string(configJSON), provider,
	)
	if err != nil {
		return nil, fmt.Errorf("updating provider: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		id := uuid.NewString()
		_, err = s.DB.ExecContext(ctx,
			`INSERT INTO provider_configs (id, provider, config) VALUES (?, ?, ?)`,
			id, provider, string(configJSON),
		)
		if err != nil {
			return nil, fmt.Errorf("inserting provider: %w", err)
		}
	}

	return s.Get(ctx, provider)
}

// GetDefault returns the first configured provider or nil.
func (s *ProviderService) GetDefault(ctx context.Context) (*model.ProviderConfig, error) {
	configs, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range configs {
		if c.Configured {
			return &c, nil
		}
	}
	return nil, nil
}

func parseProviderConfig(id, provider, configJSON string) model.ProviderConfig {
	var settings providerSettings
	json.Unmarshal([]byte(configJSON), &settings)
	return model.ProviderConfig{
		ID:           id,
		Provider:     provider,
		APIKey:       settings.APIKey,
		BaseURL:      settings.BaseURL,
		DefaultModel: settings.DefaultModel,
		Configured:   settings.APIKey != "",
	}
}

func hasProvider(configs []model.ProviderConfig, provider string) bool {
	for _, cfg := range configs {
		if cfg.Provider == provider {
			return true
		}
	}
	return false
}

func envProviderConfig(provider string) *model.ProviderConfig {
	switch provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil
		}
		defaultModel := os.Getenv("OPENAI_DEFAULT_MODEL")
		if defaultModel == "" {
			defaultModel = "gpt-4o-mini"
		}
		return &model.ProviderConfig{
			ID:           "env-openai",
			Provider:     "openai",
			APIKey:       apiKey,
			BaseURL:      os.Getenv("OPENAI_BASE_URL"),
			DefaultModel: defaultModel,
			Configured:   true,
		}
	default:
		return nil
	}
}

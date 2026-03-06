package bootstrap

import (
	"context"
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/internal/config"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

type providerCreator interface {
	Create(ctx context.Context, provider *marketdata.Provider) error
}

type providerBootstrapConfig struct {
	name    marketdata.ProviderName
	baseURI string
	apiKeys []string
}

func Providers(ctx context.Context, pc providerCreator, cfg config.Providers, logger logging.Logger) {
	if pc == nil {
		panic(fmt.Errorf("bootstrap providers: provider creator is required"))
	}

	configs := []providerBootstrapConfig{
		{
			name:    marketdata.ProviderMarketStack,
			baseURI: cfg.MarketStack.BaseURI,
			apiKeys: cfg.MarketStack.APIKeys,
		},
		{
			name:    marketdata.ProviderAlphaVantage,
			baseURI: cfg.AlphaVantage.BaseURI,
			apiKeys: cfg.AlphaVantage.APIKeys,
		},
	}

	manualProviders := []marketdata.ProviderName{
		marketdata.ProviderBrandNewDay,
	}

	for _, cfg := range configs {
		if len(cfg.apiKeys) == 0 {
			logger.Info(ctx, "provider bootstrap skipped: no api keys configured", "provider", string(cfg.name))
			continue
		}

		for _, apiKey := range cfg.apiKeys {
			provider, err := marketdata.NewAPIProviderWithAPIKey(cfg.name, cfg.baseURI, apiKey)
			if err != nil {
				panic(fmt.Errorf("bootstrap providers: build provider %s: %w", cfg.name, err))
			}
			if err := pc.Create(ctx, provider); err != nil {
				panic(fmt.Errorf("bootstrap providers: create provider %s: %w", cfg.name, err))
			}
		}

		logger.Info(ctx, "bootstrapped provider api keys", "provider", string(cfg.name), "keys_count", len(cfg.apiKeys))
	}

	for _, providerName := range manualProviders {
		provider, err := marketdata.NewManualProvider(providerName)
		if err != nil {
			panic(fmt.Errorf("bootstrap providers: build manual provider %s: %w", providerName, err))
		}
		if err := pc.Create(ctx, provider); err != nil {
			panic(fmt.Errorf("bootstrap providers: create manual provider %s: %w", providerName, err))
		}
		logger.Info(ctx, "bootstrapped manual provider", "provider", string(providerName))
	}
}

package gcloud

import (
	"context"

	_ "embed"
	"errors"

	"github.com/influxdata/telegraf"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/common/slog"
	"github.com/influxdata/telegraf/plugins/secretstores"

	"cloud.google.com/go/auth"
	creds "cloud.google.com/go/auth/credentials"
)

//go:embed sample.conf
var sampleConfig string

func (*GCloud) SampleConfig() string {
	return sampleConfig
}

// GCloud is the main authenticator struct
type GCloud struct {
	Audience           string          `toml:"sts_audience"`
	Log                telegraf.Logger `toml:"-"`
	ServiceAccountFile string          `toml:"service_account_file"`
	common_http.HTTPClientConfig

	creds *auth.Credentials
}

func (g *GCloud) Init() error {
	if g.ServiceAccountFile == "" {
		return errors.New("service_account_file is required")
	}

	if g.Audience == "" {
		return errors.New("sts_audience is required")
	}

	httpClient, err := g.HTTPClientConfig.CreateClient(context.Background(), g.Log)
	if err != nil {
		return err
	}

	creds, err := creds.DetectDefault(&creds.DetectOptions{
		STSAudience:     g.Audience,
		CredentialsFile: g.ServiceAccountFile,
		Client:          httpClient,
		Logger:          slog.NewLogger(g.Log),
	})
	g.creds = creds

	if err != nil {
		return err
	}

	return nil
}

// Get retrieves the token. The key is ignored as this secret store only provides one secret.
func (g *GCloud) Get(_ string) ([]byte, error) {
	token, err := g.creds.Token(context.Background())
	if err != nil {
		return nil, err
	}
	return []byte(token.Value), nil
}

// List returns the list of secrets provided by this store.
func (*GCloud) List() ([]string, error) {
	return []string{"token"}, nil
}

// Set is not supported for the gcloud secret store.
func (*GCloud) Set(_, _ string) error {
	return errors.New("setting secrets is not supported")
}

// GetResolver returns a resolver function for the secret.
func (g *GCloud) GetResolver(key string) (telegraf.ResolveFunc, error) {
	return func() ([]byte, bool, error) {
		s, err := g.Get(key)
		return s, true, err
	}, nil
}

func init() {
	secretstores.Add("gcloud", func(_ string) telegraf.SecretStore {
		return &GCloud{}
	})
}

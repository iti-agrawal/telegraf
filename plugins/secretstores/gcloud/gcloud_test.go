package gcloud

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cloud.google.com/go/auth"
	"github.com/influxdata/telegraf/testutil"
)

func TestSampleConfig(t *testing.T) {
	plugin := &GCloud{}
	require.NotEmpty(t, plugin.SampleConfig())
}

func TestInit(t *testing.T) {
	tests := []struct {
		name        string
		plugin      *GCloud
		wantErr     bool
		errContains string
	}{
		{
			name: "missing service account file should fail",
			plugin: &GCloud{
				Audience: "https://localhost",
				Log:      testutil.Logger{},
			},
			wantErr:     true,
			errContains: "service_account_file is required",
		},
		{
			name: "non-existent service account file should fail",
			plugin: &GCloud{
				Audience:           "https://localhost",
				Log:                testutil.Logger{},
				ServiceAccountFile: "non-existent-file.json",
			},
			wantErr:     true,
			errContains: "no such file or directory",
		},
		{
			name: "invalid service account file json should fail",
			plugin: &GCloud{
				Audience:           "https://localhost",
				Log:                testutil.Logger{},
				ServiceAccountFile: "./testdata/invalid-json-sa-key.json",
			},
			wantErr:     true,
			errContains: "invalid character",
		},
		{
			name: "missing service account type should fail",
			plugin: &GCloud{
				Audience:           "https://localhost",
				Log:                testutil.Logger{},
				ServiceAccountFile: "./testdata/missing-type-sa-key.json",
			},
			wantErr:     true,
			errContains: "unsupported unidentified file type",
		},
		{
			name: "missing audience should fail",
			plugin: &GCloud{
				Log:                testutil.Logger{},
				ServiceAccountFile: "./testdata/valid-sa-key.json",
			},
			wantErr:     true,
			errContains: "audience is required",
		},
		{
			name: "successful init",
			plugin: &GCloud{
				Audience:           "https://localhost",
				Log:                testutil.Logger{},
				ServiceAccountFile: "./testdata/valid-sa-key.json",
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.plugin.Init()
			if tc.wantErr {
				require.ErrorContains(t, err, tc.errContains, "error mismatch")
			} else {
				require.NoError(t, err)
				require.NotNil(t, tc.plugin.Audience)
				require.NotNil(t, tc.plugin.creds)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		provider    auth.TokenProvider
		wantToken   []byte
		wantErr     bool
		errContains string
	}{
		{
			name: "successful get",
			provider: mockTokenProvider{
				token: &auth.Token{Value: "token", Expiry: time.Now().Add(time.Hour)},
			},
			wantToken: []byte("token"),
			wantErr:   false,
		},
		{
			name: "error getting token",
			provider: mockTokenProvider{
				err: errors.New("token provider error"),
			},
			wantErr:     true,
			errContains: "token provider error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &GCloud{
				creds: auth.NewCredentials(&auth.CredentialsOptions{
					TokenProvider: tt.provider,
				}),
			}

			token, err := plugin.Get("any")
			if tt.wantErr {
				require.ErrorContains(t, err, tt.errContains)
				require.Nil(t, token)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantToken, token)
			}
		})
	}
}

type mockTokenProvider struct {
	token *auth.Token
	err   error
}

func (tp mockTokenProvider) Token(context.Context) (*auth.Token, error) {
	if tp.err != nil {
		return nil, tp.err
	}
	return tp.token, nil
}

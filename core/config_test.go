// Copyright 2021 The sacloud Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"bytes"
	"io"
	"testing"

	"github.com/goccy/go-yaml"

	"github.com/sacloud/autoscaler/config"
	"github.com/stretchr/testify/require"
)

func TestConfig_Load(t *testing.T) {
	type fields struct {
		SakuraCloud *SakuraCloud
		Handlers    Handlers
		Resources   ResourceDefinitions
		AutoScaler  AutoScalerConfig
	}
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "minimal",
			fields: fields{
				SakuraCloud: &SakuraCloud{
					Credential: Credential{
						Token:  "token",
						Secret: "secret",
					},
				},
				Handlers: Handlers{
					{
						Name:     "fake",
						Endpoint: "unix:autoscaler-handlers-fake.sock",
					},
				},
				Resources: ResourceDefinitions{
					&ResourceDefServer{
						ResourceDefBase: &ResourceDefBase{
							DefName:  "test-name",
							TypeName: "Server",
						},
						Selector: &MultiZoneSelector{
							ResourceSelector: &ResourceSelector{
								Names: []string{"test-name"},
							},
							Zones: []string{"is1a"},
						},
						DedicatedCPU: true,
					},
				},
				AutoScaler: AutoScalerConfig{
					CoolDownSec: 30,
					ServerTLSConfig: &config.TLSStruct{
						TLSCertPath: "server.crt",
						TLSKeyPath:  "server.key",
						ClientAuth:  "RequireAndVerifyClientCert",
						ClientCAs:   "ca.crt",
					},
					HandlerTLSConfig: &config.TLSStruct{
						TLSCertPath: "server.crt",
						TLSKeyPath:  "server.key",
						RootCAs:     "ca.crt",
					},
					ExporterConfig: &config.ExporterConfig{
						Enabled: true,
						Address: "localhost:8080",
						TLSConfig: &config.TLSStruct{
							TLSCertPath: "server.crt",
							TLSKeyPath:  "server.key",
							ClientAuth:  "RequireAndVerifyClientCert",
							ClientCAs:   "ca.crt",
						},
					},
				},
			},
			args: args{
				reader: bytes.NewReader([]byte(`
sakuracloud:
  token: token
  secret: secret
handlers:
  - name: "fake"
    endpoint: "unix:autoscaler-handlers-fake.sock"
resources:
  - type: Server
    name: "test-name"
    selector:
      names: ["test-name"]
      zones: ["is1a"]
    dedicated_cpu: true
autoscaler:
  cooldown: 30
  server_tls_config:
    cert_file: server.crt
    key_file: server.key
    client_auth_type: RequireAndVerifyClientCert
    client_ca_file: ca.crt
  handler_tls_config:
    cert_file: server.crt
    key_file: server.key
    root_ca_file: ca.crt
  exporter_config:
    enabled: true
    address: "localhost:8080"
    tls_config:
      cert_file: server.crt
      key_file: server.key
      client_auth_type: RequireAndVerifyClientCert
      client_ca_file: ca.crt

`)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := &Config{
				SakuraCloud:    tt.fields.SakuraCloud,
				CustomHandlers: tt.fields.Handlers,
				Resources:      tt.fields.Resources,
				AutoScaler:     tt.fields.AutoScaler,
			}
			c := &Config{}
			if err := c.load(tt.args.reader); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}

			require.EqualValues(t, expected, c)
		})
	}
}

func TestHandlersConfig_UnmarshalYAML(t *testing.T) {
	data := []byte(`
disabled: true
handlers:
  foo:
    disabled: true
  dns-servers-handler:
    disabled: true
`)

	var config HandlersConfig
	if err := yaml.UnmarshalWithOptions(data, &config); err != nil {
		t.Fatal(err)
	}
	expected := HandlersConfig{
		Disabled: true,
		Handlers: map[string]*HandlerConfig{
			"foo":                 {Disabled: true},
			"dns-servers-handler": {Disabled: true},
		},
	}
	require.EqualValues(t, expected, config)
}

func TestConfig_Handlers(t *testing.T) {
	type fields struct {
		AutoScaler AutoScalerConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   Handlers
	}{
		{
			name:   "empty",
			fields: fields{},
			want:   BuiltinHandlers(),
		},
		{
			name: "disable all",
			fields: fields{
				AutoScaler: AutoScalerConfig{
					HandlersConfig: &HandlersConfig{
						Disabled: true,
					},
				},
			},
			want: nil,
		},
		{
			name: "disable per handler",
			fields: fields{
				AutoScaler: AutoScalerConfig{
					HandlersConfig: &HandlersConfig{
						Handlers: map[string]*HandlerConfig{
							"dns-servers-handler":           {Disabled: false},
							"elb-vertical-scaler":           {Disabled: false},
							"elb-servers-handler":           {Disabled: true},
							"gslb-servers-handler":          {Disabled: true},
							"load-balancer-servers-handler": {Disabled: true},
							"router-vertical-scaler":        {Disabled: true},
							"server-horizontal-scaler":      {Disabled: true},
							"server-vertical-scaler":        {Disabled: true},
						},
					},
				},
			},
			want: []*Handler{
				{Name: "dns-servers-handler"},
				{Name: "elb-vertical-scaler"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				SakuraCloud: &SakuraCloud{Credential: Credential{}},
				AutoScaler:  tt.fields.AutoScaler,
			}
			got := c.Handlers()

			var gotNames, wantNames []string
			for _, h := range got {
				gotNames = append(gotNames, h.Name)
			}
			for _, h := range tt.want {
				wantNames = append(wantNames, h.Name)
			}
			require.EqualValues(t, gotNames, wantNames)
		})
	}
}

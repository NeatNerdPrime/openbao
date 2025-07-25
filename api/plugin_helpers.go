// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"regexp"

	"github.com/go-jose/go-jose/v3/jwt"
)

const (
	// PluginAutoMTLSEnv is used to ensure AutoMTLS is used. This will override
	// setting a TLSProviderFunc for a plugin.
	PluginAutoMTLSEnv = "BAO_PLUGIN_AUTOMTLS_ENABLED"

	// PluginMetadataModeEnv is an ENV name used to disable TLS communication
	// to bootstrap mounting plugins.
	PluginMetadataModeEnv = "BAO_PLUGIN_METADATA_MODE"

	// PluginUnwrapTokenEnv is the ENV name used to pass unwrap tokens to the
	// plugin.
	PluginUnwrapTokenEnv = "BAO_UNWRAP_TOKEN"
)

// sudoPaths is a map containing the paths that require a token's policy
// to have the "sudo" capability. The keys are the paths as strings, in
// the same format as they are returned by the OpenAPI spec. The values
// are the regular expressions that can be used to test whether a given
// path matches that path or not (useful specifically for the paths that
// contain templated fields.)
var sudoPaths = map[string]*regexp.Regexp{
	"/auth/token/accessors":                         regexp.MustCompile(`^/auth/token/accessors/?$`),
	"/pki/root":                                     regexp.MustCompile(`^/pki/root$`),
	"/pki/root/sign-self-issued":                    regexp.MustCompile(`^/pki/root/sign-self-issued$`),
	"/sys/audit":                                    regexp.MustCompile(`^/sys/audit$`),
	"/sys/audit/{path}":                             regexp.MustCompile(`^/sys/audit/.+$`),
	"/sys/auth/{path}":                              regexp.MustCompile(`^/sys/auth/.+$`),
	"/sys/auth/{path}/tune":                         regexp.MustCompile(`^/sys/auth/.+/tune$`),
	"/sys/config/auditing/request-headers":          regexp.MustCompile(`^/sys/config/auditing/request-headers$`),
	"/sys/config/auditing/request-headers/{header}": regexp.MustCompile(`^/sys/config/auditing/request-headers/.+$`),
	"/sys/config/cors":                              regexp.MustCompile(`^/sys/config/cors$`),
	"/sys/config/ui/headers":                        regexp.MustCompile(`^/sys/config/ui/headers/?$`),
	"/sys/config/ui/headers/{header}":               regexp.MustCompile(`^/sys/config/ui/headers/.+$`),
	"/sys/leases":                                   regexp.MustCompile(`^/sys/leases$`),
	"/sys/leases/lookup/":                           regexp.MustCompile(`^/sys/leases/lookup/?$`),
	"/sys/leases/lookup/{prefix}":                   regexp.MustCompile(`^/sys/leases/lookup/.+$`),
	"/sys/leases/revoke-force/{prefix}":             regexp.MustCompile(`^/sys/leases/revoke-force/.+$`),
	"/sys/leases/revoke-prefix/{prefix}":            regexp.MustCompile(`^/sys/leases/revoke-prefix/.+$`),
	"/sys/plugins/catalog/{name}":                   regexp.MustCompile(`^/sys/plugins/catalog/[^/]+$`),
	"/sys/plugins/catalog/{type}":                   regexp.MustCompile(`^/sys/plugins/catalog/[\w-]+$`),
	"/sys/plugins/catalog/{type}/{name}":            regexp.MustCompile(`^/sys/plugins/catalog/[\w-]+/[^/]+$`),
	"/sys/raw":                                      regexp.MustCompile(`^/sys/raw$`),
	"/sys/raw/{path}":                               regexp.MustCompile(`^/sys/raw/.+$`),
	"/sys/remount":                                  regexp.MustCompile(`^/sys/remount$`),
	"/sys/revoke-force/{prefix}":                    regexp.MustCompile(`^/sys/revoke-force/.+$`),
	"/sys/revoke-prefix/{prefix}":                   regexp.MustCompile(`^/sys/revoke-prefix/.+$`),
	"/sys/rotate":                                   regexp.MustCompile(`^/sys/rotate$`),
	// equivalent of '/sys/rotate'
	"/sys/rotate/keyring":                regexp.MustCompile(`^/sys/rotate/keyring$`),
	"/sys/rotate/root":                   regexp.MustCompile(`^/sys/rotate/root$`),
	"/sys/internal/inspect/request":      regexp.MustCompile(`^/sys/internal/inspect/request$`),
	"/sys/internal/inspect/router/{tag}": regexp.MustCompile(`^/sys/internal/inspect/router/.+$`),
}

// PluginAPIClientMeta is a helper that plugins can use to configure TLS connections
// back to Vault.
type PluginAPIClientMeta struct {
	// These are set by the command line flags.
	flagCACert     string
	flagCAPath     string
	flagClientCert string
	flagClientKey  string
	flagServerName string
	flagInsecure   bool
}

// FlagSet returns the flag set for configuring the TLS connection
func (f *PluginAPIClientMeta) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("vault plugin settings", flag.ContinueOnError)

	fs.StringVar(&f.flagCACert, "ca-cert", "", "")
	fs.StringVar(&f.flagCAPath, "ca-path", "", "")
	fs.StringVar(&f.flagClientCert, "client-cert", "", "")
	fs.StringVar(&f.flagClientKey, "client-key", "", "")
	fs.StringVar(&f.flagServerName, "tls-server-name", "", "")
	fs.BoolVar(&f.flagInsecure, "tls-skip-verify", false, "")

	return fs
}

// GetTLSConfig will return a TLSConfig based off the values from the flags
func (f *PluginAPIClientMeta) GetTLSConfig() *TLSConfig {
	// If we need custom TLS configuration, then set it
	if f.flagCACert != "" || f.flagCAPath != "" || f.flagClientCert != "" || f.flagClientKey != "" || f.flagInsecure || f.flagServerName != "" {
		t := &TLSConfig{
			CACert:        f.flagCACert,
			CAPath:        f.flagCAPath,
			ClientCert:    f.flagClientCert,
			ClientKey:     f.flagClientKey,
			TLSServerName: f.flagServerName,
			Insecure:      f.flagInsecure,
		}

		return t
	}

	return nil
}

// VaultPluginTLSProvider wraps VaultPluginTLSProviderContext using context.Background.
func VaultPluginTLSProvider(apiTLSConfig *TLSConfig) func() (*tls.Config, error) {
	return VaultPluginTLSProviderContext(context.Background(), apiTLSConfig)
}

// VaultPluginTLSProviderContext is run inside a plugin and retrieves the response
// wrapped TLS certificate from vault. It returns a configured TLS Config.
func VaultPluginTLSProviderContext(ctx context.Context, apiTLSConfig *TLSConfig) func() (*tls.Config, error) {
	if ReadBaoVariable(PluginAutoMTLSEnv) == "true" || ReadBaoVariable(PluginMetadataModeEnv) == "true" {
		return nil
	}

	return func() (*tls.Config, error) {
		unwrapToken := ReadBaoVariable(PluginUnwrapTokenEnv)

		parsedJWT, err := jwt.ParseSigned(unwrapToken)
		if err != nil {
			return nil, fmt.Errorf("error parsing wrapping token: %w", err)
		}

		allClaims := make(map[string]interface{})
		if err = parsedJWT.UnsafeClaimsWithoutVerification(&allClaims); err != nil {
			return nil, fmt.Errorf("error parsing claims from wrapping token: %w", err)
		}

		addrClaimRaw, ok := allClaims["addr"]
		if !ok {
			return nil, errors.New("could not validate addr claim")
		}
		vaultAddr, ok := addrClaimRaw.(string)
		if !ok {
			return nil, errors.New("could not parse addr claim")
		}
		if vaultAddr == "" {
			return nil, errors.New(`no vault api_addr found`)
		}

		// Sanity check the value
		if _, err := url.Parse(vaultAddr); err != nil {
			return nil, fmt.Errorf("error parsing the vault api_addr: %w", err)
		}

		// Unwrap the token
		clientConf := DefaultConfig()
		clientConf.Address = vaultAddr
		if apiTLSConfig != nil {
			err := clientConf.ConfigureTLS(apiTLSConfig)
			if err != nil {
				return nil, fmt.Errorf("error configuring api client %w", err)
			}
		}
		client, err := NewClient(clientConf)
		if err != nil {
			return nil, fmt.Errorf("error during api client creation: %w", err)
		}

		// Reset token value to make sure nothing has been set by default
		client.ClearToken()

		secret, err := client.Logical().UnwrapWithContext(ctx, unwrapToken)
		if err != nil {
			return nil, fmt.Errorf("error during token unwrap request: %w", err)
		}
		if secret == nil {
			return nil, errors.New("error during token unwrap request: secret is nil")
		}

		// Retrieve and parse the server's certificate
		serverCertBytesRaw, ok := secret.Data["ServerCert"].(string)
		if !ok {
			return nil, errors.New("error unmarshalling certificate")
		}

		serverCertBytes, err := base64.StdEncoding.DecodeString(serverCertBytesRaw)
		if err != nil {
			return nil, fmt.Errorf("error parsing certificate: %w", err)
		}

		serverCert, err := x509.ParseCertificate(serverCertBytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing certificate: %w", err)
		}

		// Retrieve and parse the server's private key
		serverKeyB64, ok := secret.Data["ServerKey"].(string)
		if !ok {
			return nil, errors.New("error unmarshalling certificate")
		}

		serverKeyRaw, err := base64.StdEncoding.DecodeString(serverKeyB64)
		if err != nil {
			return nil, fmt.Errorf("error parsing certificate: %w", err)
		}

		serverKey, err := x509.ParseECPrivateKey(serverKeyRaw)
		if err != nil {
			return nil, fmt.Errorf("error parsing certificate: %w", err)
		}

		// Add CA cert to the cert pool
		caCertPool := x509.NewCertPool()
		caCertPool.AddCert(serverCert)

		// Build a certificate object out of the server's cert and private key.
		cert := tls.Certificate{
			Certificate: [][]byte{serverCertBytes},
			PrivateKey:  serverKey,
			Leaf:        serverCert,
		}

		// Setup TLS config
		tlsConfig := &tls.Config{
			ClientCAs:  caCertPool,
			RootCAs:    caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
			// TLS 1.2 minimum
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
			ServerName:   serverCert.Subject.CommonName,
		}

		return tlsConfig, nil
	}
}

func SudoPaths() map[string]*regexp.Regexp {
	return sudoPaths
}

// Determine whether the given path requires the sudo capability.
// Note that this uses hardcoded static path information, so will return incorrect results for paths in namespaces,
// or for secret engines mounted at non-default paths.
func IsSudoPath(path string) bool {
	// Return early if the path is any of the non-templated sudo paths.
	if _, ok := sudoPaths[path]; ok {
		return true
	}

	// Some sudo paths have templated fields in them.
	// (e.g. /sys/revoke-prefix/{prefix})
	// The values in the sudoPaths map are actually regular expressions,
	// so we can check if our path matches against them.
	for _, sudoPathRegexp := range sudoPaths {
		match := sudoPathRegexp.MatchString(path)
		if match {
			return true
		}
	}

	return false
}

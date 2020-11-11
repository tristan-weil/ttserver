package tcplistener

import (
	"os"

	"github.com/caddyserver/certmagic"
	"github.com/libdns/cloudflare"
	"github.com/libdns/digitalocean"
	"github.com/libdns/gandi"
	"github.com/libdns/hetzner"
	ttutils "github.com/tristan-weil/ttserver/utils"
)

func (t *TCPListener) configureACME() error {
	t.logger.
		Info("using a Let's Encrypt certificate")

	t.logger.
		Debug("creating/checking the Let's Encrypt certificate...")

	leConfigOpts := certmagic.Config{
		MustStaple: true,
	}

	if t.tlsConfig.ACME.Storage != nil {
		if t.tlsConfig.ACME.Storage.File != nil {
			if ttutils.NotStringEmpty(t.tlsConfig.ACME.Storage.File.Path) {
				leConfigOpts.Storage = &certmagic.FileStorage{Path: ttutils.StringValue(t.tlsConfig.ACME.Storage.File.Path)}
			}
		}
	}

	leCache := certmagic.NewCache(certmagic.CacheOptions{
		GetConfigForCert: func(cert certmagic.Certificate) (*certmagic.Config, error) {
			return &leConfigOpts, nil
		},
	})

	leConfig := certmagic.New(leCache, leConfigOpts)

	var dnsprovider interface{}
	a := ttutils.StringValue(t.tlsConfig.ACME.DNSProvider)
	switch a {
	case "gandi":
		dnsprovider = &gandi.Provider{
			APIToken: os.Getenv("GANDI_API_TOKEN"),
		}
	case "cloudflare":
		dnsprovider = &cloudflare.Provider{
			APIToken: os.Getenv("CLOUDFLARE_API_TOKEN"),
		}
	case "hetzner":
		dnsprovider = &hetzner.Provider{
			AuthAPIToken: os.Getenv("HETZNER_API_TOKEN"),
		}
	case "digitalocean":
		dnsprovider = &digitalocean.Provider{
			APIToken: os.Getenv("DO_API_TOKEN"),
		}
	}

	acmeManager := certmagic.NewACMEManager(leConfig, certmagic.ACMEManager{
		CA:     ttutils.StringValue(t.tlsConfig.ACME.CA),
		TestCA: ttutils.StringValue(t.tlsConfig.ACME.TestCA),
		Email:  ttutils.StringValue(t.tlsConfig.ACME.Email),
		Agreed: true,
		DNS01Solver: &certmagic.DNS01Solver{
			DNSProvider: dnsprovider.(certmagic.ACMEDNSProvider),
		},
	})

	leConfig.Issuer = acmeManager

	if err := leConfig.ManageSync(t.domains); err != nil {
		return err
	}

	t.logger.
		Debug("creating/checking the Let's Encrypt certificate... done!")

	// TODO: replace
	t.listener.tlsConfig = leConfig.TLSConfig()

	return nil
}

package deckgen

import (
	"crypto/sha256"
	"fmt"

	gojson "github.com/goccy/go-json"
	"github.com/google/go-cmp/cmp"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	"github.com/kong/go-kong/kong/custom"
)

// GenerateSHA generates a SHA256 checksum of targetContent, with the purpose
// of change detection.
func GenerateSHA(targetContent *file.Content, customEntities map[string][]custom.Object) ([]byte, error) {
	jsonConfig, err := gojson.Marshal(targetContent)
	if err != nil {
		return nil, fmt.Errorf("marshaling Kong declarative configuration to JSON: %w", err)
	}
	// Calculate SHA including the custom entities.
	if len(customEntities) > 0 {
		jsonCustomEntities, err := gojson.Marshal(customEntities)
		if err != nil {
			return nil, fmt.Errorf("marshaling Kong custom entities to JSON: %w", err)
		}
		jsonConfig = append(jsonConfig, jsonCustomEntities...)
	}

	shaSum := sha256.Sum256(jsonConfig)
	return shaSum[:], nil
}

// GetFCertificateFromKongCert converts a kong.Certificate to a file.FCertificate.
func GetFCertificateFromKongCert(kongCert kong.Certificate) file.FCertificate {
	var res file.FCertificate
	if kongCert.ID != nil {
		res.ID = kong.String(*kongCert.ID)
	}
	if kongCert.Key != nil {
		res.Key = kong.String(*kongCert.Key)
	}
	if kongCert.Cert != nil {
		res.Cert = kong.String(*kongCert.Cert)
	}
	res.SNIs = getCertsSNIs(kongCert)
	return res
}

func getCertsSNIs(kongCert kong.Certificate) []kong.SNI {
	snis := make([]kong.SNI, 0, len(kongCert.SNIs))
	for _, sni := range kongCert.SNIs {
		kongSNI := kong.SNI{
			Name: sni,
		}
		if kongCert.ID != nil {
			kongSNI.Certificate = &kong.Certificate{
				ID: kongCert.ID,
			}
		}
		snis = append(snis, kongSNI)
	}
	return snis
}

// PluginString returns a string representation of a FPlugin suitable as a sorting key.
//
// Deprecated. To be replaced by a predicate that compares two FPlugins.
func PluginString(plugin file.FPlugin) string {
	result := ""
	if plugin.Name != nil {
		result = *plugin.Name
	}
	if plugin.Consumer != nil && plugin.Consumer.ID != nil {
		result += *plugin.Consumer.ID
	}
	if plugin.ConsumerGroup != nil && plugin.ConsumerGroup.ID != nil {
		result += *plugin.ConsumerGroup.ID
	}
	if plugin.Route != nil && plugin.Route.ID != nil {
		result += *plugin.Route.ID
	}
	if plugin.Service != nil && plugin.Service.ID != nil {
		result += *plugin.Service.ID
	}
	return result
}

// IsContentEmpty returns true if the content is considered empty.
// This ignores meta fields like FormatVersion and Info.
func IsContentEmpty(content *file.Content) bool {
	return cmp.Equal(content, &file.Content{},
		cmp.FilterPath(
			func(p cmp.Path) bool {
				path := p.String()
				return path == "FormatVersion" || path == "Info"
			},
			cmp.Ignore(),
		),
	)
}

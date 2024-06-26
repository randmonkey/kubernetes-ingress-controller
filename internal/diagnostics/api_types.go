package diagnostics

import "github.com/kong/go-database-reconciler/pkg/file"

// ConfigDumpResponse is the GET /debug/config/[successful|failed] response schema.
type ConfigDumpResponse struct {
	ConfigHash string       `json:"hash"`
	Config     file.Content `json:"config"`
}

// FallbackResponse is the GET /debug/config/fallback response schema.
type FallbackResponse struct {
	// Status is the fallback configuration generation status.
	Status FallbackStatus `json:"status"`
	// BrokenObjects is the list of objects that are broken.
	BrokenObjects []FallbackAffectedObjectMeta `json:"brokenObjects,omitempty"`
	// ExcludedObjects is the list of objects that were excluded from the fallback configuration.
	ExcludedObjects []FallbackAffectedObjectMeta `json:"excludedObjects,omitempty"`
	// BackfilledObjects is the list of objects that were backfilled from the last valid cache state.
	BackfilledObjects []FallbackAffectedObjectMeta `json:"backfilledObjects,omitempty"`
}

// FallbackStatus describes whether the fallback configuration generation was triggered or not.
// Making this a string type not a bool to allow for potential future expansion of the status.
type FallbackStatus string

const (
	// FallbackStatusTriggered indicates that the fallback configuration generation was triggered.
	FallbackStatusTriggered FallbackStatus = "triggered"

	// FallbackStatusNotTriggered indicates that the fallback configuration generation was not triggered.
	FallbackStatusNotTriggered FallbackStatus = "not-triggered"
)

// FallbackAffectedObjectMeta is a fallback affected object metadata.
type FallbackAffectedObjectMeta struct {
	// Group is the resource group.
	Group string `json:"group"`
	// Kind is the resource kind.
	Kind string `json:"kind"`
	// Version is the resource version.
	Version string `json:"version,omitempty"`
	// Namespace is the object namespace.
	Namespace string `json:"namespace"`
	// Namespace is the object name.
	Name string `json:"name"`
	// ID is the object UID.
	ID string `json:"id"`
	// CausingObjects is the object that triggered this
	CausingObjects []string `json:"causingObjects,omitempty"`
}

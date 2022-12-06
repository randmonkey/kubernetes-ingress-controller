package konnect

const (
	NodeTypeIngressController = "ingress-controller"
	NodeTypeIngressProxy      = "ingress-proxy"
)

type NodeItem struct {
	ID                  string               `json:"id"`
	Version             string               `json:"version"`
	Hostname            string               `json:"hostname"`
	LastPing            int64                `json:"last_ping"`
	Type                string               `json:"type"`
	CreatedAt           int64                `json:"created_at"`
	UpdatedAt           int64                `json:"updated_at"`
	ConfigHash          string               `json:"config_hash"`
	CompatibilityStatus *CompatibilityStatus `json:"compatibility_status,omitempty"`
}

type CompatibilityStatus struct {
	State string `json:"state"`
	// TODO: add other fields
}

type CreateNodeRequest struct {
	ID                  string               `json:"id,omitempty"`
	Hostname            string               `json:"hostname"`
	Type                string               `json:"type"`
	LastPing            int64                `json:"last_ping"`
	Version             string               `json:"version"`
	CompatabilityStatus *CompatibilityStatus `json:"compatibility_status,omitempty"`
	ConfigHash          string               `json:"config_hash,omitempty"`
}

type CreateNodeResponse struct {
	Item *NodeItem `json:"item"`
}

type UpdateNodeRequest struct {
	Hostname   string `json:"hostname"`
	Type       string `json:"type"`
	LastPing   int64  `json:"last_ping"`
	Version    string `json:"version"`
	ConfigHash string `json:"config_hash,omitempty"`
}

type UpdateNodeResponse struct {
	Item *NodeItem `json:"item"`
}

type ListNodeResponse struct {
	Items []*NodeItem     `json:"items"`
	Page  *PaginationInfo `json:"page"`
}

type PaginationInfo struct {
	TotalCount  int32 `json:"total_count,omitempty"`
	NextPageNum int32 `json:"next_page_num,omitempty"`
}

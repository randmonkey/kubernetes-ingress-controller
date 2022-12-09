package konnect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AdminClient struct {
	Address   string
	ClusterID string
	Client    *http.Client
}

func (c *AdminClient) CreateNode(req *CreateNodeRequest) (*CreateNodeResponse, error) {
	buf, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create node request: %v", err)
	}
	reqReader := bytes.NewReader(buf)
	httpReq, err := http.NewRequest("POST", "http://"+c.Address+"/v1/nodes?cluster.id="+c.ClusterID, reqReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	httpResp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %v", err)
	}
	defer httpResp.Body.Close()

	respBuf, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if httpResp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("non-success response code from Koko: %d, resp body: %s", httpResp.StatusCode, string(respBuf))
		// TODO: parse returned body to return a more detailed error
	}

	resp := &CreateNodeResponse{}
	err = json.Unmarshal(respBuf, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON body: %v", err)
	}

	return resp, nil
}

func (c *AdminClient) UpdateNode(nodeID string, req *UpdateNodeRequest) (*UpdateNodeResponse, error) {
	buf, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal update node request: %v", err)
	}
	reqReader := bytes.NewReader(buf)
	httpReq, err := http.NewRequest("PUT", "http://"+c.Address+"/v1/nodes/"+nodeID+"?cluster.id="+c.ClusterID, reqReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request:%v", err)
	}
	httpResp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %v", err)
	}
	defer httpResp.Body.Close()

	respBuf, err := io.ReadAll(httpResp.Body)
	if err != nil {
		err := fmt.Errorf("failed to read response body: %v", err)
		return nil, err
	}

	if httpResp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("non-success response code from Koko: %d, resp body %s", httpResp.StatusCode, string(respBuf))
	}

	resp := &UpdateNodeResponse{}
	err = json.Unmarshal(respBuf, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON body: %v", err)
	}
	return resp, nil
}

func (c *AdminClient) ListNodes() (*ListNodeResponse, error) {
	httpResp, err := c.Client.Get("http://" + c.Address + "/v1/nodes?cluster.id=" + c.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %v", err)
	}

	defer httpResp.Body.Close()

	respBuf, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if httpResp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("non-success response from Koko: %d, resp body %s", httpResp.StatusCode, string(respBuf))
	}

	resp := &ListNodeResponse{}
	err = json.Unmarshal(respBuf, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	return resp, nil
}

func (c *AdminClient) DeleteNode(nodeID string) error {
	httpReq, err := http.NewRequest("DELETE", "http://"+c.Address+"/v1/nodes/"+nodeID+"?cluster.id="+c.ClusterID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request:%v", err)
	}
	httpResp, err := c.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to get response: %v", err)

	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode/100 != 2 {
		return fmt.Errorf("non-success response from Koko: %d", httpResp.StatusCode)
	}

	return nil
}

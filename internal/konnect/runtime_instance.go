package konnect

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
)

type RuntimeInstanceAgent struct {
	Address     string
	TLSCertPath string
	TLSKeyPath  string

	Hostname    string
	NodeID      string
	KongVersion string

	Logger logr.Logger
}

func (a *RuntimeInstanceAgent) dialWebsocket(url string) (*websocket.Conn, *http.Response, error) {
	dialer := websocket.DefaultDialer
	certContent, err := os.ReadFile(a.TLSCertPath)
	if err != nil {
		return nil, nil, err
	}
	keyContent, err := os.ReadFile(a.TLSKeyPath)
	if err != nil {
		return nil, nil, err
	}

	tlsCert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		return nil, nil, err
	}

	dialer.Subprotocols = []string{"wrpc.konghq.com"}
	dialer.TLSClientConfig = &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}
	return dialer.Dial(url, nil)
}

func (a *RuntimeInstanceAgent) RunWrpcWebsocket() error {
	socketURL := "wss://" + a.Address + "/v1/wrpc?" +
		fmt.Sprintf("node_id=%s&node_hostname=%s&node_version=%s",
			a.NodeID, a.Hostname, a.KongVersion)

	// currently dialWebsocket still gets a 400 response. Need more investigation here.
	conn, resp, err := a.dialWebsocket(socketURL)
	if resp != nil {
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			a.Logger.Error(err, "failed to read response body")
		}
		a.Logger.Info("got response from "+socketURL, "code", resp.StatusCode, "body", string(respBody))
		for k, v := range resp.Header {
			a.Logger.Info("resp header from "+socketURL, k, v)
		}
	}

	if err != nil {
		a.Logger.Error(err, "failed to connect to "+socketURL)
		return err
	}

	// TODO: keep the websocket alive
	defer conn.Close()
	return nil
}

func (a *RuntimeInstanceAgent) RunIngestWebsocket() error {
	socketURL := "wss://" + a.Address + "/v1/ingest?" +
		fmt.Sprintf("node_id=%s&node_hostname=%s&node_version=%s",
			a.NodeID, a.Hostname, a.KongVersion)

	conn, resp, err := a.dialWebsocket(socketURL)
	if resp != nil {
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			a.Logger.Error(err, "failed to read response body")
		}
		a.Logger.Info("got response from "+socketURL, "code", resp.StatusCode, "body", string(respBody))
		for k, v := range resp.Header {
			a.Logger.Info("resp header from "+socketURL, k, v)
		}
	}

	if err != nil {
		a.Logger.Error(err, "failed to connect to "+socketURL)
		return err
	}
	// TODO: keep the websocket alive
	defer conn.Close()
	return nil
}

func (a *RuntimeInstanceAgent) RunAnalyticsWebsocket() error {
	socketURL := "wss://" + a.Address + "/v1/analytics/reqlog?" +
		fmt.Sprintf("node_id=%s&node_hostname=%s&node_version=%s",
			a.NodeID, a.Hostname, a.KongVersion)

	conn, resp, err := a.dialWebsocket(socketURL)
	if resp != nil {
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			a.Logger.Error(err, "failed to read response body")
		}
		a.Logger.Info("got response from "+socketURL, "code", resp.StatusCode, "body", string(respBody))
		for k, v := range resp.Header {
			a.Logger.Info("resp header from "+socketURL, k, v)
		}
	}

	if err != nil {
		a.Logger.Error(err, "failed to connect to "+socketURL)
		return err
	}
	// TODO: keep the websocket alive
	defer conn.Close()
	return nil
}

func (a *RuntimeInstanceAgent) Run() {
	go a.RunWrpcWebsocket()
	// go a.RunIngestWebsocket()
	// go a.RunAnalyticsWebsocket()
}

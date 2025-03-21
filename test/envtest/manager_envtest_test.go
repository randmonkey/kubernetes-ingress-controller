//go:build envtest

package envtest

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest/observer"

	"github.com/kong/kubernetes-ingress-controller/v3/test/internal/helpers"
)

// TestManagerDoesntStartUntilKubernetesAPIReachable ensures that the manager and its Runnables are not start until the
// Kubernetes API server is reachable.
func TestManagerDoesntStartUntilKubernetesAPIReachable(t *testing.T) {
	t.Parallel()

	scheme := Scheme(t, WithKong)
	envcfg := Setup(t, scheme)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	t.Log("Setting up a proxy for Kubernetes API server so that we can interrupt it")
	u, err := url.Parse(envcfg.Host)
	require.NoError(t, err)
	apiServerProxy, err := helpers.NewTCPProxy(u.Host)
	require.NoError(t, err)
	go func() {
		err := apiServerProxy.Run(ctx)
		assert.NoError(t, err)
	}()
	apiServerProxy.StopHandlingConnections()

	t.Log("Replacing Kubernetes API server address with the proxy address")
	envcfg.Host = fmt.Sprintf("https://%s", apiServerProxy.Address())

	_, loggerHook := RunManager(ctx, t, envcfg, AdminAPIOptFns())
	hasLog := func(expectedLog string) bool {
		return lo.ContainsBy(loggerHook.All(), func(entry observer.LoggedEntry) bool {
			return strings.Contains(entry.Message, expectedLog)
		})
	}

	t.Log("Ensuring manager is waiting for Kubernetes API to be ready")
	const expectedKubernetesAPICheckErrorLog = "Retrying Kubernetes API readiness check after error"
	require.Eventually(t, func() bool { return hasLog(expectedKubernetesAPICheckErrorLog) }, time.Minute, time.Millisecond)

	t.Log("Ensure manager hasn't been started yet and no config sync has happened")
	const configurationSyncedToKongLog = "Successfully synced configuration to Kong"
	const startingManagerLog = "Starting manager"
	require.False(t, hasLog(configurationSyncedToKongLog))
	require.False(t, hasLog(startingManagerLog))

	t.Log("Starting accepting connections in Kubernetes API proxy so that manager can start")
	apiServerProxy.StartHandlingConnections()

	t.Log("Ensuring manager has been started and config sync has happened")
	require.Eventually(t, func() bool {
		return hasLog(startingManagerLog) &&
			hasLog(configurationSyncedToKongLog)
	}, time.Minute, time.Millisecond)
}

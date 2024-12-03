// Code generated by hack/generators/cache-stores/main.go; DO NOT EDIT.
// If you want to add a new type to the cache store, you need to add a new entry to the supportedTypes list in spec.go.
package store_test

import (
	"testing"

	kongv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	kongv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	"github.com/kong/kubernetes-ingress-controller/v3/internal/gatewayapi"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	netv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kongv1beta1 "github.com/kong/kubernetes-configuration/api/configuration/v1beta1"
	incubatorv1alpha1 "github.com/kong/kubernetes-configuration/api/incubator/v1alpha1"
	"github.com/kong/kubernetes-ingress-controller/v3/internal/store"
)

func TestCacheStores(t *testing.T) {
	testCases := []struct {
		name          string
		objectToStore client.Object
	}{

		{
			name:          "Ingress",
			objectToStore: &netv1.Ingress{},
		},

		{
			name:          "IngressClass",
			objectToStore: &netv1.IngressClass{},
		},

		{
			name:          "Service",
			objectToStore: &corev1.Service{},
		},

		{
			name:          "Secret",
			objectToStore: &corev1.Secret{},
		},

		{
			name:          "EndpointSlice",
			objectToStore: &discoveryv1.EndpointSlice{},
		},

		{
			name:          "HTTPRoute",
			objectToStore: &gatewayapi.HTTPRoute{},
		},

		{
			name:          "UDPRoute",
			objectToStore: &gatewayapi.UDPRoute{},
		},

		{
			name:          "TCPRoute",
			objectToStore: &gatewayapi.TCPRoute{},
		},

		{
			name:          "TLSRoute",
			objectToStore: &gatewayapi.TLSRoute{},
		},

		{
			name:          "GRPCRoute",
			objectToStore: &gatewayapi.GRPCRoute{},
		},

		{
			name:          "ReferenceGrant",
			objectToStore: &gatewayapi.ReferenceGrant{},
		},

		{
			name:          "Gateway",
			objectToStore: &gatewayapi.Gateway{},
		},

		{
			name:          "BackendTLSPolicy",
			objectToStore: &gatewayapi.BackendTLSPolicy{},
		},

		{
			name:          "KongPlugin",
			objectToStore: &kongv1.KongPlugin{},
		},

		{
			name:          "KongClusterPlugin",
			objectToStore: &kongv1.KongClusterPlugin{},
		},

		{
			name:          "KongConsumer",
			objectToStore: &kongv1.KongConsumer{},
		},

		{
			name:          "KongConsumerGroup",
			objectToStore: &kongv1beta1.KongConsumerGroup{},
		},

		{
			name:          "KongIngress",
			objectToStore: &kongv1.KongIngress{},
		},

		{
			name:          "TCPIngress",
			objectToStore: &kongv1beta1.TCPIngress{},
		},

		{
			name:          "UDPIngress",
			objectToStore: &kongv1beta1.UDPIngress{},
		},

		{
			name:          "KongUpstreamPolicy",
			objectToStore: &kongv1beta1.KongUpstreamPolicy{},
		},

		{
			name:          "IngressClassParameters",
			objectToStore: &kongv1alpha1.IngressClassParameters{},
		},

		{
			name:          "KongServiceFacade",
			objectToStore: &incubatorv1alpha1.KongServiceFacade{},
		},

		{
			name:          "KongVault",
			objectToStore: &kongv1alpha1.KongVault{},
		},

		{
			name:          "KongCustomEntity",
			objectToStore: &kongv1alpha1.KongCustomEntity{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := store.NewCacheStores()
			err := s.Add(tc.objectToStore)
			require.NoError(t, err)

			storedObj, ok, err := s.Get(tc.objectToStore)
			require.NoError(t, err)
			require.True(t, ok)
			require.Equal(t, tc.objectToStore, storedObj)

			err = s.Delete(tc.objectToStore)
			require.NoError(t, err)

			_, ok, err = s.Get(tc.objectToStore)
			require.NoError(t, err, err)
			require.False(t, ok)
		})
	}
}
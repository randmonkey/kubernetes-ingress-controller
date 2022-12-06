package konnect

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	refreshNodeInterval = 5 * time.Second
)

const (
	KongProxyContainerName = "proxy"
)

type RuntimeInstanceAgent struct {
	Address     string
	TLSCertPath string
	TLSKeyPath  string

	Hostname    string
	NodeID      string
	KongVersion string

	ClusterID string
	// service of Kong gateway.
	ServiceName string
	// namespace of service of Kong gateway.
	ServiceNamespace string
	K8sClient        client.Client

	Logger      logr.Logger
	adminClient *AdminClient
}

func (a *RuntimeInstanceAgent) createNode() error {
	createNodeReq := &CreateNodeRequest{
		ID:       a.NodeID,
		Hostname: a.Hostname,
		Type:     NodeTypeIngressController,
		Version:  a.KongVersion,
		LastPing: time.Now().Unix(),
	}
	resp, err := a.adminClient.CreateNode(createNodeReq)
	if err != nil {
		a.Logger.Error(err, "failed to create node")
	}
	a.NodeID = resp.Item.ID
	a.Logger.Info("created node for KIC", "cluster_id", a.ClusterID, "node_id", a.NodeID)
	return nil
}

func (a *RuntimeInstanceAgent) updateNode() error {
	updateNodeReq := &UpdateNodeRequest{
		Hostname: a.Hostname,
		Type:     NodeTypeIngressController,
		Version:  a.KongVersion,
		LastPing: time.Now().Unix(),
		// TODO: fill in config hash & compatibliity status.
	}
	_, err := a.adminClient.UpdateNode(a.NodeID, updateNodeReq)
	if err != nil {
		a.Logger.Error(err, "failed to update node for KIC")
		return err
	}
	a.Logger.Info("updated last ping time of node for KIC", "cluster_id", a.ClusterID, "node_id", a.NodeID)
	return nil
}

func (a *RuntimeInstanceAgent) updateKongProxyNodes() error {
	resp, err := a.adminClient.ListNodes()
	if err != nil {
		return fmt.Errorf("failed to list node in Koko: %v", err)
	}
	// TODO: retreive more nodes if there are more than one pages
	existingNodes := resp.Items
	// existingNodeMap is used to store existing nodes for kong proxy containers in Koko.
	// it maps hostname to nodeID. The node ID is used to call following update or delete API.
	existingNodeMap := make(map[string]string, len(resp.Items))
	for _, node := range existingNodes {
		if node.Type == NodeTypeIngressProxy {
			existingNodeMap[node.Hostname] = node.ID
		}
	}

	// get all pods in the publish service.
	publishService := &corev1.Service{}
	a.Logger.Info("get service for kong proxy", "service_namespace", a.ServiceNamespace, "service_name", a.ServiceName)
	err = a.K8sClient.Get(context.Background(), types.NamespacedName{Namespace: a.ServiceNamespace, Name: a.ServiceName}, publishService)
	if err != nil {
		return fmt.Errorf("failed to get kong admin service: %v", err)
	}

	// list kong gateway pods and refresh their status.
	kongProxyPods := corev1.PodList{}
	err = a.K8sClient.List(
		context.Background(), &kongProxyPods,
		// pods should in the namespace of publish service.
		client.InNamespace(a.ServiceNamespace),
		// pods should match selector of the service.
		client.MatchingLabelsSelector{
			Selector: labels.SelectorFromSet(publishService.Labels),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to list kong gateway pods: %v", err)
	}
	// kongProxyPodMap used to store kong proxy pods by their names.
	// if hostname of an existing node in the cluster does not exist here,
	// that means the node does not map to a kong proxy pod in k8s cluster,
	// so the node should be deleted.
	var kongProxyPodMap = make(map[string]struct{}, len(kongProxyPods.Items))

	for _, pod := range kongProxyPods.Items {
		a.Logger.Info("create/update for kong proxy pod", "pod_namespace", pod.Namespace, "pod_name", pod.Name)
		podName := pod.Name
		var kongProxyContainerSpec *corev1.Container
		for i, container := range pod.Spec.Containers {
			if container.Name == KongProxyContainerName {
				kongProxyContainerSpec = &pod.Spec.Containers[i]
				break
			}
		}
		if kongProxyContainerSpec == nil {
			continue
		}

		kongProxyPodMap[podName] = struct{}{}
		nodeID, ok := existingNodeMap[podName]
		if !ok {
			// create a node if the node with pod name as its hostname does not exist.
			createNodeReq := &CreateNodeRequest{
				Hostname: podName,
				Type:     NodeTypeIngressProxy,
				// TODO: get kong version from the container, and use it here.
				Version:  a.KongVersion,
				LastPing: time.Now().Unix(),
			}
			resp, err := a.adminClient.CreateNode(createNodeReq)
			if err != nil {
				return fmt.Errorf("failed to create node for kong pod %s: %v", podName, err)
			}
			a.Logger.Info("created node for kong proxy instance", "pod_name", podName, "node_id", resp.Item.ID, "cluster_id", a.ClusterID)
		} else {
			updateNodeReq := &UpdateNodeRequest{
				Hostname: podName,
				Type:     NodeTypeIngressProxy,
				// TODO: get kong version from the container, and use it here.
				Version:  a.KongVersion,
				LastPing: time.Now().Unix(),
			}
			_, err := a.adminClient.UpdateNode(nodeID, updateNodeReq)
			if err != nil {
				return fmt.Errorf("failed to update node for kong pod %s: %v", podName, err)
			}
			a.Logger.Info("updated node for kong proxy instance", "pod_name", podName, "node_id", nodeID, "cluster_id", a.ClusterID)
		}
	}

	// delete nodes not representing existing kong proxy pods.
	for _, node := range existingNodes {
		_, ok := kongProxyPodMap[node.Hostname]
		if !ok {
			err = a.adminClient.DeleteNode(node.ID)
			if err != nil {
				return fmt.Errorf("failed to outdated delete node %s: %v", node.ID, err)
			}
			a.Logger.Info("deleted outdated node for kong proxy", "node_hostname", node.Hostname, "node_id", node.ID, "cluster_id", a.ClusterID)
		}
	}

	return nil
}

func (a *RuntimeInstanceAgent) updateNodeLoop() {
	ticker := time.NewTicker(refreshNodeInterval)
	for range ticker.C {
		err := a.updateNode()
		if err != nil {
			a.Logger.Error(err, "failed to update node", "cluster_id", a.ClusterID, "node_id", a.NodeID)
		}
		err = a.updateKongProxyNodes()
		if err != nil {
			a.Logger.Error(err, "failed to update kong proxy nodes", "cluster_id", a.ClusterID)
		}
	}
}

func (a *RuntimeInstanceAgent) Run() {
	a.adminClient = &AdminClient{
		Address:   a.Address,
		ClusterID: a.ClusterID,
		Client:    &http.Client{},
	}
	err := a.createNode()
	if err != nil {
		a.Logger.Error(err, "failed to create node, agent abort")
		return
	}

	go a.updateNodeLoop()
}

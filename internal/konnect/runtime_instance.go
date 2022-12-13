package konnect

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/kubernetes-ingress-controller/v2/internal/dataplane/failures"
)

const (
	refreshNodeInterval = 5 * time.Second
)

const (
	KongProxyContainerName = "proxy"
)

type RuntimeInstanceAgent struct {
	Address string
	// mTLS certificate pair, not used in current Demo,
	// but may be used for mTLS authentication and RG registration.
	// TLSCertPath string
	// TLSKeyPath  string

	Hostname    string
	NodeID      string
	KongVersion string

	ClusterID string
	// service of Kong gateway.
	ServiceName string
	// namespace of service of Kong gateway.
	ServiceNamespace string
	K8sClient        client.Client

	currentConfigHash string
	ConfigHashChan    chan []byte

	translationFailures    []failures.ResourceFailure
	TranslationFailureChan chan []failures.ResourceFailure

	kongUpdateErr     error
	KongUpdateErrChan chan error

	Logger      logr.Logger
	adminClient *AdminClient
}

// createNode creates the node representing kuberenetes ingress controller itself.
func (a *RuntimeInstanceAgent) createNode() error {
	// remove existing node with type "ingress-controller".
	nodes, err := a.adminClient.ListNodes()
	if err != nil {
		return fmt.Errorf("failed to list existing nodes: %v", err)
	}

	for _, node := range nodes.Items {
		if node.Type == NodeTypeIngressController {
			// TODO: judge whether we should delete this node.
			err := a.adminClient.DeleteNode(node.ID)
			if err != nil {
				return fmt.Errorf("failed to delete outdated node %s: %v", node.ID, err)
			}
			a.Logger.Info("deleted outdated node for KIC", "node_hostname", node.Hostname, "node_id", node.ID, "cluster_id", a.ClusterID)
		}
	}

	createNodeReq := &CreateNodeRequest{
		ID:       a.NodeID,
		Hostname: a.Hostname,
		Type:     NodeTypeIngressController,
		Version:  a.KongVersion,
		LastPing: time.Now().Unix(),
		// TODO: get the real running status and fill it here.
		IngressControllerStatus: &IngressControllerStatus{
			State: IngressControllerStateOperational,
		},
	}
	resp, err := a.adminClient.CreateNode(createNodeReq)
	if err != nil {
		a.Logger.Error(err, "failed to create node")
		return err
	}
	a.NodeID = resp.Item.ID
	a.Logger.Info("created node for KIC", "cluster_id", a.ClusterID, "node_id", a.NodeID)
	return nil
}

func (a *RuntimeInstanceAgent) updateNode() error {
	ingressControllerStatus := &IngressControllerStatus{
		State: IngressControllerStateOperational,
	}

	if a.kongUpdateErr != nil {
		ingressControllerStatus.State = IngressControllerStateInoperable
	} else {
		if len(a.translationFailures) > 0 {
			ingressControllerStatus.State = IngressControllerStatePartialConfigFail
		}
	}
	for _, translationFailure := range a.translationFailures {
		objects := translationFailure.CausingObjects()
		for _, object := range objects {
			ingressControllerStatus.Issues = append(ingressControllerStatus.Issues, &KubernetesObjectIssue{
				Group:       object.GetObjectKind().GroupVersionKind().Group,
				Version:     object.GetObjectKind().GroupVersionKind().Version,
				Kind:        object.GetObjectKind().GroupVersionKind().Kind,
				Namespace:   object.GetNamespace(),
				Name:        object.GetName(),
				Description: translationFailure.Message(),
			})
		}
	}

	updateNodeReq := &UpdateNodeRequest{
		Hostname:                a.Hostname,
		Type:                    NodeTypeIngressController,
		Version:                 a.KongVersion,
		LastPing:                time.Now().Unix(),
		ConfigHash:              a.currentConfigHash,
		IngressControllerStatus: ingressControllerStatus,
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
				Hostname:   podName,
				Type:       NodeTypeIngressProxy,
				Version:    a.KongVersion,
				LastPing:   time.Now().Unix(),
				ConfigHash: a.currentConfigHash,
				// TODO: get the real compatibility status from kong container.
				CompatabilityStatus: &CompatibilityStatus{
					State: CompatibilityStateFullyCompatible,
				},
			}
			resp, err := a.adminClient.CreateNode(createNodeReq)
			if err != nil {
				return fmt.Errorf("failed to create node for kong pod %s: %v", podName, err)
			}
			a.Logger.Info("created node for kong proxy instance", "pod_name", podName, "node_id", resp.Item.ID, "cluster_id", a.ClusterID)
		} else {
			updateNodeReq := &UpdateNodeRequest{
				Hostname:   podName,
				Type:       NodeTypeIngressProxy,
				Version:    a.KongVersion,
				LastPing:   time.Now().Unix(),
				ConfigHash: a.currentConfigHash,
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

func (a *RuntimeInstanceAgent) receiveFromManager() {
	for {
		select {
		// receive config hash
		case hashSum := <-a.ConfigHashChan:
			hashStr := hex.EncodeToString(hashSum)
			a.Logger.Info("updated config hash", "hash", hashStr)
			if len(hashStr) > 32 {
				hashStr = hashStr[:32]
			}
			a.currentConfigHash = hashStr
		// receive translation failures
		case translationFailures := <-a.TranslationFailureChan:
			a.Logger.Info(fmt.Sprintf("%d translation failures", len(translationFailures)))
			a.translationFailures = translationFailures
		// receive errors on kong config
		case updateErr := <-a.KongUpdateErrChan:
			if updateErr != nil {
				a.Logger.Info("failed to update config on Kong", "error", updateErr)
			}
			a.kongUpdateErr = updateErr
		}
	}
}

func (a *RuntimeInstanceAgent) Run() {
	// TODO: run a leader election process, and only the instance
	// which became the leader can run the following node
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
	go a.receiveFromManager()
}

/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	UsageAnnotation  = "usage"
	DefaultNamespace = "default"
	//UsageLabel       = "node-usage"
	UsageLabel    = "node.cnstack.alibabacloud.com/usage-info"
	KeyNodePrefix = "node-"
)

// NodeMetricsReconciler reconciles a NodeMetrics object
type NodeMetricsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=metrics.k8s.io,resources=nodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=metrics.k8s.io,resources=nodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=metrics.k8s.io,resources=nodes/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NodeMetrics object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *NodeMetricsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	// TODO(user): your logic here
	//klog.Infof("req %v", req)
	nodeMetrics := &metricsv1beta1.NodeMetrics{}
	if err := r.Client.Get(ctx, req.NamespacedName, nodeMetrics); err != nil {
		klog.Error(err, "list node metrics failed")
		return ctrl.Result{}, err
	}
	nodeList := &v1.NodeList{}
	if err := r.Client.List(ctx, nodeList); err != nil {
		klog.Error(err, "list nodes failed")
		return ctrl.Result{}, err
	}
	podList := &v1.PodList{}
	if err := r.Client.List(ctx, podList); err != nil {
		klog.Error(err, "list pods failed")
		return ctrl.Result{}, err
	}

	configmapList := &v1.ConfigMapList{}
	if err := r.Client.List(ctx, configmapList, &ctrlclient.ListOptions{
		Namespace:     DefaultNamespace,
		LabelSelector: labels.SelectorFromSet(map[string]string{UsageLabel: ""})}); err != nil {
		klog.Error(err, "list old cm failed")
		return ctrl.Result{}, err
	}
	// createNodeUsages
	newNodeUsages := r.createNodeUsageInfo(*nodeList, *nodeMetrics, *podList)

	// compare newUsage and OldUsage get create,update,del
	createUsages, updateUsages, deleteUsages := r.getUpdateNodeUsageInfo(newNodeUsages, configmapList.Items)

	if len(createUsages) != 0 {
		for _, v := range createUsages {
			if err := r.Client.Create(ctx, v.DeepCopy()); err != nil {
				klog.Error("create %s node's usage info cm %s failed %v", nodeMetrics.Name, v.Name, err.Error())
				return ctrl.Result{}, err
			}
		}
	}

	if len(updateUsages) != 0 {
		for _, v := range updateUsages {
			if err := r.Client.Update(ctx, v.DeepCopy()); err != nil {
				klog.Error("update %s node's usage info cm %s failed %v", nodeMetrics.Name, v.Name, err.Error())
				return ctrl.Result{}, err
			}
		}
	}

	if len(deleteUsages) != 0 {
		for _, v := range deleteUsages {
			if err := r.Client.Delete(ctx, v.DeepCopy()); err != nil {
				klog.Error("delete no need node usage info cm %s failed %v", v.Name, err.Error())
				return ctrl.Result{}, err
			}
		}
	}

	//podMap := map[string][]v1.Pod{}
	//for _, pod := range podList.Items {
	//	if pod.Spec.NodeName != "" && pod.Status.Phase != v1.PodSucceeded && pod.Status.Phase != v1.PodFailed {
	//		l := podMap[pod.Spec.NodeName]
	//		podMap[pod.Spec.NodeName] = append(l, pod)
	//	}
	//}
	//nodeLen := len(nodeList.Items)
	//usageConfigMapList := make([]v1.ConfigMap, 0)
	//i, k, splitLen, data := 0, 0, 2000, make(map[string]string)
	//for i < nodeLen {
	//	node := nodeList.Items[i]
	//	nodePodList := podMap[node.Name]
	//	metrics := nodeMetrics
	//	if metrics.Name == node.Name {
	//		resourceUsage := getNodeResourceUsages(node, *nodeMetrics, nodePodList)
	//		bytes, _ := json.Marshal(resourceUsage)
	//		data[node.Name] = string(bytes)
	//	}
	//	if i%splitLen == splitLen-1 || i == nodeLen-1 {
	//		cm := v1.ConfigMap{
	//			ObjectMeta: metav1.ObjectMeta{
	//				Labels: map[string]string{
	//					UsageLabel: "",
	//				},
	//				Name:      "resource-usage-" + fmt.Sprintf("%d", k),
	//				Namespace: DefaultNamespace,
	//			},
	//			Data: data,
	//		}
	//		usageConfigMapList = append(usageConfigMapList, cm)
	//		data = make(map[string]string)
	//		k++
	//	}
	//
	//	i++
	//}
	//oldUsageConfigMaps := &v1.ConfigMapList{}
	//if err := r.Client.List(ctx, oldUsageConfigMaps, &ctrlclient.ListOptions{
	//	LabelSelector: labels.SelectorFromSet(map[string]string{UsageLabel: ""})}); err != nil {
	//	klog.Error(err, "list old cm failed")
	//	return ctrl.Result{}, err
	//}
	//existedMap := map[string]bool{}
	//for _, v := range usageConfigMapList {
	//	var existedItem *v1.ConfigMap
	//	for _, item := range oldUsageConfigMaps.Items {
	//		if v.Name == item.Name {
	//			existedItem = item.DeepCopy()
	//			existedMap[v.Name] = true
	//			break
	//		}
	//	}
	//	if existedItem != nil {
	//		newData := existedItem.DeepCopy().Data
	//		if u, ok := v.Data[nodeMetrics.Name]; ok {
	//			newData[nodeMetrics.Name] = u
	//		}
	//		updateItem := v.DeepCopy()
	//		updateItem.Data = newData
	//		// if no diff update
	//		if !reflect.DeepEqual(existedItem.Data, newData) {
	//			if err := r.Client.Update(ctx, updateItem); err != nil {
	//				klog.Errorf("update %s node usage cm failed %v", existedItem.Name, err.Error())
	//				return ctrl.Result{}, err
	//			} else {
	//				klog.Infof("update existed usage %s \n old data is %s \n new data is %s\n", nodeMetrics.Name, existedItem.Data, newData)
	//			}
	//		}
	//	} else {
	//		tmpV := v.DeepCopy()
	//		if err := r.Client.Create(ctx, tmpV); err != nil {
	//			klog.Errorf("create %s node usage cm failed %v", tmpV.Name, err.Error())
	//			return ctrl.Result{}, err
	//		}
	//	}
	//}
	//// sync map
	//for _, item := range oldUsageConfigMaps.Items {
	//	if _, ok := existedMap[item.Name]; !ok {
	//		delItem := &item
	//		if err := r.Client.Delete(ctx, delItem); err != nil && !errors.IsNotFound(err) {
	//			klog.Errorf("delete no need %s node usage cm failed , error is %v", item.Name, err.Error())
	//			return ctrl.Result{}, err
	//		}
	//	}
	//}
	return ctrl.Result{}, nil
}

// createNodeUsageInfo
// put node metrics usage,allocatable,requests,limits to the configmaps with the constants.NodeUsageLabelKey
func (r *NodeMetricsReconciler) createNodeUsageInfo(nodeList v1.NodeList, nodeMetrics metricsv1beta1.NodeMetrics, podList v1.PodList) []v1.ConfigMap {
	podMap := map[string][]v1.Pod{}
	for _, pod := range podList.Items {
		if pod.Spec.NodeName != "" && pod.Status.Phase != v1.PodSucceeded && pod.Status.Phase != v1.PodFailed {
			l := podMap[pod.Spec.NodeName]
			podMap[pod.Spec.NodeName] = append(l, pod)
		}
	}
	// init node usages per 2000 split
	nodeUsages, nodeLen := make([]v1.ConfigMap, 0), len(nodeList.Items)
	i, k, splitLen, data := 0, 0, 2000, make(map[string]string)
	for i < nodeLen {
		node := nodeList.Items[i]
		nodePodList := podMap[node.Name]
		if nodeMetrics.Name == node.Name {
			resourceUsage := getNodeResourceUsages(node, nodeMetrics, nodePodList)
			bytes, _ := json.Marshal(resourceUsage)
			data[node.Name] = string(bytes)
		}
		if i%splitLen == splitLen-1 || i == nodeLen-1 {
			cm := v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						UsageLabel: "",
					},
					Name:      "resource-usage-" + fmt.Sprintf("%d", k),
					Namespace: DefaultNamespace,
				},
				Data: data,
			}
			nodeUsages = append(nodeUsages, cm)
			data = make(map[string]string)
			k++
		}
		i++
	}
	return nodeUsages
}

// getUpdateNodeUsageInfo
// compare diff between newNodeUsages and oldNodeUsages then get updateNodeUsages
// return createUsages updateUsages deleteUsages []v1.ConfigMaps
func (r *NodeMetricsReconciler) getUpdateNodeUsageInfo(newNodeUsages []v1.ConfigMap, oldNodeUsages []v1.ConfigMap) (createUsages []v1.ConfigMap, updateUsages []v1.ConfigMap, deleteUsages []v1.ConfigMap) {
	createUsages, updateUsages, deleteUsages = make([]v1.ConfigMap, 0), make([]v1.ConfigMap, 0), make([]v1.ConfigMap, 0)
	for _, n := range newNodeUsages {
		var oldUsage *v1.ConfigMap
		for _, o := range oldNodeUsages {
			if n.Name == o.Name {
				oldUsage = &o
				break
			}
		}
		if oldUsage == nil {
			newUsage := n
			createUsages = append(createUsages, newUsage)
		} else {
			oldData, newData := oldUsage.DeepCopy().Data, oldUsage.DeepCopy().Data
			for k, v := range n.Data {
				newData[k] = v
			}
			if !reflect.DeepEqual(oldData, newData) {
				klog.Infof("old data %v\n new data %v\n", oldData, newData)
				newUsage := oldUsage.DeepCopy()
				newUsage.Data = newData
				updateUsages = append(updateUsages, *newUsage)
			}
		}
	}
	for _, o := range oldNodeUsages {
		isNotNeed := true
		for _, n := range oldNodeUsages {
			if n.Name == o.Name {
				isNotNeed = false
				break
			}
		}
		if isNotNeed {
			deleteUsages = append(deleteUsages, o)
		}
	}
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeMetricsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&metricsv1beta1.NodeMetrics{}).
		Complete(r)
}

func getNodeResourceUsages(node v1.Node, nodeMetrics metricsv1beta1.NodeMetrics, podList []v1.Pod) ResourceUsages {
	resourceUsages := ResourceUsages{
		CPU:              ResourceUsage{},
		Memory:           ResourceUsage{},
		EphemeralStorage: ResourceUsage{},
		PodNum:           ResourceUsage{},
	}
	allocatable := node.Status.Allocatable
	if len(allocatable) == 0 {
		allocatable = node.Status.Capacity
	}
	usage := nodeMetrics.Usage
	reqs, limits := GetPodsTotalRequestsAndLimits(&v1.PodList{Items: podList})
	resourceUsages.CPU = func() ResourceUsage {
		req, limit, alloc, usage := reqs[v1.ResourceCPU], limits[v1.ResourceCPU], allocatable[v1.ResourceCPU], usage[v1.ResourceCPU]
		return ResourceUsage{
			Reqs:        req.MilliValue(),
			Limits:      limit.MilliValue(),
			Allocatable: alloc.MilliValue(),
			Usage:       usage.MilliValue(),
		}
	}()
	resourceUsages.Memory = func() ResourceUsage {
		req, limit, alloc, usage := reqs[v1.ResourceMemory], limits[v1.ResourceMemory], allocatable[v1.ResourceMemory], usage[v1.ResourceMemory]
		return ResourceUsage{
			Reqs:        req.Value(),
			Limits:      limit.Value(),
			Allocatable: alloc.Value(),
			Usage:       usage.Value(),
		}
	}()
	resourceUsages.PodNum = func() ResourceUsage {
		return ResourceUsage{
			Allocatable: allocatable.Pods().Value(),
			Usage:       int64(len(podList)),
		}
	}()
	resourceUsages.EphemeralStorage = func() ResourceUsage {
		req, limit, alloc := reqs[v1.ResourceEphemeralStorage], limits[v1.ResourceEphemeralStorage], allocatable[v1.ResourceEphemeralStorage]
		return ResourceUsage{
			Reqs:        req.Value(),
			Limits:      limit.Value(),
			Allocatable: alloc.Value(),
			Usage:       req.Value(),
		}
	}()
	return resourceUsages

}

type ResourceUsage struct {
	Reqs        int64 `json:"reqs"`
	Limits      int64 `json:"limits"`
	Allocatable int64 `json:"allocatable"`
	Usage       int64 `json:"usage"`
}

type ResourceUsages struct {
	CPU              ResourceUsage `json:"cpu"`
	Memory           ResourceUsage `json:"memory"`
	EphemeralStorage ResourceUsage `json:"ephemeralStorage"`
	PodNum           ResourceUsage `json:"podNum"`
}

// PodRequestsAndLimits returns a dictionary of all defined resources summed up for all
// containers of the pod. If pod overhead is non-nil, the pod overhead is added to the
// total container resource requests and to the total container limits which have a
// non-zero quantity.
func PodRequestsAndLimits(pod *v1.Pod) (reqs, limits v1.ResourceList) {
	reqs, limits = v1.ResourceList{}, v1.ResourceList{}
	for _, container := range pod.Spec.Containers {
		addResourceList(reqs, container.Resources.Requests)
		addResourceList(limits, container.Resources.Limits)
	}
	// init containers define the minimum of any resource
	for _, container := range pod.Spec.InitContainers {
		maxResourceList(reqs, container.Resources.Requests)
		maxResourceList(limits, container.Resources.Limits)
	}

	// Add overhead for running a pod to the sum of requests and to non-zero limits:
	if pod.Spec.Overhead != nil {
		addResourceList(reqs, pod.Spec.Overhead)

		for name, quantity := range pod.Spec.Overhead {
			if value, ok := limits[name]; ok && !value.IsZero() {
				value.Add(quantity)
				limits[name] = value
			}
		}
	}
	return
}

// addResourceList adds the resources in newList to list
func addResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
		} else {
			value.Add(quantity)
			list[name] = value
		}
	}
}

// maxResourceList sets list to the greater of list/newList for every resource
// either list
func maxResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
			continue
		} else {
			if quantity.Cmp(value) > 0 {
				list[name] = quantity.DeepCopy()
			}
		}
	}
}

func GetPodsTotalRequestsAndLimits(podList *v1.PodList) (reqs map[v1.ResourceName]resource.Quantity, limits map[v1.ResourceName]resource.Quantity) {
	reqs, limits = map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}
	for _, pod := range podList.Items {
		podReqs, podLimits := resourcehelper.PodRequestsAndLimits(&pod)
		for podReqName, podReqValue := range podReqs {
			if value, ok := reqs[podReqName]; !ok {
				reqs[podReqName] = podReqValue.DeepCopy()
			} else {
				value.Add(podReqValue)
				reqs[podReqName] = value
			}
		}
		for podLimitName, podLimitValue := range podLimits {
			if value, ok := limits[podLimitName]; !ok {
				limits[podLimitName] = podLimitValue.DeepCopy()
			} else {
				value.Add(podLimitValue)
				limits[podLimitName] = value
			}
		}
	}
	return
}

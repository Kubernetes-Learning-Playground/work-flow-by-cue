package k8s_utils

import (
	"bytes"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/util/sets"
	"log"
	"strconv"
	"strings"
	"time"
)

var PrinterMap map[schema.GroupVersionKind]func(obj *unstructured.Unstructured) *metav1.Table

func init() {
	//初始化。 注册 类型 对应的  解析函数
	PrinterMap = make(map[schema.GroupVersionKind]func(obj *unstructured.Unstructured) *metav1.Table)
	PrinterMap[schema.GroupVersionKind{"apps", "v1", "Deployment"}] = PrintDeployment
	PrinterMap[schema.GroupVersionKind{"", "v1", "Service"}] = PrintService

}

// 打印 对象 返回table
func PrintObject(obj *unstructured.Unstructured) *metav1.Table {
	if f, ok := PrinterMap[obj.GroupVersionKind()]; ok {
		return f(obj)
	}
	return defaultTable(obj)
}

// 对于没有注册实现的  显示默认的  只有名称、创建时间什么的
func defaultTable(obj *unstructured.Unstructured) *metav1.Table {
	defaultColumns := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
	}
	t := &metav1.Table{
		ColumnDefinitions: defaultColumns,
	}
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}
	row.Cells = append(row.Cells, obj.GetName(), translateTimestampSince(obj.GetCreationTimestamp()))
	t.Rows = []metav1.TableRow{row}
	return t
}

func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}
	return duration.HumanDuration(time.Since(timestamp.Time))
}

func layoutContainerCells(containers []apiv1.Container) (names string, images string) {
	var namesBuffer bytes.Buffer
	var imagesBuffer bytes.Buffer

	for i, container := range containers {
		namesBuffer.WriteString(container.Name)
		imagesBuffer.WriteString(container.Image)
		if i != len(containers)-1 {
			namesBuffer.WriteString(",")
			imagesBuffer.WriteString(",")
		}
	}
	return namesBuffer.String(), imagesBuffer.String()
}

func PrintDeployment(obj *unstructured.Unstructured) *metav1.Table {
	deploy := &appv1.Deployment{}
	err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(obj.UnstructuredContent(), deploy)
	if err != nil {
		log.Println("转换deploy出错:", err.Error())
		return nil
	}

	deploymentColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Ready", Type: "string", Description: "Number of the pod with ready state"},
		{Name: "Up-to-date", Type: "string", Description: extensionsv1beta1.DeploymentStatus{}.SwaggerDoc()["updatedReplicas"]},
		{Name: "Available", Type: "string", Description: extensionsv1beta1.DeploymentStatus{}.SwaggerDoc()["availableReplicas"]},
		{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
		{Name: "Containers", Type: "string", Priority: 1, Description: "Names of each container in the template."},
		{Name: "Images", Type: "string", Priority: 1, Description: "Images referenced by each container in the template."},
		{Name: "Selector", Type: "string", Priority: 1, Description: extensionsv1beta1.DeploymentSpec{}.SwaggerDoc()["selector"]},
	}
	t := &metav1.Table{
		ColumnDefinitions: deploymentColumnDefinitions,
	}
	t.Kind = "Table"
	t.APIVersion = "meta.k8s.io/v1"
	rows, err := printDeployment(deploy) // 得到行
	if err != nil {
		log.Println(err)
		return nil
	}
	t.Rows = rows
	return t
}

func printDeployment(obj *appv1.Deployment) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj}, //这一步有待商榷， 是否一定要把对象返回出去。暂时先返回 。后台无所谓的带宽性能
	}
	desiredReplicas := obj.Spec.Replicas
	updatedReplicas := obj.Status.UpdatedReplicas
	readyReplicas := obj.Status.ReadyReplicas
	availableReplicas := obj.Status.AvailableReplicas
	age := translateTimestampSince(obj.CreationTimestamp)
	containers := obj.Spec.Template.Spec.Containers
	selector, err := metav1.LabelSelectorAsSelector(obj.Spec.Selector)
	selectorString := ""
	if err != nil {
		selectorString = "<invalid>"
	} else {
		selectorString = selector.String()
	}
	getContainers, images := layoutContainerCells(containers)
	row.Cells = append(row.Cells, obj.Name,
		fmt.Sprintf("%d/%d", int64(readyReplicas), int64(*desiredReplicas)),
		int64(updatedReplicas), int64(availableReplicas), age, getContainers,
		images, selectorString)

	return []metav1.TableRow{row}, nil
}

func PrintService(obj *unstructured.Unstructured) *metav1.Table {
	svc := &apiv1.Service{}
	err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(obj.UnstructuredContent(), svc)
	if err != nil {
		log.Println("转换Service出错:", err.Error())
		return nil
	}

	serviceColumnDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Type", Type: "string", Description: apiv1.ServiceSpec{}.SwaggerDoc()["type"]},
		{Name: "Cluster-IP", Type: "string", Description: apiv1.ServiceSpec{}.SwaggerDoc()["clusterIP"]},
		{Name: "External-IP", Type: "string", Description: apiv1.ServiceSpec{}.SwaggerDoc()["externalIPs"]},
		{Name: "Port(s)", Type: "string", Description: apiv1.ServiceSpec{}.SwaggerDoc()["ports"]},
		{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
		{Name: "Selector", Type: "string", Priority: 1, Description: apiv1.ServiceSpec{}.SwaggerDoc()["selector"]},
	}
	t := &metav1.Table{
		ColumnDefinitions: serviceColumnDefinitions,
	}
	t.Kind = "Table"
	t.APIVersion = "meta.k8s.io/v1"
	rows, err := printService(svc) // 得到行
	if err != nil {
		log.Println(err)
		return nil
	}
	t.Rows = rows
	return t
}

const loadBalancerWidth = 16

func makePortString(ports []apiv1.ServicePort) string {
	pieces := make([]string, len(ports))
	for ix := range ports {
		port := &ports[ix]
		pieces[ix] = fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		if port.NodePort > 0 {
			pieces[ix] = fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol)
		}
	}
	return strings.Join(pieces, ",")
}

func loadBalancerStatusStringer(s apiv1.LoadBalancerStatus, wide bool) string {
	ingress := s.Ingress
	result := sets.NewString()
	for i := range ingress {
		if ingress[i].IP != "" {
			result.Insert(ingress[i].IP)
		} else if ingress[i].Hostname != "" {
			result.Insert(ingress[i].Hostname)
		}
	}

	r := strings.Join(result.List(), ",")
	if !wide && len(r) > loadBalancerWidth {
		r = r[0:(loadBalancerWidth-3)] + "..."
	}
	return r
}

func getServiceExternalIP(svc *apiv1.Service, wide bool) string {
	switch svc.Spec.Type {
	case apiv1.ServiceTypeClusterIP:
		if len(svc.Spec.ExternalIPs) > 0 {
			return strings.Join(svc.Spec.ExternalIPs, ",")
		}
		return "<none>"
	case apiv1.ServiceTypeNodePort:
		if len(svc.Spec.ExternalIPs) > 0 {
			return strings.Join(svc.Spec.ExternalIPs, ",")
		}
		return "<none>"
	case apiv1.ServiceTypeLoadBalancer:
		lbIps := loadBalancerStatusStringer(svc.Status.LoadBalancer, wide)
		if len(svc.Spec.ExternalIPs) > 0 {
			results := []string{}
			if len(lbIps) > 0 {
				results = append(results, strings.Split(lbIps, ",")...)
			}
			results = append(results, svc.Spec.ExternalIPs...)
			return strings.Join(results, ",")
		}
		if len(lbIps) > 0 {
			return lbIps
		}
		return "<pending>"
	case apiv1.ServiceTypeExternalName:
		return svc.Spec.ExternalName
	}
	return "<unknown>"
}

func printService(obj *apiv1.Service) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}
	svcType := obj.Spec.Type
	internalIP := "<none>"
	if len(obj.Spec.ClusterIPs) > 0 {
		internalIP = obj.Spec.ClusterIPs[0]
	}

	externalIP := getServiceExternalIP(obj, true)
	svcPorts := makePortString(obj.Spec.Ports)
	if len(svcPorts) == 0 {
		svcPorts = "<none>"
	}

	row.Cells = append(row.Cells, obj.Name, string(svcType), internalIP, externalIP, svcPorts, translateTimestampSince(obj.CreationTimestamp), labels.FormatLabels(obj.Spec.Selector))

	return []metav1.TableRow{row}, nil
}

func printPod(pod *apiv1.Pod) ([]metav1.TableRow, error) {
	restarts := 0
	totalContainers := len(pod.Spec.Containers)
	readyContainers := 0
	lastRestartDate := metav1.NewTime(time.Time{})

	reason := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}

	row := metav1.TableRow{
		//Object: runtime.RawExtension{Object: pod},
	}

	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		restarts += int(container.RestartCount)
		if container.LastTerminationState.Terminated != nil {
			terminatedDate := container.LastTerminationState.Terminated.FinishedAt
			if lastRestartDate.Before(&terminatedDate) {
				lastRestartDate = terminatedDate
			}
		}
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		restarts = 0
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			restarts += int(container.RestartCount)
			if container.LastTerminationState.Terminated != nil {
				terminatedDate := container.LastTerminationState.Terminated.FinishedAt
				if lastRestartDate.Before(&terminatedDate) {
					lastRestartDate = terminatedDate
				}
			}
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {

				readyContainers++
			}
		}

	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		reason = "Unknown"
	} else if pod.DeletionTimestamp != nil {
		reason = "Terminating"
	}

	restartsStr := strconv.Itoa(restarts)
	if !lastRestartDate.IsZero() {
		restartsStr = fmt.Sprintf("%d (%s ago)", restarts, translateTimestampSince(lastRestartDate))

	}

	row.Cells = append(row.Cells, pod.Name, fmt.Sprintf("%d/%d", readyContainers, totalContainers), reason, restartsStr, translateTimestampSince(pod.CreationTimestamp))

	return []metav1.TableRow{row}, nil
}

package handlers

import (
	"bytes"
	"context"
	"cuelang.org/go/cue"
	"cuelang.org/go/tools/flow"
	"fmt"
	"io"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	// FIXME: 需要改到更通用的位置，目前已经改为接口注册
	PodFlowTpl  = "./work_flow_template/workflow.cue"
	PodFlowRoot = "workflow"
)

//func init() {
//	// TODO: 注册转为 接口注册
//	f := NewFlowFunc(PodFlowTpl, PodFlowRoot)
//	Register("workflow", "工作流操作", f)
//}

func getPodLogs(obj *resource.Info) string {
	podLogOpts := &v1.PodLogOptions{}
	req := obj.Client.Get().Namespace(obj.Namespace).
		Name(obj.Name).
		Resource(obj.ResourceMapping().Resource.Resource).
		SubResource("log").
		VersionedParams(podLogOpts, scheme.ParameterCodec)
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return err.Error()
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return err.Error()
	}
	str := buf.String()

	return str
}

// waitForStatusByInformer 使用informer监听等待pod状态
// TODO: 可考虑加入超时机制，传入ctx，做超时
// FIXME: 当已经有此资源时，会阻塞在此
func waitForStatusByInformer(ctx context.Context, obj *resource.Info, objType string) error {
	var err error
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	// 如果是service configmap secret资源对象，直接返回
	if objType == "service" || objType == "configmap" || objType == "secret" {
		return nil
	}

	lw := cache.NewListWatchFromClient(obj.Client, obj.ResourceMapping().Resource.Resource, obj.Namespace, fields.Everything())
	informer := cache.NewSharedIndexInformer(lw, obj.Object, 0, nil)
	ch := make(chan struct{})
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			var ot runtime.Object
			if objType == "pod" {
				ot = &v1.Pod{}
			} else if objType == "deployment" {
				ot = &appv1.Deployment{}
			}

			err := runtime.DefaultUnstructuredConverter.FromUnstructured(newObj.(*unstructured.Unstructured).UnstructuredContent(), ot)
			if err != nil {
				klog.Errorf("object informer error")
				close(ch)
				return
			}

			// pod 直到状态为running时关闭，deployment 则是当期望副本与实际副本相等时关闭
			switch objType {
			case "pod":
				pod := ot.(*v1.Pod)
				// 当此pod是running时，关闭informer
				if pod.Status.Phase == v1.PodRunning {
					klog.Infof("pod name [%s] namespace [%s], success", pod.Name, pod.Namespace)
					klog.Infof("pod container log: %v", getPodLogs(obj))
					close(ch)
				}
			case "deployment":
				dep := ot.(*appv1.Deployment)
				if dep.Status.Replicas == dep.Status.ReadyReplicas {
					klog.Infof("deployment name [%s] namespace [%s], success", dep.Name, dep.Namespace)
					close(ch)
				}
			}

		},
	})
	// 使用 context.WithTimeout 创建带有超时的上下文
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	// 阻塞运行，直到informer停止或超时
	go informer.Run(ch)
	select {
	case <-ctx.Done():
		return fmt.Errorf("waitForStatusByInformer err: %v, "+
			"maybe this resource [%v] [%v/%v] has not been modified or deployment failed\n",
			ctx.Err(), obj.Object.GetObjectKind().GroupVersionKind().Kind, obj.Name, obj.Namespace)
	case <-ch:
		return err
	}
}

// workflowHandler 执行工作流
// FIXME: 拆分方法，把不同功能分出去
func workflowHandler(v cue.Value) (flow.Runner, error) {
	l, b := v.Label()
	// 如果是根节点，跳过
	if !b || l == PodFlowRoot {
		return nil, nil
	}
	return flow.RunnerFunc(func(t *flow.Task) error {
		fmt.Printf("----------------------%s-------------------------------\n", t.Path())
		klog.Infof("current workflow index: %v", t.Index())
		for _, d := range t.Dependencies() {
			klog.Infof("current dependency index: %v", d.Path())
		}
		if t.Index() != 0 {
			action := getField(t.Value(), "action", "apply")
			taskType := getField(t.Value(), "type", "k8s")

			// 执行k8s流程
			if taskType == "k8s" {
				objType := getField(t.Value(), "objType", "pod")
				podJson, err := jsonField(t.Value(), "template")
				if err != nil {
					return err
				}
				// 区分两种动作 apply delete
				if action == "apply" {
					res, err := apply(podJson)
					if err != nil {
						return err
					}

					// TODO: 如果需要支持其他资源对象，需要修改informer
					// 如果返回的pod对象有多个，则调用waitForStatusByInformer
					if len(res) > 0 {
						err = waitForStatusByInformer(context.Background(), res[0], objType)
						if err != nil {
							klog.Error("waitForStatusByInformer error: ", err)
						}
					}
				} else {
					err = delete(podJson)
					if err != nil {
						return err
					}
				}
			}

			// 执行脚本流程
			if taskType == "bash" {

				scriptToRun := getField(t.Value(), "script", "")
				// 创建命令对象
				cmd := exec.Command("bash", "-s")
				cmd.Stdin = strings.NewReader(scriptToRun)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				// 执行脚本
				err := cmd.Start()
				if err != nil {
					klog.Errorf("launching script error: %v", err)
					return err
				}

				// 等待脚本执行完成
				err = cmd.Wait()
				if err != nil {
					klog.Errorf("executing script error: %v", err)
					return err
				}
			}
		}

		return nil
	}), nil
}

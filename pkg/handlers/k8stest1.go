package handlers

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/tools/flow"
	"fmt"
	"github.com/practice/workflow-practice/pkg/utils"
)

const K8sTest1Root = "workflow" //代表 根节点

// 工作流启动函数
func K8stest1Handler(v cue.Value) (flow.Runner, error) {
	l, b := v.Label()
	if !b || l == K8sTest1Root {
		return nil, nil
	}
	return flow.RunnerFunc(func(t *flow.Task) error {
		fmt.Println("工作流节点", t.Path())

		k8sJson, err := t.Value().MarshalJSON()
		if err != nil {
			return err
		}
		err = utils.K8sApply(k8sJson, utils.K8sRestConfig, *utils.K8sRestMapper)
		if err != nil {
			return err
		}
		return nil
	}), nil
}

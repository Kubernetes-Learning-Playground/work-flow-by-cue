package handlers

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/tools/flow"
	"fmt"
	"github.com/practice/workflow-practice/pkg/utils"
	"os"
	"os/exec"
	"strings"
)

const K8sTest1Root = "workflow" //代表 根节点

// 工作流启动函数
func K8stest1Handler(v cue.Value) (flow.Runner, error) {
	l, b := v.Label()
	//fmt.Println(l)
	if !b || l == K8sTest1Root {
		return nil, nil
	}
	return flow.RunnerFunc(func(t *flow.Task) error {
		fmt.Println("工作流节点", t.Path())
		//fmt.Println(t.Value())
		fmt.Println(t.Value().LookupPath(cue.ParsePath("type")))
		// TODO: 这里可以区分，使用bash 脚本还是部署k8s服务
		tt, _ := t.Value().LookupPath(cue.ParsePath("type")).String()
		if tt == "k8s" {
			action := getField(t.Value(), "action", "apply")
			k8sJson, err := t.Value().LookupPath(cue.ParsePath("component")).MarshalJSON()
			if err != nil {
				return err
			}
			fmt.Println(string(k8sJson))
			if action == "apply" {
				_, err = utils.K8sApply(k8sJson, utils.K8sRestConfig, *utils.K8sRestMapper)
				if err != nil {
					return err
				}

			} else {
				err = utils.K8sDelete(k8sJson, utils.K8sRestConfig, *utils.K8sRestMapper)
				if err != nil {
					return err
				}
			}
		} else if tt == "bash" {
			result1, _ := t.Value().LookupPath(cue.ParsePath("script")).String()
			////// 创建命令对象
			cmd := exec.Command("bash", "-s")
			cmd.Stdin = strings.NewReader(result1)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			// 执行脚本
			err := cmd.Start()
			if err != nil {
				fmt.Println("启动脚本时出错:", err)
				return err
			}

			// 等待脚本执行完成
			err = cmd.Wait()
			if err != nil {
				fmt.Println("执行脚本时出错:", err)
				return err
			}
		}

		return nil
	}), nil
}

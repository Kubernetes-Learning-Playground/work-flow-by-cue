package handlers

import (
	"context"
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/tools/flow"
	"fmt"
	"strings"
	"sync"
)

// 这里放一个全局变量 用来收集 WorkFlow
type WorkFlow struct {
	Name    string
	Desc    string
	NewFlow func(string) *flow.Controller
	flow    *flow.Controller
	//上节 是 Flow  *flow.Controller
	Status    string
	Successed int
	Failed    int
	Params    string //提交的参数   可以有也可以没有
	lock      sync.Mutex
	Order     int //用于排序
}

// 重置
func (wf *WorkFlow) Reset() {
	wf.lock.Lock()
	defer wf.lock.Unlock()

	wf.Successed = 0
	wf.Failed = 0
	wf.Status = ""
	wf.flow = nil
	wf.Status = ""
	wf.Params = "" // 参数要清掉
}

// 专门封了一个函数
func (wf *WorkFlow) Run(ctx context.Context) error {
	if wf.lock.TryLock() {
		defer wf.lock.Unlock()
	}
	wf.flow = nil // 很关键
	return wf.GetFlow().Run(ctx)
}

// 获取当前工作流
func (wf *WorkFlow) GetFlow() *flow.Controller {
	if wf.lock.TryLock() {
		defer wf.lock.Unlock()
	}

	if wf.flow == nil { // 如果flow=nil 重新初始化
		wf.flow = wf.NewFlow(wf.Params)
	}
	return wf.flow
}

var WorkFlows = make(map[string]*WorkFlow)

// 排序工作流  ---还没租
func SortFlows() []*WorkFlow {
	//flows := []*WorkFlow{}
	//for _, v := range WorkFlows {
	//
	//}
	return nil
}

// 注册
func Register(name, desc string, newFlow func(params string) *flow.Controller, order int) {
	WorkFlows[name] = &WorkFlow{
		Name: name, Desc: desc, NewFlow: newFlow, Order: order,
	}
}

// NewFlowFunc 创建工作流方法
// tplPath模板路径 root工作流根节点
func NewFlowFunc(tplPath, root string, taskFunc flow.TaskFunc) func(params string) *flow.Controller {
	return func(params string) *flow.Controller {
		inst := load.Instances([]string{tplPath}, nil)[0]
		cc := cuecontext.New()
		cv := cc.BuildInstance(inst)
		if cv.Err() != nil {
			panic(cv.Err().Error())
		}

		// filledCv 是覆盖后的结果
		filledCv := cv

		// 解析并填充cue模版，判断params是否有值
		if strings.Trim(params, " ") != "" {
			pv := cc.CompileString(params)
			if pv.Err() != nil {
				panic(pv.Err().Error())
			}

			if fields, err := pv.Fields(); err == nil {
				for fields.Next() {
					fillPath := cue.ParsePath(root + "." + fields.Label()) // workflow.params

					if !cv.LookupPath(fillPath).Exists() {
						continue
					}
					// 填充
					filledCv = filledCv.FillPath(fillPath, fields.Value())
					if filledCv.Err() != nil {
						panic(filledCv.Err())
					}

				}
			}
		}

		return flow.New(&flow.Config{Root: cue.ParsePath(root)}, filledCv, taskFunc)
	}
}

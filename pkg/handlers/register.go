package handlers

import (
	"context"
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/tools/flow"
	"strings"
	"sync"
)

// WorkFlow 工作流对象
type WorkFlow struct {
	// 工作流名
	Name string
	// 描述
	Desc string
	// NewFlow 初始化工作流func
	NewFlow func(string) *flow.Controller
	// flow对象
	flow *flow.Controller
	// Status 工作流状态
	Status string
	// Successful 成功次数
	Successful int
	// Failed 失败次数
	Failed int
	// Params 传入参数
	Params string

	lock sync.Mutex
}

// Reset 重置工作流对象
func (wf *WorkFlow) Reset() {
	wf.lock.Lock()
	defer wf.lock.Unlock()

	// 清理字段
	wf.Successful = 0
	wf.Failed = 0
	wf.Status = ""
	wf.flow = nil
	wf.Status = ""
	wf.Params = ""
}

// Run 先获取该工作流，并执行Run方法
func (wf *WorkFlow) Run(ctx context.Context) error {
	if wf.lock.TryLock() {
		defer wf.lock.Unlock()
	}
	// 需要把原来的flow至为空
	wf.flow = nil
	return wf.GetFlow().Run(ctx)
}

// GetFlow 获取当前工作流
func (wf *WorkFlow) GetFlow() *flow.Controller {
	if wf.lock.TryLock() {
		defer wf.lock.Unlock()
	}
	// 如果flow=nil 重新初始化
	if wf.flow == nil {
		wf.flow = wf.NewFlow(wf.Params)
	}
	return wf.flow
}

// WorkFlows 可存放多个工作流对象，
// 接口执行时，使用flow的name查找对应的工作流
var WorkFlows = make(map[string]*WorkFlow)

// Register 注册工作流
func Register(name, desc string, newFlow func(params string) *flow.Controller) {
	WorkFlows[name] = &WorkFlow{Name: name, Desc: desc, NewFlow: newFlow}
}

// NewFlowFunc 创建工作流方法
// tplPath cue模板路径 root工作流根节点
// FIXME: 不要全都是 panic
func NewFlowFunc(tplPath, root string) func(params string) *flow.Controller {
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
					// workflow.params
					fillPath := cue.ParsePath(root + "." + fields.Label())

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

		// 执行工作流的流程模版
		return flow.New(&flow.Config{Root: cue.ParsePath(root)}, filledCv, workflowHandler)
	}
}

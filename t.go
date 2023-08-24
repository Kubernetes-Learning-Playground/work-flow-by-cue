package main

import (
	"context"
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/tools/flow"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

// 工作流对象
const tasks = `
    reg: {
		uname: "admin"
		pass: "12345"
	}
	regrsp: {
		regname: reg.uname
		result: string
	}
	test:{
		data: int
	}
     `

func main() {
	cc := cuecontext.New()
	cv := cc.CompileString(tasks)
	// cue 工作流对象
	regFlow := flow.New(nil, cv, regFlowFunc)
	flowCtx, flowCancel := context.WithCancel(context.Background())

	r := gin.New()

	r.LoadHTMLGlob("test/workflow/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{"tasks": regFlow.Tasks()})
	})

	// 重置工作流
	r.GET("/reset", func(c *gin.Context) {
		regFlow = flow.New(nil, cv, regFlowFunc)
		c.Redirect(302, "/")
	})

	// 删除工作流
	r.GET("/cancel", func(c *gin.Context) {
		flowCancel()
		c.Redirect(302, "/")
	})

	// 执行工作流
	r.POST("/run", func(c *gin.Context) {
		go func() {
			if err := regFlow.Run(flowCtx); err != nil {
				log.Println(err)
			}
		}()
		c.Redirect(302, "/")
	})

	r.Run(":8088")
}

func regFlowFunc(v cue.Value) (flow.Runner, error) {
	uname := v.LookupPath(cue.ParsePath("uname"))
	result := v.LookupPath(cue.ParsePath("result"))
	if !uname.Exists() && !result.Exists() {
		return nil, nil
	}

	return flow.RunnerFunc(func(t *flow.Task) error {
		// 如果工作流的程流是uname存在时的操作
		if t.Path().String() == "reg" {
			//假设模拟 数据库很耗时
			time.Sleep(time.Second * 5)
			unameStr, err := uname.String()
			if err != nil {
				return err
			}
			if unameStr == "admin" {
				return fmt.Errorf("不能注册为admin用户名")
			}
			return nil
		}

		lastUname, err := t.Value().LookupPath(cue.ParsePath("regname")).String()
		if err != nil {
			return err
		}
		return t.Fill(map[string]interface{}{
			"result": lastUname + "用户注册成功",
		})

	}), nil
}

package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/practice/workflow-practice/pkg/common"
	"github.com/practice/workflow-practice/pkg/handlers"
	"k8s.io/klog/v2"
)

func HttpServer(c *common.ServerConfig) {

	if !c.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(gin.Recovery())

	// TODO: 接口返回时，暴露更多信息

	// 重置工作流接口
	r.POST("/reset/:name", func(c *gin.Context) {
		name := c.Param("name") //获取工作流名称
		if flow, ok := handlers.WorkFlows[name]; ok {
			flow.Reset()
			c.JSON(200, gin.H{"ok": "ok"})
			return
		}
		c.JSON(400, gin.H{"error": "not found workflow " + name})
		return
	})

	// 启动工作流接口
	// ex: http://localhost:8085/start/workflow
	r.POST("/start/:name", func(c *gin.Context) {
		name := c.Param("name")        // 获取工作流名称
		params := c.PostForm("params") // 获取提交参数
		if flow, ok := handlers.WorkFlows[name]; ok {
			flow.Params = params
			if err := flow.Run(context.Background()); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"ok": "ok"})
		} else {
			c.JSON(400, gin.H{"error": "not found workflow " + name})
			return
		}
	})

	// 注册模版接口
	r.POST("/register", func(c *gin.Context) {
		var rr *RegisterRequest
		if err := c.ShouldBindJSON(&rr); err != nil {
			klog.Error("bind json err!")
			c.JSON(400, gin.H{"error": err})
			return
		}
		handlers.Register(rr.WorkFlowName, rr.WorkFlowDesc, handlers.NewFlowFunc(rr.TemplatePath, handlers.PodFlowRoot))
		c.JSON(200, gin.H{"ok": "ok"})
	})

	err := r.Run(fmt.Sprintf(":%v", c.Port))
	fmt.Println(err)
}

type RegisterRequest struct {
	// WorkFlowName 工作流名
	WorkFlowName string `json:"workFlowName"`
	// WorkFlowDesc 工作流描述
	WorkFlowDesc string `json:"workFlowDesc"`
	// TemplatePath 模版路径，可以绝对路径，也可以在项目"根目录"的相对路径
	TemplatePath string `json:"templatePath"`
}

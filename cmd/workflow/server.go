package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/practice/workflow-practice/pkg/handlers"
)

func errorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if e := recover(); e != nil {
				c.HTML(200, "error.html", gin.H{"message": e})
			}
		}()
		c.Next()
	}
}

func main() {
	r := gin.New()
	r.Use(errorHandler())
	r.LoadHTMLGlob("test/workflow/*")

	// 首页
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{"flows": handlers.WorkFlows})
	})

	// 重置工作流接口
	r.POST("/reset/:name", func(c *gin.Context) {
		name := c.Param("name") //获取工作流名称
		if flow, ok := handlers.WorkFlows[name]; ok {
			flow.Reset()
		}
		c.Redirect(302, "/")
	})

	// 启动工作流接口
	r.POST("/start/:name", func(c *gin.Context) {

		name := c.Param("name")        //获取工作流名称
		params := c.PostForm("params") // 获取 提交参数 ---也可能没有  不一定有
		if flow, ok := handlers.WorkFlows[name]; ok {
			flow.Params = params

			if err := flow.Run(context.TODO()); err != nil {
				fmt.Println("err: ", err)
				flow.Status = err.Error() //设置当前状态
				flow.Failed++
			} else {
				flow.Status = "success"
				flow.Successful++
			}
			c.Redirect(302, "/")

		} else {
			panic("workflow not found")
		}

	})

	r.Run(":8085")
}

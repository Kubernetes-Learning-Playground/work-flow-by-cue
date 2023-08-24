package flowtpls
import (
   "github.com/workflow/yamls"
)
workflow: {
   step1: yamls.task_svc
   step3: yamls.task_bash
   step2: yamls.task_deploy
}
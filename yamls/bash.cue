package yamls

task_bash: {
 type: "bash",
 script: """
  #!/bin/bash

  for i in {1..10}; do
    echo $i
    sleep 1
  done
  """
}
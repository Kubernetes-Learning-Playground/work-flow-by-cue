package pods

// 脚本内容
script: """
#!/bin/bash

for i in {1..10}; do
echo $i
sleep 1
done
"""
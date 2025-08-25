# DSVS
data share system


## cmd:
./bin/DrynxServerRun.sh  
go run ./ginsrv/.

v2:
chmod +x ./bin/drynxnoderun.sh
./bin/drynxnoderun.sh


## SqL eg.
SELECT SUM(visits) FROM statistics_experiment_data WHERE age > 45

SELECT MAX(visits) FROM statistics_experiment_data WHERE department = 'Pediatrics'

sudo mysql -u root -p

## fixed
1. 验证节点为一次性  已经解决:每次创建新的链
2. 数据大小不能超过10000  go/pkg/mod/github.com/ldsec/unlynx/lib/ crypto.go： 修改const MaxHomomorphicInt int64 = 10_000_000，但是速度及其的慢，尝试使用cuttingFactors
3. 目前没有连接数据库       已经使用了mysql并新增了dp启动方式
4. 目前部门IP为手动设置     已经解决 by import cropID
5. 创建一个数据库表来存DP的信息，最少有公钥、地址,ID           完成


## todo
1. sql只返回一个值而不再是一列
2. 小数以及cuttingfactors
3. 真机部署

### need fix:
drynxclient:  需要提前知道表的数据的最值
---
loadServerInfo: 需要研究清楚这个端口号的意义

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

## 需要解决的问题
1. 验证节点为一次性  已经解决:每次创建新的链
2. 数据大小不能超过10000  go/pkg/mod/github.com/ldsec/unlynx/lib/ crypto.go： 修改const MaxHomomorphicInt int64 = 10_000_000，但是速度及其的慢，尝试使用cuttingFactors
3. 数据只能是整数          plan使用定点数试试
4. 目前没有连接数据库       已经使用了mysql并新增了dp启动方式
5. 目前部门IP为手动设置     已经解决 by import cropID
6. 没有部署在真机上

下一步计划另DP自己发送自己的IP地址和公钥，实际生产环境，是没办法获取到DP的IP的
另sql只返回一个值而不再是一列
还有就是解决小数以及cuttingfactors
真机部署

need fix:
drynxclient
loadServerInfo.go
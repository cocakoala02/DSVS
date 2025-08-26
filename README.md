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
1. 验证节点为一次性                                     已经解决:每次创建新的链
2. 数据大小不能超过10000                                go/pkg/mod/github.com/ldsec/unlynx/lib/ crypto.go： 修改const MaxHomomorphicInt int64 = 10_000_000，但是速度及其的慢，尝试使用cuttingFactors
3. 目前没有连接数据库                                    已经使用了mysql并新增了dp启动方式
4. 目前部门IP为手动设置                                  已经解决 by import cropID
5. 创建一个数据库表来存DP的信息，最少有公钥、地址,ID         已经完成
INSERT INTO dp_info_table (ID, Addr, Pub) VALUES
('A','127.0.0.1:7013','303414b7221c34f41a028657cb5f5ef025f3a875673735d568586e4458d7bbae20b75d911b5379f1fbedf2cf57437377522c58f489f13bba1ccf78346d48b870'),
('B','127.0.0.1:7015','5eec4664ccb37b460c3841aaf470a768d345418a84e88644abcd7b5fe900214b01d8e6ae17e038deced6862a42f5d05588767cbb58a1e418e030e8edb82860a4'),
('C','127.0.0.1:7017','3b0a9ea5dfb72d8bcd9e600c2bd82da604482c68832c4451fd260082c05d78ef7d1a804edcd3ca92a07b1d8c8c7baea03bb01b7adf00fde94959c864bd0918f9'),
('D','127.0.0.1:7019','47b4e6d3e51cacbb081b9900047d912bc7ce34195e8d2efdef6d9d1465942d2a83b7b3fb63fe132695f336439062ce576943451748eb3bdf899167f9a2c5a023'),
('E','127.0.0.1:7021','237ad8023d955ac0543e261139634cf6ed7bb8f90c8bca05d4f2bb6f087dadd93d56f8177e9220691a621e21c16c5d42be18499555649799255f0ab9c291f1aa'),
('F','127.0.0.1:7023','604fbf1f8ccc2af2edcc91ee75c0a603ff364f6103fd7b2b3c4fe8dae6668b2d370ad06a25184c2778bc45368ce2581b986f6644e09b43a693dea6472faa7457'),
('G','127.0.0.1:7025','0a104c13677fe5684e1dcf5b212eb3817caef57dbf7492165dd1acb91814a82888200fde0d89392687d5cda6165e7ed72f74c83070b60a4a6f55f9188c728825'),
('H','127.0.0.1:7027','2ebf5199e233dbef3eaa4e0d1e01c8b67c941dae1b10e3315973c8c3e67a32f20be64f474e9ecd36a23294d2bc0f8c31e9cfee7dd7b28344c2f34b8af6c46de5'),
('I','127.0.0.1:7029','0a93c70a7cf1a61f415d564ecef4839ab2b1711531805892fe7ece6f68d2ac56720ba50234d99ee729d914cf82f81d3d71bc57dc2f3373ee7dc90d8a88b80b96');
ON DUPLICATE KEY UPDATE
  Addr = VALUES(Addr),
  Pub  = VALUES(Pub);

6. sql只返回一个值而不再是一列                            已经完成，对于sum、avg、count仅仅返回值  

## todo
1. cuttingfactors
2. 小数
3. 真机部署

### need fix:
---
    drynxclient:  需要提前知道表的数据的最值
---
    loadServerInfo:   需要研究清楚这个端口号的意义

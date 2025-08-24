## DP custom run method 为了DP能嵌入数据库信息

cmd:
go build -o DrynxDP ../cmd/dp

./bin/DrynxDP gen 127.0.0.1:7013 127.0.0.1:7014 > ./bin/dpA.toml
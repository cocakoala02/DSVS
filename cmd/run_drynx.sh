#!/usr/bin/env bash

# 出现错误就终止脚本
set -e

go build -o drynx-server ./server
go build -o drynx-client ./client

echo "=== 1) 生成节点配置文件 ==="
./drynx-server gen 127.0.0.1:7001 127.0.0.1:7002 > node1_config.toml
./drynx-server gen 127.0.0.1:7003 127.0.0.1:7004 > node2_config.toml
./drynx-server gen 127.0.0.1:7005 127.0.0.1:7006 > node3_config.toml
./drynx-server gen 127.0.0.1:7007 127.0.0.1:7008 > node4_config.toml
./drynx-server gen 127.0.0.1:7009 127.0.0.1:7010 > node5_config.toml
./drynx-server gen 127.0.0.1:7011 127.0.0.1:7012 > node6_config.toml
./drynx-server gen 127.0.0.1:7013 127.0.0.1:7014 > node7_config.toml
./drynx-server gen 127.0.0.1:7015 127.0.0.1:7016 > node8_config.toml
./drynx-server gen 127.0.0.1:7017 127.0.0.1:7018 > node9_config.toml
./drynx-server gen 127.0.0.1:7019 127.0.0.1:7020 > node10_config.toml
./drynx-server gen 127.0.0.1:7021 127.0.0.1:7022 > node11_config.toml
./drynx-server gen 127.0.0.1:7023 127.0.0.1:7024 > node12_config.toml
./drynx-server gen 127.0.0.1:7025 127.0.0.1:7026 > node13_config.toml
./drynx-server gen 127.0.0.1:7027 127.0.0.1:7028 > node14_config.toml
./drynx-server gen 127.0.0.1:7029 127.0.0.1:7030 > node15_config.toml




# 休眠几秒，视系统性能可调整
sleep 3

echo -e "\\n=== 2) 分别启动 9 个节点(在新 xterm 窗口) ==="

# 使用 xterm 打开窗口，每个节点运行: cat nodeX_config.toml | ./drynx-server.exe run
# 并且 -hold 保留窗口，防止干脆退出
gnome-terminal -- bash -c "cat node1_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node2_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node3_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node4_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node5_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node6_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node7_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node8_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node9_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node10_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node11_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node12_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node13_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node14_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&
gnome-terminal -- bash -c "cat node15_config.toml | ./drynx-server run; read -p 'Press Enter to close...'"&





echo -e "\\n=== 等待节点启动... ==="
sleep 3

echo -e "\\n=== 3) 提取节点公钥 ==="

# 封装一个函数: 从 toml 文件里匹配形如 `Public = "..."`
get_public_key_from_toml() {
    local configFile="$1"
    # grep 匹配第一行, 然后删除 '=' 左侧，去除双引号和多余空格
    grep '^Public\s*=' "$configFile" \
      | head -n1 \
      | sed 's/.*=//;s/"//g;s/^[[:space:]]*//;s/[[:space:]]*$//'
}

pub1=$(get_public_key_from_toml node1_config.toml)
pub2=$(get_public_key_from_toml node2_config.toml)
pub3=$(get_public_key_from_toml node3_config.toml)
pub4=$(get_public_key_from_toml node4_config.toml)
pub5=$(get_public_key_from_toml node5_config.toml)
pub6=$(get_public_key_from_toml node6_config.toml)
pub7=$(get_public_key_from_toml node7_config.toml)
pub8=$(get_public_key_from_toml node8_config.toml)
pub9=$(get_public_key_from_toml node9_config.toml)
pub10=$(get_public_key_from_toml node10_config.toml)
pub11=$(get_public_key_from_toml node11_config.toml)
pub12=$(get_public_key_from_toml node12_config.toml)
pub13=$(get_public_key_from_toml node13_config.toml)
pub14=$(get_public_key_from_toml node14_config.toml)
pub15=$(get_public_key_from_toml node15_config.toml)





# 也可选地 echo 出来调试
# echo "Node1 公钥: $pub1"
# echo "Node2 公钥: $pub2"
# echo "Node3 公钥: $pub3"
# echo "Node4 公钥: $pub4"
# echo "Node5 公钥: $pub5"
# echo "Node6 公钥: $pub6"
# echo "Node7 公钥: $pub7"
# echo "Node8 公钥: $pub8"
# echo "Node9 公钥: $pub9"

echo -e "\\n=== 4) 创建 network.toml ==="
./drynx-client network new |
    ./drynx-client network add-node 127.0.0.1:7001 "$pub1" |
    ./drynx-client network add-node 127.0.0.1:7003 "$pub2" |
    ./drynx-client network add-node 127.0.0.1:7005 "$pub3" |
    ./drynx-client network add-node 127.0.0.1:7007 "$pub4" |
    ./drynx-client network add-node 127.0.0.1:7009 "$pub5" |
    ./drynx-client network add-node 127.0.0.1:7011 "$pub6" |
    ./drynx-client network add-node 127.0.0.1:7013 "$pub7" |
    ./drynx-client network add-node 127.0.0.1:7015 "$pub8" |
    ./drynx-client network add-node 127.0.0.1:7017 "$pub9" |
    ./drynx-client network add-node 127.0.0.1:7019 "$pub10" |
    ./drynx-client network add-node 127.0.0.1:7021 "$pub11" |
    ./drynx-client network add-node 127.0.0.1:7023 "$pub12" |
    ./drynx-client network add-node 127.0.0.1:7025 "$pub13" |
    ./drynx-client network add-node 127.0.0.1:7027 "$pub14" |
    ./drynx-client network add-node 127.0.0.1:7029 "$pub15" |
    ./drynx-client network set-client 127.0.0.1:7002 \
    > network.toml

echo -e "\\n=== 5) 创建 survey.toml ==="
./drynx-client survey new test |
    ./drynx-client survey set-operation sum \
    > survey.toml

echo -e "\\n=== 6) 执行查询 ==="
cat network.toml survey.toml | ./drynx-client survey run

echo -e "\\n=== 全部操作结束 ==="

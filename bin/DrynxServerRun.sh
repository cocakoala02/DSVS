#!/bin/bash

file_path="./bin/DrynxServer"

if [ -f "$file_path" ]; then
    rm "$file_path"
fi

go build -o ./bin/DrynxServer  ./cmd/server/main.go

node_path="./bin/node1_config.toml"

if [ ! -f "$node_path" ]; then
    echo "reset"
    ./bin/DrynxServer gen 127.0.0.1:7001 127.0.0.1:7002 > ./bin/node1_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7003 127.0.0.1:7004 > ./bin/node2_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7005 127.0.0.1:7006 > ./bin/node3_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7007 127.0.0.1:7008 > ./bin/node4_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7009 127.0.0.1:7010 > ./bin/node5_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7011 127.0.0.1:7012 > ./bin/node6_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7013 127.0.0.1:7014 > ./bin/node7_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7015 127.0.0.1:7016 > ./bin/node8_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7017 127.0.0.1:7018 > ./bin/node9_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7019 127.0.0.1:7020 > ./bin/node10_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7021 127.0.0.1:7022 > ./bin/node11_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7023 127.0.0.1:7024 > ./bin/node12_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7025 127.0.0.1:7026 > ./bin/node13_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7027 127.0.0.1:7028 > ./bin/node14_config.toml
    ./bin/DrynxServer gen 127.0.0.1:7029 127.0.0.1:7030 > ./bin/node15_config.toml

    
    
fi

gnome-terminal -- bash -c "cat ./bin/node1_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node2_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node3_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node4_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node5_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node6_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node7_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node8_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node9_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node10_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node11_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node12_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node13_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node14_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"
gnome-terminal -- bash -c "cat ./bin/node15_config.toml | ./bin/DrynxServer run; echo 'Press any key to exit...'; read -n 1; exec bash"


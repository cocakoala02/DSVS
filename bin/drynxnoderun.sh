#!/usr/bin/env bash
set -euo pipefail

BIN_DIR="./bin"
SERVER_BIN="${BIN_DIR}/DrynxServer"
DP_BIN="${BIN_DIR}/DrynxDP"

# ========== 构建 ==========
build_bins() {
  echo "[build] compiling binaries..."
  rm -f "${SERVER_BIN}" "${DP_BIN}" || true
  go build -o "${SERVER_BIN}" ./cmd/server/main.go
  go build -o "${DP_BIN}"     ./cmd/dp/main.go
}

# # ========== 清理旧配置 ==========
# clean_old_configs() {
#   echo "[clean] removing old config toml files..."
#   rm -f ${BIN_DIR}/node*_config.toml ${BIN_DIR}/dp*_config.toml || true
# }

# ========== 生成配置 ==========
# gen_cn() {
#   # CN: node1..node3  7001/7002, 7003/7004, 7005/7006
#   echo "[gen] generating CN (DrynxServer) configs..."
#   "${SERVER_BIN}" gen 127.0.0.1:7001 127.0.0.1:7002 > "${BIN_DIR}/cn1_config.toml"
#   "${SERVER_BIN}" gen 127.0.0.1:7003 127.0.0.1:7004 > "${BIN_DIR}/cn2_config.toml"
#   "${SERVER_BIN}" gen 127.0.0.1:7005 127.0.0.1:7006 > "${BIN_DIR}/cn3_config.toml"
# }

# gen_vn() {
#   # VN: node4..node6  7007/7008, 7009/7010, 7011/7012
#   echo "[gen] generating VN (DrynxServer) configs..."
#   "${SERVER_BIN}" gen 127.0.0.1:7007 127.0.0.1:7008 > "${BIN_DIR}/vn1_config.toml"
#   "${SERVER_BIN}" gen 127.0.0.1:7009 127.0.0.1:7010 > "${BIN_DIR}/vn2_config.toml"
#   "${SERVER_BIN}" gen 127.0.0.1:7011 127.0.0.1:7012 > "${BIN_DIR}/vn3_config.toml"
# }


# gen_dp() {
#   # DP: dp7..dp15  7013/7014 .. 7029/7030
#   echo "[gen] generating DP (DrynxDP) configs..."
#   local n=7 node_port=7013 client_port=7014
#   while [ $n -le 15 ]; do
#     "${DP_BIN}" gen "127.0.0.1:${node_port}" "127.0.0.1:${client_port}" > "${BIN_DIR}/dp${n}_config.toml"
#     echo " -> ${BIN_DIR}/dp${n}_config.toml"
#     n=$((n+1)); node_port=$((node_port+2)); client_port=$((client_port+2))
#   done

#   # 如需覆盖 [database] 段，可在此 sed 批量替换：
#   # for f in ${BIN_DIR}/dp{7..15}_config.toml; do
#   #   sed -i -e 's/^driver =.*/driver = "mysql"/' \
#   #          -e 's|^dsn =.*|dsn = "user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true&timeout=2s&readTimeout=2s"|' "$f"
#   # done
# }

# ========== 运行 ==========
run_cn() {
  echo "[run] starting CN nodes..."
  for i in 1 2 3; do
    cfg="${BIN_DIR}/cn${i}_config.toml"
    gnome-terminal -- bash -c "cat '${cfg}' | '${SERVER_BIN}' run; echo 'CN ${i} exited. Press any key to close...'; read -n 1; exec bash" &
  done
}

run_vn() {
  echo "[run] starting VN nodes..."
  for i in 1 2 3; do
    cfg="${BIN_DIR}/vn${i}_config.toml"
    gnome-terminal -- bash -c "cat '${cfg}' | '${SERVER_BIN}' run; echo 'VN ${i} exited. Press any key to close...'; read -n 1; exec bash" &
  done
}

run_dp() {
  echo "[run] starting DP nodes..."
  for i in $(seq 7 15); do
    local cfg="${BIN_DIR}/dp${i}_config.toml"
    gnome-terminal -- bash -c "cat '${cfg}' | '${DP_BIN}' run; echo 'DP node ${i} exited. Press any key to close...'; read -n 1; exec bash" &
  done
}

# ========== 主流程 ==========
build_bins
# clean_old_configs
# gen_cn
# gen_vn
# gen_dp
run_cn
run_vn
run_dp

echo "[done] launched 3 CN + 3 VN + 9 DP nodes."

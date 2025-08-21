# a basic version without VNs
1. The usage method of it
   1. delete all of the .toml first
   2. keep the work path in the .../drynx/
   3. and then run ./bin/DrynxServerRun.sh in bash
   4. then run the gin-server,just:
        ```bash
        go run ./ginsrv/.
        ```
   5. it will auto load the config and set the network
   6. if you let the config.Ranges be 0, it will do not run the proof process
   7. and it can only work in this case
2. why can not run whithin VNs
   1. ERROR:send 1195725856 bytes packet to the VN and it is bigger than the limit, and then the vn close or reset the connection
   2. Problems:why just send to the 7022, which is the second vns in the roster
   3. maybe the 1195725856 has its special meaning, because it is impossible for each node to send such a big packet

# TODO
1. finish the search function in DPs
2. add a function to make the CNtoDPs map
3. adapt the nums of the cns and dps
4. data struct:
   1. tablename
   2. sql
   3. dps to path
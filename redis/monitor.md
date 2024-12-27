# run docker compose
```sh 
docker-compose up
```

# create the cluster

```sh 
docker exec -it redis1 \
redis-cli --cluster create \
redis1:7000 redis2:7001 redis3:7002 \
--cluster-replicas 0
```


```>>> Performing hash slots allocation on 3 nodes...
Master[0] -> Slots 0 - 5460
Master[1] -> Slots 5461 - 10922
Master[2] -> Slots 10923 - 16383
M: ae739172a0539039ec31763d2bf22ee4b8368bfd redis1:7000
   slots:[0-5460] (5461 slots) master
M: 7371a2f8537392aa13cfdbb51f05364bc610d1b6 redis2:7001
   slots:[5461-10922] (5462 slots) master
M: 6fb30c9de6628d69264b60502d06118c18e2425b redis3:7002
   slots:[10923-16383] (5461 slots) master
Can I set the above configuration? (type 'yes' to accept):

>>> Nodes configuration updated
>>> Assign a different config epoch to each node
>>> Sending CLUSTER MEET messages to join the cluster
Waiting for the cluster to join
...
>>> Performing Cluster Check (using node redis1:7000)
M: ae739172a0539039ec31763d2bf22ee4b8368bfd redis1:7000
   slots:[0-5460] (5461 slots) master
M: 7371a2f8537392aa13cfdbb51f05364bc610d1b6 172.19.0.2:7001
   slots:[5461-10922] (5462 slots) master
M: 6fb30c9de6628d69264b60502d06118c18e2425b 172.19.0.3:7002
   slots:[10923-16383] (5461 slots) master
[OK] All nodes agree about slots configuration.
>>> Check for open slots...
>>> Check slots coverage...
[OK] All 16384 slots covered.

```

# verify the cluster is working 

```sh
docker exec -it redis1 redis-cli -p 7000 cluster info
```

```
cluster_state:ok
cluster_slots_assigned:16384
cluster_slots_ok:16384
cluster_slots_pfail:0
cluster_slots_fail:0
cluster_known_nodes:3
cluster_size:3
cluster_current_epoch:3
cluster_my_epoch:1
cluster_stats_messages_ping_sent:71
cluster_stats_messages_pong_sent:76
cluster_stats_messages_sent:147
cluster_stats_messages_ping_received:74
cluster_stats_messages_pong_received:71
cluster_stats_messages_meet_received:2
cluster_stats_messages_received:147
total_cluster_links_buffer_limit_exceeded:0
```

# write some sample data
```sh
docker exec -it redis1 /bin/bash

# start redis-benchmark in cluster mode
 redis-benchmark --cluster -p 7000 -t set -n 50000 -c 10 -d 100 -r 100000
```

# check cluster slots

```
127.0.0.1:7000> cluster slots

1) 1) (integer) 0
   2) (integer) 5460
   3) 1) "172.19.0.4"
      2) (integer) 7000
      3) "ae739172a0539039ec31763d2bf22ee4b8368bfd"
      4) (empty array)
2) 1) (integer) 5461
   2) (integer) 10922
   3) 1) "172.19.0.2"
      2) (integer) 7001
      3) "7371a2f8537392aa13cfdbb51f05364bc610d1b6"
      4) (empty array)
3) 1) (integer) 10923
   2) (integer) 16383
   3) 1) "172.19.0.3"
      2) (integer) 7002
      3) "6fb30c9de6628d69264b60502d06118c18e2425b"
      4) (empty array)

```

```shell    

127.0.0.1:7000> cluster nodes
7371a2f8537392aa13cfdbb51f05364bc610d1b6 172.19.0.2:7001@17001 master - 0 1735255359071 2 connected 5461-10922
6fb30c9de6628d69264b60502d06118c18e2425b 172.19.0.3:7002@17002 master - 0 1735255360109 3 connected 10923-16383
ae739172a0539039ec31763d2bf22ee4b8368bfd 172.19.0.4:7000@17000 myself,master - 0 1735255358000 1 connected 0-5460
```

```shell

# verifying dbsize 

redis-cli -p 7000 dbsize
redis-cli -p 7001 dbsize
redis-cli -p 7002 dbsize

```

## scan big keys
```shell
redis-cli --bigkeys
```

## get the memory taken by big keys 
```shell
redis-cli --memkeys -i 0.1
```

under the hood, the `--memkeys` option uses the `SCAN` command to iterate over all keys in the database. It then fetches the memory usage of each key using the `MEMORY USAGE` command. The `--memkeys` option is useful for identifying memory-hungry keys in a Redis database.

1. Start a `SCAN` at cursor `0`.
2. Fetch a batch of keys (e.g., a few thousand) at a time.
3. For `--bigkeys`:
    - For each key, run commands like `TYPE` and a type-specific “length” command (e.g., `LLEN` for lists, `SCARD` for sets, etc.).
    - Keep track of which key is “largest” in each category as it goes.
4. For `--memkeys`:
    - For each key, call `MEMORY USAGE <key>` and record the value.
    - Keep a sorted list of the top memory-consuming keys.
5. Repeat until `SCAN` returns cursor `0`, meaning the entire keyspace has been traversed.

## what happens when one of the masters goes down in this setup

```
➜  experiments git:(main) ✗ redis-cli -p 7000 cluster nodes

7371a2f8537392aa13cfdbb51f05364bc610d1b6 172.19.0.2:7001@17001 master - 0 1735328483416 2 connected 5461-10922
6fb30c9de6628d69264b60502d06118c18e2425b 172.19.0.3:7002@17002 master - 0 1735328483518 3 connected 10923-16383
ae739172a0539039ec31763d2bf22ee4b8368bfd 172.19.0.4:7000@17000 myself,master - 0 1735328482000 1 connected 0-5460
```

### check cluster info
```

➜  experiments git:(main) ✗ redis-cli -p 7000 cluster info 

cluster_state:ok
cluster_slots_assigned:16384
cluster_slots_ok:16384
cluster_slots_pfail:0
cluster_slots_fail:0
cluster_known_nodes:3
cluster_size:3
cluster_current_epoch:3
cluster_my_epoch:1
cluster_stats_messages_ping_sent:4423
cluster_stats_messages_pong_sent:4430
cluster_stats_messages_sent:8853
cluster_stats_messages_ping_received:4428
cluster_stats_messages_pong_received:4423
cluster_stats_messages_meet_received:2
cluster_stats_messages_received:8853
total_cluster_links_buffer_limit_exceeded:0
```
### stop a node
```
➜  experiments git:(main) ✗ docker-compose stop redis1
```
### check cluster info
```
redis-cli -p 7001 cluster info

cluster_state:fail
cluster_slots_assigned:16384
cluster_slots_ok:10923
cluster_slots_pfail:0
cluster_slots_fail:5461
cluster_known_nodes:3
cluster_size:3
cluster_current_epoch:3
cluster_my_epoch:2
cluster_stats_messages_ping_sent:4559
cluster_stats_messages_pong_sent:4656
cluster_stats_messages_meet_sent:1
cluster_stats_messages_fail_sent:2
cluster_stats_messages_sent:9218
cluster_stats_messages_ping_received:4656
cluster_stats_messages_pong_received:4559
cluster_stats_messages_fail_received:1
cluster_stats_messages_received:9216
total_cluster_links_buffer_limit_exceeded:0
```

### verify status
```
redis-cli -p 7001 cluster nodes

7371a2f8537392aa13cfdbb51f05364bc610d1b6 172.19.0.2:7001@17001 myself,master - 0 1735328550000 2 connected 5461-10922
6fb30c9de6628d69264b60502d06118c18e2425b 172.19.0.3:7002@17002 master - 0 1735328552445 3 connected 10923-16383
ae739172a0539039ec31763d2bf22ee4b8368bfd 172.19.0.4:7000@17000 master,fail - 1735328520780 1735328518179 1 connected 0-5460
```

### try to query some keys 
```
127.0.0.1:7001> get "key:{06S}:000000031243"
(error) CLUSTERDOWN The cluster is down
```

There's a redis setting called `cluster-require-full-coverage` which is set to `yes` by default. This setting ensures that all slots in the cluster are covered by at least one node. If a master node goes down, the cluster will be in a `fail` state until the slot is covered by another node.

The default behaviour is that `cluster-require-full-coverage` is set to yes, which means that the cluster will not be able to serve queries when a master node goes down. This is because the slot that the master was responsible for is no longer covered by any node. 
The cluster will be in a fail state until the slot is covered by another node.

we could change it to `yes` and then only the impacted hash-slot will be unavailable. 

### bring the node back up
```
docker-compose start redis1
```

### verify status
```
redis-cli -p 7001 cluster nodes

7371a2f8537392aa13cfdbb51f05364bc610d1b6 172.19.0.2:7001@17001 myself,master - 0 1735329153000 2 connected 5461-10922
6fb30c9de6628d69264b60502d06118c18e2425b 172.19.0.3:7002@17002 master - 0 1735329156250 3 connected 10923-16383
ae739172a0539039ec31763d2bf22ee4b8368bfd 172.19.0.4:7000@17000 master - 0 1735329155215 1 connected 0-5460

```
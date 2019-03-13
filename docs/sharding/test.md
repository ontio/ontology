
##

start main chain:

```
./ontology --testmode --testmode-gen-block-time 10
```


test with ontology-tool

```
./main -t ShardInit
./main -t ShardCreate
./main -t ShardConfig
./main -t ShardPeerJoin
./main -t ShardActivate
./main -t ShardGasInit
./main -t ShardDepositGas
./main -t ShardQueryGas
```

start shard chain:

```
./ontology --ShardID 1 --ShardPort 20341 --enable-shard-rpc --ParentShardPort 20340
./ontology --ShardID 2 --ShardPort 20342 --enable-shard-rpc --ParentShardPort 20340
```



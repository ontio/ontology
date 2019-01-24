
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
```

start shard chain:

```
./ontology --ShardID 1 --ShardPort 20341 --ParentShardID 0 --ParentShardPort 20340
```




###Solo

start main chain:
```
./ontology --testmode --testmode-gen-block-time 10 --networkid 300 --enable-solo-shard --gasprice 0
```
1、wait gene some block then stop main chain copy -r chain dir to  shard chain dir
2、start main chain,get some ong use cli withdrawong method

3、test with ontology-tool

```
./main -t ShardInit
./main -t ShardCreate
./main -t ShardConfig
./main -t ShardPeerApply
./main -t ShardPeerApprove
./main -t ShardPeerJoin
./main -t ShardActivate
```

4、start shard chain:
create peers.recent like:
```
{"300":["127.0.0.1:20338"]}
```

```
./ontology --testmode --ShardID  1 --networkid 300 --restport 30334 --wsport 30335 --rpcport 30336 --nodeport 30338 --enable-consensus  --enable-solo-shard
```

###vbft
1、start main chain,you should set networkid
```
./ontology --config config.json  --gaslimit 20000  --networkid 600

```
2、test with ontology-tool
``` 
./main -t TransferOntMultiSign
./main -t TransferFromOngMultiSign
./main -t ShardInit
./main -t ShardCreate
./main -t ShardConfig
./main -t ShardPeerApply
./main -t ShardPeerApprove
./main -t ShardPeerJoin
./main -t ShardActivate
```
3、start shard chain:
```
 ./ontology  --ShardID  1  --config config.json --networkid 600 --loglevel 2 --restport 30334 --wsport 30335 --rpcport 30336 --nodeport 30338 --enable-consensus
 ```

 #### shardchain networkid must be consistent rootchain networkid
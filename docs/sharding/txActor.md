
## TxActor 

handle all new transactions.

all remote transactions are processed by TxActor

1. http add received transaction to txn-pool
2. http get transaction from txn-pool, get transaction status
3. p2p add received transaction to txn-pool
4. p2p get transaction from txn-pool


## TxPoolActor
support consensus module

requests processed by TxPoolActor

1. http: count of pending transaction
2. consensus: pending transaction to be consensused


## VerifyRspActor

handle the response from validators

request processed by VerifyRspActor

1. notify verification results from txnpool_worker


## NetActor

to send msg to the net actor

broadcast new transactions after they are verified



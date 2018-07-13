# Ontology Config File Instruction

[English | [中文](config_CN.md)]

Ontology would support multiple consensus algorithm(VBFT/DBFT/RBFT/SBFT/PoW)，and provide a pluggable consensus switching mechanism. Ontology has 
supported VBFT and DBFT, they have different config method. It is mainly embodied in the configuration file.

## Content
* [DBFT Configuration](#dbft-configuration)
* [VBFT Configuration](#vbft-configuration)

## DBFT Configuration

DBFT configure file is [`config-dbft.json`](config-dbft.json)，as follows:

```json
{
  "SeedList": [
    "ip1:20318",
    "ip2:20318",
    "ip3:20318",
    "ip4:20318"
  ],
  "ConsensusType":"dbft",
  "DBFT":{
    "Bookkeepers": [
      "bookKeeper1",
      "bookKeeper2",
      "bookKeeper3",
      "bookKeeper4"
    ],
    "GenBlockTime":6
  }
}
```

SeedList：This is used to configure seed node of ontology network。Seed node is the link entrance of ontology network. When a new node link to 
ontology network, it request network information from seed node. At least one seed node should be configured in the configuration file.

ConsensusType：Consensus algorithm type, it indicates waht consensus is configured in the configuration file. The value could be set as "VBFT"
or "DBFT".

DBFT: DBFT consensus configuration，as follows：

- Bookkeepers：bookkeeper, the configuration is bookkeeper's public key;

- GenBlockTime：block generation interval, define as seconds.

## VBFT Configuration

VBFTis the default consensus mechanism for ontology at present, with [`config-vbft.json`](config-vbft.json) configuration files, as follows:

```json
{
  "SeedList": [
    "127.0.0.1:20338"
  ],
  "ConsensusType":"vbft",
  "VBFT":{
    "n":7,
    "c":2,
    "k":7,
    "l":112,
    "block_msg_delay":10000,
    "hash_msg_delay":10000,
    "peer_handshake_timeout":10,
    "max_block_change_view":1000,
    "admin_ont_id":"did:ont:AVaSGN1ugQJBS7R7ZcVwAoWLVK6onBgfyg",
    "min_init_stake":100000,
    "vrf_value":"1c9810aa9822e511d5804a9c4db9dd08497c31087b0daafa34d768a3253441fa20515e2f30f81741102af0ca3cefc4818fef16adb825fbaa8cad78647f3afb590e",
    "vrf_proof":"c57741f934042cb8d8b087b44b161db56fc3ffd4ffb675d36cd09f83935be853d8729f3f5298d12d6fd28d45dde515a4b9d7f67682d182ba5118abf451ff1988",
    "peers":[
      {
        "index":1,
        "peerPubkey":"028541d32f3b09180b00affe67a40516846c16663ccb916fd2db8106619f087527",
        "address":"AVaSGN1ugQJBS7R7ZcVwAoWLVK6onBgfyg",
        "initPos":100000
      },
      {
        "index":2,
        "peerPubkey":"02dfb161f757921898ec2e30e3618d5c6646d993153b89312bac36d7688912c0ce",
        "address":"AVaSGN1ugQJBS7R7ZcVwAoWLVK6onBgfyg",
        "initPos":200000
      },
      {
        "index":3,
        "peerPubkey":"039dab38326268fe82fb7967fe2e7f5f6eaced6ec711148a66fbb8480c321c19dd",
        "address":"AVaSGN1ugQJBS7R7ZcVwAoWLVK6onBgfyg",
        "initPos":300000
      },
      {
        "index":4,
        "peerPubkey":"0384f2729bc5d9b14dcbf17aba108261dc7ad867127e413d3c8bfb4731739687b3",
        "address":"AVaSGN1ugQJBS7R7ZcVwAoWLVK6onBgfyg",
        "initPos":400000
      },
      {
        "index":5,
        "peerPubkey":"03362f99284daa9f581fab596516f75475fc61a5f80de0e268a68430dc7589859c",
        "address":"AVaSGN1ugQJBS7R7ZcVwAoWLVK6onBgfyg",
        "initPos":300000
      },
      {
        "index":6,
        "peerPubkey":"03db6e37a2d897f2d61b42dcd478323a8a20c3444af4ee29653849f38d0bdb67f4",
        "address":"AVaSGN1ugQJBS7R7ZcVwAoWLVK6onBgfyg",
        "initPos":200000
      },
      {
        "index":7,
        "peerPubkey":"0298fe9f22e9df64f6bfcc1c2a14418846cffdbbf510d261bbc3fa6d47073df9a2",
        "address":"AVaSGN1ugQJBS7R7ZcVwAoWLVK6onBgfyg",
        "initPos":100000
      }
    ]
  }
}
```
SeedList：same as SeedList in DBFT configuration。

ConsensusType：same as ConsensusType in DBFT configuration。

VBFT：VBFT consensus configuration，as follows：

- n：the number of node in whole ontology network;
- c：the number of fault-tolerant nodes;
- k：the number of consensus nodes;
- l：the number of POS table items;
- block_msg_delay：block message maximum broadcast delay;
- hash_msg_delay：hash message maximum broadcast delay;
- peer_handshake_timeout：node handshake connection timeout time limit;
- max_block_change_view：Consensus period;
- admin_ont_id：ONT ID of administrator;
- min_init_stake：the minimum amount of initial stack ONT;
- vrf_value：compose a verifiable random value with'vrf_proof';
- vrf_proof：compose a verifiable random value with'vrf_value';
- peers：bookkeeper in consensus nodes configuration, as follows：

	- index：index;
	- peerPubkey：bookkeeper's public key;
	- address：bookkeeper's address;
	- initPos：initial mortgage ONT number of this bookkeeper.


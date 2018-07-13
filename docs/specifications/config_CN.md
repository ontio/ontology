# Ontology 配置文件说明

[[English](config.md) | 中文]

Ontology计划支持多种共识算法，并提供可插拔式的共识切换机制。目前已支持的共识算法为VBFT与DBFT，这两种算法有不同的配置方法，主要体现在配置文件里。
现对配置文件的内容和结构解释如下。

## 目录
* [DBFT 配置](#dbft-配置)
* [VBFT 配置](#vbft-配置)

## DBFT 配置

DBFT配置文件为[`config-dbft.json`](config-dbft.json)，如下：

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
SeedList：用来配置ontology网络的种子节点。种子节点是ontology网络的链接入口，新节点加入ontology网络时，会先向种子节点请求网络相关信息。
配置文件里至少需配置一个种子节点。

ConsensusType：共识模式，指示该配置文件配置的是何种共识，目前支持"VBFT"和"DBFT"。

DBFT: DBFT共识配置，内容如下：

- Bookkeepers：记账人，用来配置记账人的公钥，需要配置四个；

- GenBlockTime：区块生成时间，单位为秒。

## VBFT 配置

VBFT是目前ontology的默认共识机制，配置文件为[`config-vbft.json`](config-vbft.json)，如下：

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
SeedList：与DBFT配置文件的SeedList相同。

ConsensusType：与DBFT配置文件的ConsensusType相同。

VBFT：VBFT共识配置，内容如下：

- n：链上节点总数(网络规模)；
- c：容错节点数量；
- k：共识节点数目；
- l：POS表条目数；
- block_msg_delay：区块消息最大广播延迟；
- hash_msg_delay：哈希消息最大广播延迟；
- peer_handshake_timeout：节点握手连接超时时限；
- max_block_change_view：共识周期；
- admin_ont_id：管理员的ONT ID；
- min_init_stake：初始抵押ONT的最小数量；
- vrf_value：与'vrf_proof'一起构成可验证随机值；
- vrf_proof：与'vrf_value'一起构成可验证随机值；
- peers：共识节点(记账人)配置，内容如下：

	- index：序号；
	- peerPubkey：记账人公钥；
	- address：记账人地址；
	- initPos：该记账人初始抵押ONT数目。


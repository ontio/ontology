# 治理合约API
## 简介
本文档主要描述Ontology治理合约的API接口，用户通过该合约可以申请参与共识节点的竞选，抵押投票给参选节点，退出共识节点的竞选等，抵押的ONT会按照一定的规则产生收益。
## API
### InitConfig
功能：初始化治理合约，仅在在创世块创建时调用，系统方法。

```text
方法名："initConfig"

参数：无

返回值：bool， error
```
### RegisterCandidate
功能：抵押一定的ONT，消耗一定的额外ONG，申请成为候选节点。

```text
方法名："registerCandidate"

参数：
0       String       节点公钥
1       Address      钱包地址
2       Uint32       抵押的ONT数量
3       ByteArray    调用者的OntID
4       Uint64       调用者公钥序号

返回值：bool， error
```
### RegisterCandidateTransferFrom
功能：抵押一定的ONT，消耗一定的额外ONG，申请成为候选节点，供合约调用。

```text
方法名："registerCandidateTransferFrom"

参数：
0       String       节点公钥
1       Address      钱包地址
2       Uint32       抵押的ONT数量
3       ByteArray    调用者的OntID
4       Uint64       调用者公钥序号

返回值：bool， error
```
### BlackNode
功能：管理员审核，将节点放入黑名单，同时触发节点退出流程，不返还节点的InitPos。

```text
方法名："blackNode"

参数：
0       Array{String}   要放入黑名单的节点列表

返回值：bool， error
```
### WhiteNode
功能：管理员审核，将节点从黑名单中移除，节点的InitPos退还。

```text
方法名："whiteNode"

参数：
0       String       节点公钥

返回值：bool， error
```
### QuitNode
功能：节点申请退出，进入正常退出流程，钱包地址要与申请时相同。

```text
方法名："quitNode"

参数：
0       String       节点公钥
1       Address      钱包地址

返回值：bool， error
```
### AuthorizeForPeer
功能：通过抵押ONT的方式向节点投票。

```text
方法名："authorizeForPeer"

参数：
0       Address         钱包地址
1       Array{String}   要投票的节点列表
2       Array{Uint32}   要给节点投的票数

返回值：bool， error
```
### AuthorizeForPeerTransferFrom
功能：通过抵押ONT的方式向节点投票，供合约调用。

```text
方法名："authorizeForPeerTransferFrom"

参数：
0       Address         钱包地址
1       Array{String}   要投票的节点列表
2       Array{Uint32}   要给节点投的票数

返回值：bool， error
```
### UnAuthorizeForPeer
功能：赎回抵押ONT的方式向节点取消投票。

```text
方法名："unAuthorizeForPeer"

参数：
0       Address         钱包地址
1       Array{String}   要取消投票的节点列表
2       Array{Uint32}   要向节点取消的票数

返回值：bool， error
```
### Withdraw
功能：取出处于未冻结状态的抵押ONT。

```text
方法名："withdraw"

参数：
0       Address         钱包地址
1       Array{String}   要从哪些节点去吃抵押的列表
2       Array{Uint32}   要从节点取出抵押数

返回值：bool， error
```
### WithdrawOng
功能：提取解绑ong。

```text
方法名："withdrawOng"

参数：
0       Address         钱包地址

返回值：bool， error
```

### WithdrawFee
功能：提取手续费分红。

```text
方法名："WithdrawFee"

参数：
0       Address         钱包地址

返回值：bool， error
```

### CommitDpos
功能：共识切换，按照当前投票结果切换共识，系统方法。

```text
方法名："commitDpos"

参数：无

返回值：bool， error
```
### UpdateConfig
功能：更新共识配置，只能由管理员调用。

```text
方法名："updateConfig"

参数：
0       Uint32      网络规模
1       Uint32      容错数目
2       Uint32      共识节点数
3       Uint32      Pos表长度
4       Uint32      区块消息最大广播延迟(ms)
5       Uint32      哈希消息最大广播延迟(ms)
6       Uint32      节点握手超时时间(s)
7       Uint32      共识周期

返回值：bool， error
```
### UpdateGlobalParam
功能：更新全局参数，只能由管理员调用。

```text
方法名："updateGlobalParam"

参数：
0       Uint32      节点申请参与共识选举的摩擦费
1       Uint32      节点申请参与共识选举的最小抵押
2       Uint32      共识和候选节点总数上限
3       Uint32      节点能接受的投票上限倍数
4       Uint32      共识节点激励比例(0-100)
5       Uint32      候选节点激励比例(0-100)
6       Uint32      激励系数
7       UInt32      惩罚系数

返回值：bool， error
```
### UpdateSplitCurve
功能：更新ONG分配曲线，只能由管理员调用。

```text
方法名："updateSplitCurve"

参数：
0       Array{Uint64}      分配曲线的Y轴散点值

返回值：bool， error
```
### TransferPenalty
功能：取出作恶节点的扣留抵押，只能由管理员调用。

```text
方法名："transferPenalty"

参数：
0       String      节点公钥
1       Address     钱包地址

返回值：bool， error
```

### ChangeMaxAuthorization
功能：节点修改自己接受的最大授权ONT数量。

```text
方法名："changeMaxAuthorization"

参数：
0       String      节点公钥
1       Address     钱包地址
2       Uint32      接受的最大授权

返回值：bool， error
```

### SetFeePercentage

功能：节点设置自己独占激励的比例。

```text
方法名："setFeePercentage"

参数：
0       String      节点公钥
1       Address     钱包地址
2       Uint32      独占节点的激励比例
3       Uint32      独占用户的激励比例

返回值：bool， error
```

### AddInitPos

功能：节点增加initPos接口，只能由节点所有者调用。

```text
方法名："addInitPos"

参数：
0       String      节点公钥
1       Address     钱包地址
2       Uint32      增加的抵押数量

返回值：bool， error
```

### ReduceInitPos
功能：节点减少initPos接口，只能由节点所有者调用，initPos不能低于承诺值，不能低于已接受授权数量的1/10。

```text
方法名："reduceInitPos"

参数：
0       String      节点公钥
1       Address     钱包地址
2       Uint32      减少的抵押数量

返回值：bool， error
```

### SetPromisePos
功能：设置节点的承诺抵押，只有管理员可以调用。

```text
方法名："setPromisePos"

参数：
0       String      节点公钥
1       Uint32      承诺抵押数量

返回值：bool， error
```

### UpdateGlobalParam2
功能：设置合约全局参数，只有管理员可以调用。

```text
方法名："updateGlobalParam2"

参数：
0       Uint32      授权的最小ONT倍数
1       Uint32      能够分到激励的节点数
2       Uint32      Dapp获得的奖励比例

返回值：bool， error
```

### SetGasAddress
功能：设置Dapp收钱账户地址，只有管理员可以调用，不设置默认不给Dapp账户分钱。

```text
方法名："setGasAddress"

参数：
0       Address      Dapp的收钱地址

返回值：bool， error
```
### GetPeerPool
功能：查询共识节点和候选节点详细信息列表

```text
方法名："getPeerPool"

参数：无

返回值：[]byte， error
```
返回值的序列化：
```golang
type PeerPoolListForVm struct {
	PeerPoolList []*PeerPoolItemForVm
}

func (this *PeerPoolListForVm) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(len(this.PeerPoolList)))
	for _, v := range this.PeerPoolList {
		v.Serialization(sink)
	}
}

type PeerPoolItemForVm struct {
	Index       uint32         //peer index
	PeerAddress common.Address //peer address
	Address     common.Address //peer owner
	Status      Status         //peer status
	InitPos     uint64         //peer initPos
	TotalPos    uint64         //total authorize pos this peer received
}

func (this *PeerPoolItemForVm) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Index)
	this.PeerAddress.Serialization(sink)
	this.Address.Serialization(sink)
	this.Status.Serialization(sink)
	sink.WriteUint64(this.InitPos)
	sink.WriteUint64(this.TotalPos)
}
```
### GetPeerInfo
功能：根据节点地址查询节点详细信息

```text
方法名："getPeerInfo"

参数：
0       Address      节点地址

返回值：[]byte， error
```
返回值的序列化：
```golang
type PeerPoolItemForVm struct {
	Index       uint32         //peer index
	PeerAddress common.Address //peer address
	Address     common.Address //peer owner
	Status      Status         //peer status
	InitPos     uint64         //peer initPos
	TotalPos    uint64         //total authorize pos this peer received
}

func (this *PeerPoolItemForVm) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Index)
	this.PeerAddress.Serialization(sink)
	this.Address.Serialization(sink)
	this.Status.Serialization(sink)
	sink.WriteUint64(this.InitPos)
	sink.WriteUint64(this.TotalPos)
}
```

### GetPeerPoolByAddress
功能：根据质押地址查询节点详细信息列表

```text
方法名："getPeerPoolByAddress"

参数：
0       Address      节点地址

返回值：[]byte， error
```
返回值的序列化：
```golang
type PeerPoolListForVm struct {
	PeerPoolList []*PeerPoolItemForVm
}

func (this *PeerPoolListForVm) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(len(this.PeerPoolList)))
	for _, v := range this.PeerPoolList {
		v.Serialization(sink)
	}
}

type PeerPoolItemForVm struct {
	Index       uint32         //peer index
	PeerAddress common.Address //peer address
	Address     common.Address //peer owner
	Status      Status         //peer status
	InitPos     uint64         //peer initPos
	TotalPos    uint64         //total authorize pos this peer received
}

func (this *PeerPoolItemForVm) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Index)
	this.PeerAddress.Serialization(sink)
	this.Address.Serialization(sink)
	this.Status.Serialization(sink)
	sink.WriteUint64(this.InitPos)
	sink.WriteUint64(this.TotalPos)
}
```

### GetAuthorizeInfo

```text
方法名："getAuthorizeInfo"

参数：
0       PublicKey    节点公钥
1       Address      投票人地址

返回值：[]byte， error
```

返回值的序列化：

```go
type AuthorizeInfo struct {
	PeerPubkey           string
	Address              common.Address
	ConsensusPos         uint64 //pos deposit in consensus node
	CandidatePos         uint64 //pos deposit in candidate node
	NewPos               uint64 //deposit new pos to consensus or candidate node, it will be calculated in next epoch, you can withdrawal it at any time
	WithdrawConsensusPos uint64 //unAuthorized pos from consensus pos, frozen until next next epoch
	WithdrawCandidatePos uint64 //unAuthorized pos from candidate pos, frozen until next epoch
	WithdrawUnfreezePos  uint64 //unfrozen pos, can withdraw at any time
}

func (this *AuthorizeInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	this.Address.Serialization(sink)
	sink.WriteUint64(this.ConsensusPos)
	sink.WriteUint64(this.CandidatePos)
	sink.WriteUint64(this.NewPos)
	sink.WriteUint64(this.WithdrawConsensusPos)
	sink.WriteUint64(this.WithdrawCandidatePos)
	sink.WriteUint64(this.WithdrawUnfreezePos)
}
```

### GetAddressFee

```text
方法名："getAddressFee"

参数：
0       Address      用户地址

返回值：[]byte， error
```

返回值的序列化：

```go
type SplitFeeAddress struct { //table record each address's ong motivation
	Address common.Address
	Amount  uint64
}

func (this *SplitFeeAddress) Serialization(sink *common.ZeroCopySink) {
	this.Address.Serialization(sink)
	sink.WriteUint64(this.Amount)
}
```


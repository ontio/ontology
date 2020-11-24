## 数据
```
当前view值:
key: []byte("governanceView")
value:
type GovernanceView struct {
	View   uint32           view值
	Height uint32           切换高度
	TxHash common.Uint256   切换交易Hash
}
用途: 记录当前view值
```


```
节点池:
key: []byte("peerPool") + view
value:
type PeerPoolMap struct {
	PeerPoolMap map[string]*PeerPoolItem
}
type PeerPoolItem struct {
	Index      uint32           节点唯一ID
	PeerPubkey string           节点公钥
	Address    common.Address   节点钱包地址
	Status     Status           节点状态
	InitPos    uint64           节点初始抵押
	TotalPos   uint64           节点总抵押
}
用途: 记录节点的信息
```

```
用户质押信息：
key: []byte("authorizeInfoPool") + peerPubKey + address
value:
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
用途：记录用户的投票信息
```

```
共识参数配置: 
key: []byte("vbftConfig")
value: 
type Configuration struct {
	N                    uint32     网络规模
	C                    uint32     容错数目
	K                    uint32     共识节点数
	L                    uint32     Pos表长度
	BlockMsgDelay        uint32     区块消息最大广播延迟(ms)
	HashMsgDelay         uint32     哈希消息最大广播延迟(ms)
	PeerHandshakeTimeout uint32     节点握手超时时间(s)
	MaxBlockChangeView   uint32     共识周期
}
用途: 记录合约的参数配置
```

```
合约参数配置: 
key: []byte("globalParam")
value: 
type GlobalParam struct {
	CandidateFee    uint32  节点申请参与共识选举的摩擦费
	MinInitStake    uint32  节点申请参与共识选举的最小抵押
	CandidateNum    uint32  共识和候选节点总数上限
	PosLimit        uint32  节点能接受的投票上限倍数
	A               uint32  共识节点激励比例(0-100)
	B               uint32  候选节点激励比例(0-100)
	Yita            uint32  激励系数
	Penalty         uint32  惩罚系数
}
用途: 记录合约的参数配置
```

```
合约参数配置2: 
key: []byte("globalParam2")
value: 
type GlobalParam2 struct {
	MinAuthorizePos      uint32 //min ONT of each authorization, 500 default
	CandidateFeeSplitNum uint32 //num of peer can receive motivation(include consensus and candidate)
	DappFee              uint32 //fee split to dapp bonus
	Field2               []byte //reserved field
	Field3               []byte //reserved field
	Field4               []byte //reserved field
	Field5               []byte //reserved field
	Field6               []byte //reserved field
}
用途: 记录合约的参数配置
```

```
ONG分配曲线: 
key: []byte("splitCurve")
value: 
type SplitCurve struct {
	Yi      []uint64     分配曲线的Y轴散点值
}
用途: 记录ONG分配曲线
```

```
新节点ID：
key: []byte("candidateIndex")
value: uint32
用途: 记录待分配的节点ID
```

```
黑名单：
key: []byte("blackList") + peerPubKey
value:
type BlackListItem struct {
	PeerPubkey string         //peerPubkey in black list
	Address    common.Address //the owner of this peer
	InitPos    uint64         //initPos of this peer
}
用途: 记录进入黑名单的节点
```

```
用户总质押：
key: []byte("totalStake") + address
value:
type TotalStake struct { //table record each address's total stake in this contract
	Address    common.Address
	Stake      uint64
	TimeOffset uint32
}
用途: 记录用户进入治理合约的总质押
```

```
惩罚的质押：
key: []byte("penaltyStake") + peerPubKey
value:
type PenaltyStake struct { //table record penalty stake of peer
	PeerPubkey   string //peer pubKey of penalty stake
	InitPos      uint64 //initPos penalty
	AuthorizePos uint64 //authorize pos penalty
	TimeOffset   uint32 //time used for calculate unbound ong
	Amount       uint64 //unbound ong that this penalty unbounded
}
用途: 记录节点拉黑后惩罚的总质押
```

```
节点属性：
key: []byte("peerAttributes") + peerPubKey
value:
type PeerAttributes struct {
	PeerPubkey   string
	MaxAuthorize uint64 //max authorzie pos this peer can receive(number of ont), set by peer owner
	T2PeerCost   uint64 //candidate or consensus node doesn't share initpos income percent with authorize users, 100 means node will take all incomes, it will take effect in view T + 2
	T1PeerCost   uint64 //candidate or consensus node doesn't share initpos income percent with authorize users, 100 means node will take all incomes, it will take effect in view T + 1
	TPeerCost    uint64 //candidate or consensus node doesn't share initpos income percent with authorize users, 100 means node will take all incomes, it will take effect in view T
	T2StakeCost  uint64 //candidate or consensus node doesn't share stake income percent with authorize users, it will take effect in view T + 2, 101 means 0, 0 means null
	T1StakeCost  uint64 //candidate or consensus node doesn't share stake income percent with authorize users, it will take effect in view T + 1, 101 means 0, 0 means null
	TStakeCost   uint64 //candidate or consensus node doesn't share stake income percent with authorize users, it will take effect in view T, 101 means 0, 0 means null
	Field4       []byte //reserved field
}

```

```
治理合约中已分但未提取的ong
key: []byte("splitFee")
value: uint64
用途: 记录治理合约中已分但未提取的ong
```

```text
用户未提取的ong
key: []byte("splitFeeAddress") + address
value:
type SplitFeeAddress struct { //table record each address's ong motivation
	Address common.Address
	Amount  uint64
}
用途: 记录用户未提取的ong
```

```
节点的承诺质押：
key: []byte("promisePos") + peerPubKey
value:
type PromisePos struct {
	PeerPubkey string
	PromisePos uint64
}
用途: 记录节点的承诺质押，其初始质押不能少于承诺质押
```

```
治理合约还未生效的配置：
key: []byte("preConfig")
value:
type PreConfig struct {
	Configuration *Configuration
	SetView       uint32
}
type Configuration struct {
	N                    uint32
	C                    uint32
	K                    uint32
	L                    uint32
	BlockMsgDelay        uint32
	HashMsgDelay         uint32
	PeerHandshakeTimeout uint32
	MaxBlockChangeView   uint32
}
用途: 记录治理合约还未生效的配置，此配置将在周期切换后生效
```

```
dApp激励地址：
key: []byte("gasAddress")
value: 
type GasAddress struct {
	Address common.Address
}
用途: dApp激励地址，手续费先按照比例分给该地址
```


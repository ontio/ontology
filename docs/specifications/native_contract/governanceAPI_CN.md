# 治理合约API
## 简介
本文档主要描述Ontology治理合约的API接口，包括调用示例。
## API
### InitConfig
功能：初始化治理合约，仅在在创世块创建时调用，系统方法。

方法名："initConfig"

参数：无

调用示例：
```go
	contractAddress := genesis.GovernanceContractAddress
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "callSplit",
	}
```
### RegisterSyncNode
功能：抵押一定的ONT，消耗一定的额外ONG，申请成为同步节点。 

方法名："registerSyncNode"

参数：smartcontract/service/native/governance.RegisterSyncNodeParam

调用示例：
```go
	params := &governance.RegisterSyncNodeParam{
	    ...
	}
	contractAddress := genesis.GovernanceContractAddress
	bf := new(bytes.Buffer)
	if err := params.Serialize(bf); err != nil {
		ctx.LogError("Serialize params error:%s", err)
		return false
	}
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "registerSyncNode",
		Args:    bf.Bytes(),
	}
```
### ApproveSyncNode
功能：管理员审核，通过则成为同步节点，只有管理员能够调用。

方法名："approveSyncNode"

参数：smartcontract/service/native/governance.ApproveSyncNodeParam

调用示例：
```go
	params := &governance.ApproveCandidateParam{
		...
	}
	contractAddress := genesis.GovernanceContractAddress
	bf := new(bytes.Buffer)
	if err := params.Serialize(bf); err != nil {
		ctx.LogError("Serialize params error:%s", err)
		return false
	}
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "approveSyncNode",
		Args:    bf.Bytes(),
	}
```
### RegisterCandidate
功能：申请成为候选节点，钱包地址要与申请同步节点时相同。

方法名："registerCandidate"

参数：smartcontract/service/native/governance.RegisterCandidateParam

调用示例：
```go
	params := &governance.RegisterCandidateParam{
        ...
	}
	contractAddress := genesis.GovernanceContractAddress
	bf := new(bytes.Buffer)
	if err := params.Serialize(bf); err != nil {
		ctx.LogError("Serialize params error:%s", err)
		return false
	}
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "registerCandidate",
		Args:    bf.Bytes(),
	}
```

### ApproveCandidate
功能：管理员审核，通过则成为候选节点，只有管理员能够调用。

方法名："approveCandidate"

参数：smartcontract/service/native/governance.ApproveCandidateParam

调用示例：
```go
	params := &governance.ApproveCandidateParam{
		...
	}
	contractAddress := genesis.GovernanceContractAddress
	bf := new(bytes.Buffer)
	if err := params.Serialize(bf); err != nil {
		ctx.LogError("Serialize params error:%s", err)
		return false
	}
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "approveCandidate",
		Args:    bf.Bytes(),
	}
```
### BlackNode
功能：管理员审核，将节点放入黑名单，同时触发节点退出流程，不返还节点的InitPos。

方法名："blackNode"

参数：smartcontract/service/native/governance.BlackNodeParam

调用示例：
```go
	params := &governance.BlackNodeParam{
		...
	}
	contractAddress := genesis.GovernanceContractAddress
	bf := new(bytes.Buffer)
	if err := params.Serialize(bf); err != nil {
		ctx.LogError("Serialize params error:%s", err)
		return false
	}
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "blackNode",
		Args:    bf.Bytes(),
	}
```
### WhiteNode
功能：管理员审核，将节点从黑名单中移除，节点的InitPos退还。

方法名："whiteNode"

参数：smartcontract/service/native/governance.WhiteNodeParam

调用示例：
```go
	params := &governance.WhiteNodeParam{
		...
	}
	contractAddress := genesis.GovernanceContractAddress
	bf := new(bytes.Buffer)
	if err := params.Serialize(bf); err != nil {
		ctx.LogError("Serialize params error:%s", err)
		return false
	}
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "whiteNode",
		Args:    bf.Bytes(),
	}
```
### QuitNode
功能：节点申请退出，进入正常退出流程，钱包地址要与申请时相同。

方法名："quitNode"

参数：smartcontract/service/native/governance.QuitNodeParam

调用示例：
```go
	params := &governance.QuitNodeParam{
		...
	}
	contractAddress := genesis.GovernanceContractAddress
	bf := new(bytes.Buffer)
	if err := params.Serialize(bf); err != nil {
		ctx.LogError("Serialize params error:%s", err)
		return false
	}
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "quitNode",
		Args:    bf.Bytes(),
	}
```
### VoteForPeer
功能：向节点投票，票数为负值表示取消对该节点的投票。

方法名："voteForPeer"

参数：smartcontract/service/native/governance.VoteForPeerParam

调用示例：
```go
	params := &governance.VoteForPeerParam{
		...
	}
	contractAddress := genesis.GovernanceContractAddress
	bf := new(bytes.Buffer)
	if err := params.Serialize(bf); err != nil {
		ctx.LogError("Serialize params error:%s", err)
		return false
	}
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "voteForPeer",
		Args:    bf.Bytes(),
	}
```
### Withdraw
功能：取出处于未冻结状态的抵押ONT。

方法名："withdraw"

参数：smartcontract/service/native/governance.WithdrawParam

调用示例：
```go
	params := &governance.WithdrawParam{
		...
	}
	contractAddress := genesis.GovernanceContractAddress
	bf := new(bytes.Buffer)
	if err := params.Serialize(bf); err != nil {
		ctx.LogError("Serialize params error:%s", err)
		return false
	}
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "withdraw",
		Args:    bf.Bytes(),
	}
```
### CommitDpos
功能：共识切换，按照当前投票结果切换共识，系统方法。

方法名："commitDpos"

参数：无

调用示例：
```go
	contractAddress := genesis.GovernanceContractAddress
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "commitDpos",
	}
```
### VoteCommitDpos
功能：抵押一定的ONT来投票，要求马上切换共识。

方法名："voteCommitDpos"

参数：smartcontract/service/native/governance.VoteCommitDposParam

调用示例：
```go
	params := &governance.VoteCommitDposParam{
		...
	}
	contractAddress := genesis.GovernanceContractAddress
	bf := new(bytes.Buffer)
	if err := params.Serialize(bf); err != nil {
		ctx.LogError("Serialize params error:%s", err)
		return false
	}
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "voteCommitDpos",
		Args:    bf.Bytes(),
	}
```
### UpdateConfig
功能：更新共识配置，只能由管理员调用。

方法名："updateConfig"

参数：smartcontract/service/native/governance.Configuration

调用示例：
```go
    config = &governance.Configuration{
		...
	}
	contractAddress := genesis.GovernanceContractAddress
	bf := new(bytes.Buffer)
	if err := config.Serialize(bf); err != nil {
		ctx.LogError("Serialize params error:%s", err)
		return false
	}
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "updateConfig",
		Args:    bf.Bytes(),
	}
```
### UpdateGlobalParam
功能：更新全局参数，只能由管理员调用。

方法名："updateGlobalParam"

参数：smartcontract/service/native/governance.GlobalParam

调用示例：
```go
    config = &governance.GlobalParam{
		...
	}
	contractAddress := genesis.GovernanceContractAddress
	bf := new(bytes.Buffer)
	if err := globalParam.Serialize(bf); err != nil {
		ctx.LogError("Serialize params error:%s", err)
		return false
	}
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "updateGlobalParam",
		Args:    bf.Bytes(),
	}
```
### CallSplit
功能：管理员调用，按照上一轮投票结果进行ONG分配。

方法名："callSplit"

参数：无

调用示例：
```go
	contractAddress := genesis.GovernanceContractAddress
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "callSplit",
	}
```
# Governance Contract API
## Introduction
This document describes api of governance contract used in gntology network, include invoke examples.
## API
### InitConfig
function: Initialize the governance contract, only can be invoked in the genesis block, system api。

method: "initConfig"

args: nil

invoke example:
```go
	contractAddress := genesis.GovernanceContractAddress
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "callSplit",
	}
```
### RegisterSyncNode
function: deposit some ONT, expend some extra ONG, apply for a syncNode. 

method: "registerSyncNode"

args: smartcontract/service/native/governance.RegisterSyncNodeParam

invoke example:
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
function: review by administrator, become syncNode if passed, only can be invoked by administrator.

method: "approveSyncNode"

args: smartcontract/service/native/governance.ApproveSyncNodeParam

invoke example:
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
function: apply for a candidate node, address must be the same as registerSyncNode.

method: "registerCandidate"

args: smartcontract/service/native/governance.RegisterCandidateParam

invoke example:
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
function: review by administrator, become candidate node if passed, only can be invoked by administrator.

method: "approveCandidate"

args: smartcontract/service/native/governance.ApproveCandidateParam

invoke example:
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
function: review by administrator, put a node into black list. Trigger node quit process at the same time but freeze the initPos.

method: "blackNode"

args: smartcontract/service/native/governance.BlackNodeParam

invoke example:
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
function: review by administrator, remove node from black list, and refund initPos.

method: "whiteNode"

args: smartcontract/service/native/governance.WhiteNodeParam

invoke example:
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
function: node applies for quiting, trigger normal quit process, address must be the same as registerSyncNode.

method: "quitNode"

args: smartcontract/service/native/governance.QuitNodeParam

invoke example:
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
function: vote for a node, negative pos means cancel vote.

method: "voteForPeer"

args: smartcontract/service/native/governance.VoteForPeerParam

invoke example:
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
function: withdraw unfreezed ONT deposited.

method: "withdraw"

args: smartcontract/service/native/governance.WithdrawParam

invoke example:
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
function: change consensus according to vote result, system invoke.

method: "commitDpos"

args: nil

invoke example:
```go
	contractAddress := genesis.GovernanceContractAddress
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "commitDpos",
	}
```
### VoteCommitDpos
function: deposit some ONT to trigger change consensus.

method: "voteCommitDpos"

args: smartcontract/service/native/governance.VoteCommitDposParam

invoke example:
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
function: update consensus config, only can be invoked by administrator.

method: "updateConfig"

args: smartcontract/service/native/governance.Configuration

invoke example:
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
function: update global params, only can be invoked by administrator.

method: "updateGlobalParam"

args: smartcontract/service/native/governance.GlobalParam

invoke example:
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
function: split ONG according to vote result of previous view, only can be invoked by administrator. 

method: "callSplit"

args: nil

invoke example:
```go
	contractAddress := genesis.GovernanceContractAddress
	crt := &cstates.Contract{
		Address: contractAddress,
		Method:  "callSplit",
	}
```
# Native Contract API : Param
* [Introduction](#introduction)
* [Contract Method](#contract-method)
* [How to get a parameter](#how-to-get-a-parameter)

## Introduction
This document describes the global parameter manager native contract used in the ontology network.

## Contract Method

### ParamInit
Initialize the contract, invoked in the gensis block.

method: init

args: nil

#### example
```
  init := states.Contract{
		Address: ParamContractAddress,
		Method:  "init",
	}
```
### TransferAdmin
Transfer the administrator of this contract, transferd account should be approved by old admin.

method: transferAdmin

args: smartcontract/service/native/states.Admin

#### example
```
	var destinationAdmin states.Admin
	destinationAdmin.Version = 0x2d
	destinationAdmin.Address, _ = common.AddressFromBase58("TA4YPDwvAieSfoJEJXzAU6ruQZgfx4KYrD")
	adminBuffer := new(bytes.Buffer)
	if err := destinationAdmin.Serialize(adminBuffer); err != nil {
		fmt.Println("Serialize admin struct error.")
		os.Exit(1)
	}

	contract := &sstates.Contract{
		Address: genesis.ParamContractAddress,
		Method:  "transferAdmin",
		Args:    adminBuffer.Bytes(),
	}
```

### ApproveAdmin
Approved administrator to a other account, should be invoked by administrator.

method: approveAdmin

args: smartcontract/service/native/states.Admin

#### example
```
	var destinationAdmin states.Admin
	destinationAdmin.Version = 0x2d
	destinationAdmin.Address, _ = common.AddressFromBase58("TA4YPDwvAieSfoJEJXzAU6ruQZgfx4KYrD")
	adminBuffer := new(bytes.Buffer)
	if err := destinationAdmin.Serialize(adminBuffer); err != nil {
		fmt.Println("Serialize admins struct error.")
		os.Exit(1)
	}
	contract := &sstates.Contract{
		Address: genesis.ParamContractAddress,
		Method:  "approveAdmin",
		Args:    adminBuffer.Bytes(),
	}
```

### SetParam
Administrator set global parameter, is prepare value, won't take effect immediately.

method: setParam

args: smartcontract/service/native/states.Params

#### example
```
    params := new(states.Params)
	params.Version = 0x2d
	var paramList = make([]*states.Param, 3)
	for i := 0; i < 3; i++ {
		paramList[i] = &states.Param{
			Version: 0x2d,
			K:       "key-test" + strconv.Itoa(i) + "-" + value,
			V:       "value-test" + strconv.Itoa(i) + "-" + value,
		}
	}
	params.ParamList = paramList
	paramsBuffer := new(bytes.Buffer)
	if err := params.Serialize(paramsBuffer); err != nil {
		fmt.Println("Serialize params struct error.")
		os.Exit(1)
	}

	contract := &sstates.Contract{
		Address: genesis.ParamContractAddress,
		Method:  "setParam",
		Args:    paramsBuffer.Bytes(),
	}
```

### EnforceParam
Administrator make prepare parameter effective.

method: enforceParam

args: smartcontract/service/native/states.Params

#### example
```
    params := new(states.Params)
	params.Version = 0x2d
	var paramList = make([]*states.Param, 3)
	for i := 0; i < 3; i++ {
		paramList[i] = &states.Param{
			Version: 0x2d,
			K:       "key-test" + strconv.Itoa(i) + "-" + value,
		}
	}
	params.ParamList = paramList
	paramsBuffer := new(bytes.Buffer)
	if err := params.Serialize(paramsBuffer); err != nil {
		fmt.Println("Serialize params struct error.")
		os.Exit(1)
	}

	contract := &sstates.Contract{
		Address: genesis.ParamContractAddress,
		Method:  "enforceParam",
		Args:    paramsBuffer.Bytes(),
	}
```

## How to get a parameter
Call the function "GetGlobalPramValue" to get a global parameter value.

args: smartcontract/service/native.NativeService, the NativeServe instance<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;string, parameter name
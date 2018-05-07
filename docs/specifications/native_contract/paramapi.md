# Native Contract API : Param
* [Introduction](#introduction)
* [Contract Method](#contract-method)
* [How to get a parameter](#how-to-get-a-parameter)

## Introduction
This document describes the global parameter manager native contract used in the ontology network.

## Contract Method

### ParamInit
Initialize the contract, invoked in the genesis block.

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
Transfer the administrator of this contract, should be invoked by administrator.

method: transferAdmin

args: smartcontract/service/native/states.Admin

#### example
```
    var destinationAdmin states.Admin
	address, _ := common.AddressFromBase58("TA4knXiWFZ8K4W3e5fAnoNntdc5G3qMT7C")
	copy(destinationAdmin[:], address[:])
	adminBuffer := new(bytes.Buffer)
	if err := destinationAdmin.Serialize(adminBuffer); err != nil {
		fmt.Println("Serialize admins struct error.")
		os.Exit(1)
	}
	contract := &sstates.Contract{
		Address: genesis.ParamContractAddress,
		Method:  "transferAdmin",
		Args:    adminBuffer.Bytes(),
	}
```

### AcceptAdmin
Accept administrator permission of the contract.

method: acceptAdmin

args: smartcontract/service/native/states.Admin

#### example
```
    var destinationAdmin states.Admin
	address, _ := common.AddressFromBase58("TA4knXiWFZ8K4W3e5fAnoNntdc5G3qMT7C")
	copy(destinationAdmin[:], address[:])
	adminBuffer := new(bytes.Buffer)
	if err := destinationAdmin.Serialize(adminBuffer); err != nil {
		fmt.Println("Serialize admin struct error.")
		os.Exit(1)
	}

	contract := &sstates.Contract{
		Address: genesis.ParamContractAddress,
		Method:  "acceptAdmin",
		Args:    adminBuffer.Bytes(),
	}
```

### SetGlobalParam
Administrator set global parameter, is prepare value, won't take effect immediately.

method: setGlobalParam

args: smartcontract/service/native/states.Params

#### example
```
    params := new(states.Params)
	*params = make(map[string]string)
	for i := 0; i < 3; i++ {
		k := "key-test" + strconv.Itoa(i) + "-" + key
		v := "value-test" + strconv.Itoa(i) + "-" + value
		(*params)[k] = v
	}
	paramsBuffer := new(bytes.Buffer)
	if err := params.Serialize(paramsBuffer); err != nil {
		fmt.Println("Serialize params struct error.")
		os.Exit(1)
	}

	contract := &sstates.Contract{
		Address: genesis.ParamContractAddress,
		Method:  "setGlobalParam",
		Args:    paramsBuffer.Bytes(),
	}
```

### CreateSnapshot
Administrator make prepare parameter effective.

method: createSnapshot

args: nil

#### example
```
    contract := &sstates.Contract{
		Address: genesis.ParamContractAddress,
		Method:  "createSnapshot",
	}
```

## How to get a parameter
Call the function "GetGlobalParam" to get a global parameter value.

args: smartcontract/service/native.NativeService, the NativeServe instance<br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;string, parameter name
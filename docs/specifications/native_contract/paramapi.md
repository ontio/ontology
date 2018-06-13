# Native Contract API : Param
* [Introduction](#introduction)
* [Contract Method](#contract-method)

## Introduction
This document describes the global parameter manager native contract used in the ontology network.

## Contract Method

### ParamInit
Initialize the contract, invoked in the genesis block.

method: init

args: nil

return: bool

#### example
```
    init := sstates.Contract{
		Address: ParamContractAddress,
		Method:  "init",
	}
```
### TransferAdmin
Transfer the administrator of this contract, should be invoked by administrator.

method: transferAdmin

args: smartcontract/service/native/global_params.Admin

return: bool

#### example
```
    var destinationAdmin global_params.Admin
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

args: smartcontract/service/native/global_params.Admin

return: bool

#### example
```
    var destinationAdmin global_params.Admin
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
### SetOperator
Administrator set operator of the contract.

method: setOperator

args: smartcontract/service/native/global_params.Admin

return: bool
#### example
```
    var destinationOperator global_params.Admin
	address, _ := common.AddressFromBase58("TA4knXiWFZ8K4W3e5fAnoNntdc5G3qMT7C")
	copy(destinationOperator[:], address[:])
	adminBuffer := new(bytes.Buffer)
	if err := destinationOperator.Serialize(adminBuffer); err != nil {
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
Operator set global parameter, is prepare value, won't take effect immediately.

method: setGlobalParam

args: smartcontract/service/native/global_params.Params

return: bool

#### example
```
    params := new(global_params.Params)
	*params = make(map[string]string)
	for i := 0; i < 3; i++ {
		k := "key-test" + strconv.Itoa(i) + "-" + key
		v := "value-test" + strconv.Itoa(i) + "-" + value
		(*params) = append(*params, &global_params.Param{k,v})
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

### GetGlobalParam
Get global parameter, the method will return smartcontract/service/native/global_params.Params

method: getGlobalParam

args: smartcontract/service/native/global_params.ParamNameList

return: array

#### example
```
    nameList := new(global_params.ParamNameList)
	for i := 0; i < 3; i++ {
		k := "key-test" + strconv.Itoa(i) + "-" + key
		(*nameList) = append(*nameList, k)
	}
	nameListBuffer := new(bytes.Buffer)
	if err := nameList.Serialize(nameListBuffer); err != nil {
		fmt.Println("Serialize ParamNameList struct error.")
		os.Exit(1)
	}
	contract := &sstates.Contract{
		Address: genesis.ParamContractAddress,
		Method:  "getGlobalParam",
		Args:    nameListBuffer.Bytes(),
	}
```

### CreateSnapshot
Operator make prepare parameter effective.

method: createSnapshot

args: nil

return: bool

#### example
```
    contract := &sstates.Contract{
		Address: genesis.ParamContractAddress,
		Method:  "createSnapshot",
	}
```

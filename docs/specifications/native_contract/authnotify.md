# Authentication contract

common event format is as follows, including txhash, state, gasConsumed and notify, each native contract method have different notifies.

|key|description|
|:--|:--|
|TxHash|transaction hash|
|State|1 indicates success，0 indicates fail|
|GasConsumed|gas fee consumed by this transaction|
|Notify|Notify event|

#### InitContractAdmin

* Usage: Init admin information of a certain contract through authentication contract

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "ContractAddress": "0600000000000000000000000000000000000000", //contract address of authentication contract
      "States":[
        "initContractAdmin", //method name
        "ea1e2adf8c19f5a7e877860264ebf326e8c3aa5a", //contract address of contract which want to achieve authentication control
        "did:ont:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA" //admin ontid if above contract
      ]
    },
    //notify of gas fee transfer
    {
      "ContractAddress": "0200000000000000000000000000000000000000", //ong contract address
      "States":[
        "transfer", //method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance contract address (to)
        10000000 //gas fee amount(decimal: 9)
      ]
    }
  ]
}
```

#### Transfer

* Usage: Transfer admin to another ontid

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "ContractAddress": "0600000000000000000000000000000000000000", //contract address of authentication contract
      "States":[
        "transfer", //method name
        "ea1e2adf8c19f5a7e877860264ebf326e8c3aa5a", //contract address of contract which want to achieve authentication control
        true //status
      ]
    },
    //notify of gas fee transfer
    {
      "ContractAddress": "0200000000000000000000000000000000000000", //ong contract address
      "States":[
        "transfer", //method name
        "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
        "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance contract address (to)
        10000000 //gas fee amount(decimal: 9)
      ]
    }
  ]
}
```


#### AssignFuncsToRole

* Usage: Assign authentication of invoking a function in a certain contract to a role

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "ContractAddress": "0600000000000000000000000000000000000000", //contract address of authentication contract
      "States":[
        "assignFuncsToRole", //method name
        "ea1e2adf8c19f5a7e877860264ebf326e8c3aa5a", //contract address of contract which want to achieve authentication control
        true //status
      ]
    },
     //notify of gas fee transfer
     {
       "ContractAddress": "0200000000000000000000000000000000000000", //ong contract address
       "States":[
         "transfer", //method name
         "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
         "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance contract address (to)
         10000000 //gas fee amount(decimal: 9)
       ]
     }
  ]
}
```

#### AssignOntIDsToRole

* Usage: Assign a role to a certain ontid

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "ContractAddress": "0600000000000000000000000000000000000000", //contract address of authentication contract
      "States":[
        "assignOntIDsToRole", //method name
        "ea1e2adf8c19f5a7e877860264ebf326e8c3aa5a", //contract address of contract which want to achieve authentication control
        true //status
      ]
    },
     //notify of gas fee transfer
     {
       "ContractAddress": "0200000000000000000000000000000000000000", //ong contract address
       "States":[
         "transfer", //method name
         "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
         "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance contract address (to)
         10000000 //gas fee amount(decimal: 9)
       ]
     }
  ]
}
```

#### Delegate

* Usage: delegate authentication to another ontid

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "ContractAddress": "0600000000000000000000000000000000000000", //contract address of authentication contract
      "States":[
        "delegate",// method name
        "ea1e2adf8c19f5a7e877860264ebf326e8c3aa5a", //contract address of contract which want to achieve authentication control
        "did:ont:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //from ontid
        "did:ont:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //to ontid
        true //status
      ]
    },
     //notify of gas fee transfer
     {
       "ContractAddress": "0200000000000000000000000000000000000000", //ong contract address
       "States":[
         "transfer", //method name
         "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
         "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance contract address (to)
         10000000 //gas fee amount(decimal: 9)
       ]
     }
  ]
}
```

#### Withdraw

* Usage: Withdraw delegated authentication

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "ContractAddress": "0600000000000000000000000000000000000000", //contract address of authentication contract
      "States":[
        "withdraw",// method name
        "ea1e2adf8c19f5a7e877860264ebf326e8c3aa5a", //contract address of contract which want to achieve authentication control
        "did:ont:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //from ontid
        "did:ont:AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //to ontid
        true //status
      ]
    },
     //notify of gas fee transfer
     {
       "ContractAddress": "0200000000000000000000000000000000000000", //ong contract address
       "States":[
         "transfer", //method name
         "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
         "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance contract address (to)
         10000000 //gas fee amount(decimal: 9)
       ]
     }
  ]
}
```

#### VerifyToken

* Usage: Verify authentication of ontid

* Event and notify:
```
{
  "TxHash":"",
  "State":1,
  "GasConsumed":10000000,
  "Notify":[
    //notify of the method
    {
      "ContractAddress": "0600000000000000000000000000000000000000", //contract address of authentication contract
      "States":[
        "verifyToken", // method name
        "0700000000000000000000000000000000000000", //contract address of contract which want to achieve authentication control
        "ZGlk0m9uddpBVVhDSnM3NmlqWlUzOHNlUEg5MlNuVWFvZDdQNXRVbUV4", //invoker ontid
        "registerCandidate",// function name want to verify authentication
        true //status
      ]
    },
    //notify of gas fee transfer
    {
       "ContractAddress": "0200000000000000000000000000000000000000", //ong contract address
       "States":[
         "transfer", //method name
         "AbPRaepcpBAFHz9zCj4619qch4Aq5hJARA", //invoker's address (from)
         "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK", //governance contract address (to)
         10000000 //gas fee amount(decimal: 9)
       ]
    }
  ]
}
```
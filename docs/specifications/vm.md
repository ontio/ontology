# Package VM

* [Introduction](#introduction)
* [Definitions](#definitions)
  * [ExecutionEngine](#ExecutionEngine)
  * [OP_CODE(Constants)](#OP_CODE(Constants))
  * [OP_CODE(Flow_control)](#OP_CODE(Flow_control))
  * [OP_CODE(Stack)](#OP_CODE(Stack))
  * [OP_CODE(Splice)](#OP_CODE(Splice))
  * [OP_CODE(Bitwise_logic)](#OP_CODE(Bitwise_logic))
  * [OP_CODE(Arithmetic)](#OP_CODE(Arithmetic))
  * [OP_CODE(Crypto)](#OP_CODE(Crypto))

## Introduction

VM is an execution engine which executes smart contracts write by operation code.

## Definitions

### ExecutionEngine

Field                 | Type                   | Description
--------------------|----------------------|----------------------------------------------------------
crypto               | ICrypto                 |  Crypto interface.
table                | IScriptTable            |  ScriptTable interface.
service              | IApiService             |  ApiService interface.
signable             | ISignableObject         |  SignableObject interface.
invocationStack      | *OpStack                |  stack to load script.
nOpCount             | int                     |  control MAXSTEPS with nOpCount.
Stack                | *OpStack                |  stack to execution script.
altStack             | *OpStack                |  stack to save data when execution script.
State                | VMState                 |  VM state.
opReader             | *VmReader               |  Read format data.
opcode               | OpCode                  |  operation code.


### OP_CODE(Constants)

OP_CODE                           | Value         | Description
-------------------------------|--------------|-----------------------------------------------
OP_0 (OP_FALSE)                   | 0x00          |  An empty array of bytes is pushed onto the stack. (This is not a no-op: an item is added to the stack.)
OP_PUSHBYTES1 ~ OP_PUSHBYTES75    | 0x01 ~ 0x4B   |  0x01-0x4B The next OP_CODE bytes is data to be pushed onto the stack
OP_PUSHDATA1                      | 0x4C          |  The next byte contains the number of bytes to be pushed onto the stack.
OP_PUSHDATA2                      | 0x4D          |  The next two bytes contain the number of bytes to be pushed onto the stack.
OP_PUSHDATA4                      | 0x4E          |  The next four bytes contain the number of bytes to be pushed onto the stack.
OP_1NEGATE                        | 0x4F          |  The number -1 is pushed onto the stack.
OP_1  (OP_TRUE)                   | 0x51          |  The number 1 is pushed onto the stack.
OP_2 ~ OP_16                      | 0x52 ~ 0x60   |  0x52 ~ 0x60 The number OP_CODE is pushed onto the stack.


### OP_CODE(Flow_control)

OP_CODE                           | Value         | Description
-------------------------------|--------------|-----------------------------------------------
OP_NOP                            | 0x61          |  Does nothing.
OP_JMP                            | 0x62          |  
OP_JMPIF                          | 0x63          |  
OP_JMPIFNOT                       | 0x64          |
OP_CALL                           | 0x65          |  
OP_RET                            | 0x66          |  
OP_APPCALL                        | 0x67          |  
OP_SYSCALL                        | 0x68          |  


### OP_CODE(Stack)

OP_CODE                           | Value         | Description
-------------------------------|--------------|-----------------------------------------------
OP_TOALTSTACK                     | 0x6B          |  Puts the input onto the top of the alt stack. Removes it from the main stack.
OP_FROMALTSTACK                   | 0x6C          |  Puts the input onto the top of the main stack. Removes it from the alt stack.
OP_XDROP                          | 0x6D          |
OP_XSWAP                          | 0x72          |  
OP_XTUCK                          | 0x73          |  
OP_DEPTH                          | 0x74          |  Puts the number of stack items onto the stack.
OP_DROP                           | 0x75          |  Removes the top stack item.
OP_DUP                            | 0x76          |  Duplicates the top stack item.
OP_NIP                            | 0x77          |  Removes the second-to-top stack item.
OP_OVER                           | 0x78          |  Copies the second-to-top stack item to the top.
OP_PICK                           | 0x79          |  The item n back in the stack is copied to the top.
OP_ROLL                           | 0x7A          |  The item n back in the stack is moved to the top.
OP_ROT                            | 0x7B          |  The top three items on the stack are rotated to the left.
OP_SWAP                           | 0x7C          |  The top two items on the stack are swapped.
OP_TUCK                           | 0x7D          |  The item at the top of the stack is copied and inserted before the second-to-top item.


### OP_CODE(Splice)

OP_CODE                           | Value         | Description
-------------------------------|--------------|-----------------------------------------------
OP_CAT                             | 0x7E          |  Concatenates two strings.
OP_SUBSTR                          | 0x7F          |  Returns a section of a string.
OP_LEFT                            | 0x80          |  Keeps only characters left of the specified point in a string.
OP_RIGHT                           | 0x81          |  Keeps only characters right of the specified point in a string.
OP_SIZE                            | 0x82          |  Returns the length of the input string.


### OP_CODE(Bitwise_logic)

OP_CODE                           | Value         | Description
-------------------------------|--------------|-----------------------------------------------
OP_INVERT                          | 0x83          |  Flips all of the bits in the input.
OP_AND                             | 0x84          |  Boolean and between each bit in the inputs.
OP_OR                              | 0x85          |  Boolean or between each bit in the inputs.
OP_XOR                             | 0x86          |  Boolean exclusive or between each bit in the inputs.
OP_EQUAL                           | 0x87          |  Returns 1 if the inputs are exactly equal, 0 otherwise.


### OP_CODE(Arithmetic)

OP_CODE                           | Value         | Description
-------------------------------|--------------|-----------------------------------------------
OP_1ADD                            | 0x8B          |  1 is added to the input.
OP_1SUB                            | 0x8C          |  1 is subtracted from the input.
OP_2MUL                            | 0x8D          |  The input is multiplied by 2.
OP_2DIV                            | 0x8E          |  The input is divided by 2.
OP_NEGATE                          | 0x8F          |  The sign of the input is flipped.
OP_ABS                             | 0x90          |  The input is made positive.
OP_NOT                             | 0x91          |  If the input is 0 or 1, it is flipped. Otherwise the output will be 0.
OP_0NOTEQUAL                       | 0x92          |  Returns 0 if the input is 0. 1 otherwise.
OP_ADD                             | 0x93          |  a is added to b.
OP_SUB                             | 0x94          |  b is subtracted from a.
OP_MUL                             | 0x95          |  a is multiplied by b.
OP_DIV                             | 0x96          |  a is divided by b.
OP_MOD                             | 0x97          |  Returns the remainder after dividing a by b.
OP_LSHIFT                          | 0x98          |  Shifts a left b bits, preserving sign.
OP_RSHIFT                          | 0x99          |  Shifts a right b bits, preserving sign.
OP_BOOLAND                         | 0x9A          |  If both a and b are not 0, the output is 1. Otherwise 0.
OP_BOOLOR                          | 0x9B          |  If a or b is not 0, the output is 1. Otherwise 0.
OP_NUMEQUAL                        | 0x9C          |  Returns 1 if the numbers are equal, 0 otherwise.
OP_NUMNOTEQUAL                     | 0x9E          |  Returns 1 if the numbers are not equal, 0 otherwise.
OP_LESSTHAN                        | 0x9F          |  Returns 1 if a is less than b, 0 otherwise.
OP_GREATERTHAN                     | 0xA0          |  Returns 1 if a is greater than b, 0 otherwise.
OP_LESSTHANOREQUAL                 | 0xA1          |  Returns 1 if a is less than or equal to b, 0 otherwise.
OP_GREATERTHANOREQUAL              | 0xA2          |  Returns 1 if a is greater than or equal to b, 0 otherwise.
OP_MIN                             | 0xA3          |  Returns the smaller of a and b.
OP_MAX                             | 0xA4          |  Returns the larger of a and b.
OP_WITHIN                          | 0xA5          |  Returns 1 if x is within the specified range (left-inclusive), 0 otherwise.


### OP_CODE(Crypto)

OP_CODE                           | Value         | Description
-------------------------------|--------------|-----------------------------------------------
OP_SHA1                            | 0xA7          |  The input is hashed using SHA-1.
OP_SHA256                          | 0xA8          |  The input is hashed using SHA-256.
OP_HASH160                         | 0xA9          |  
OP_HASH256                         | 0xAA          |  
OP_CHECKSIG                        | 0xAC          |  The entire transaction's outputs, inputs, and script (from the most recently-executed OP_CODESEPARATOR to the end) are hashed. The signature used by OP_CHECKSIG must be a valid signature for this hash and public key. If it is, 1 is returned, 0 otherwise.
OP_CHECKMULTISIG                   | 0xAE          |  For each signature and public key pair, OP_CHECKSIG is executed. If more public keys than signatures are listed, some key/sig pairs can fail. All signatures need to match a public key. If all signatures are valid, 1 is returned, 0 otherwise.

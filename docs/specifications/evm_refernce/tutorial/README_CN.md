# 本体链 EVM 合约

本开发指南旨在帮助开发者在本体网络上开发部署和测试 EVM 智能合约。

本体网络已通过 Ontology EVM 虚拟机实现了对以太坊生态的兼容。目前本体支持 EVM 合约，并且已支持以太坊链的节点调用方式，开发者可以直接在本体链上使用 Truffle，Remix 和 Web3.js 等 EVM 开发工具。


* [1 开发环境工具简介](#1-开发环境工具简介)
  * [1.1 Remix](#11-Remix)
    * [1.1.1 环境初始化](#111-环境初始化)
    * [1.1.2 编译合约](#112-编译合约)
    * [1.1.3 部署合约](#113-部署合约)
    * [1.1.4 调用合约](#114-调用合约)
  * [1.2 Truffle](#12-Truffle)
    * [1.2.1 安装](#121-安装truffle)
    * [1.2.2 配置 truffle-config](#122-配置-truffle-config)
    * [1.2.3 部署合约到本体链](#123-部署合约到本体链)
  * [1.3 Hardhat](#13-Hardhat)
    * [1.3.1 安装](#131-安装)
    * [1.3.2 配置 hardhat-config](#132-配置-hardhat-config)
    * [1.3.3 部署合约到本体链](#133-部署合约到本体链)
* [2 网络详情](#2-网络详情)
    * [2.1 节点网络](#21-节点网络)
    * [2.2 本体链上的 EVM 资产](#22-本体链上的-EVM-资产)
    * [2.3 OEP-4 资产列表](#23-OEP-4-资产列表)
    * [2.4 手续费](#24-手续费)
* [3 使用 MetaMask 管理密钥](#3-使用-MetaMask-管理密钥)
    * [3.1 初始化 Web3](#31-初始化-Web3)
    * [3.2 获取账户](#32-获取账户)
    * [3.3 合约初始化](#33-合约初始化)
    * [3.4 调用函数](#34-调用函数)
* [4 以太坊资产迁移至本体链](#4-以太坊资产迁移至本体链)
* [5 合约开发流程](#5-合约开发流程)
  * [5.1 环境准备](#51-环境准备)
  * [5.2 合约设计](#52-合约设计)
    * [5.2.1 合约逻辑](#521-合约逻辑)
    * [5.2.2 定义合约事件](#522-定义合约事件)
    * [5.2.3 定义函数](#523-定义函数)
  * [5.3 使用 Hardhat 编译和测试合约](#53-使用-Hardhat-编译和测试合约)
    * [5.3.1 创建 Hardhat 项目](#531-创建-Hardhat-项目)
    * [5.3.2 修改 hardhat.config.js文件](#532-修改hardhat.config.js文件)
    * [5.3.3 文件准备](#533-文件准备)
    * [5.3.4 在 test 文件夹下添加测试代码](#534-在-test-文件夹下添加测试代码)
    * [5.3.5 编译合约](#535-编译合约)
    * [5.3.6 测试合约](#536-测试合约)
* [6 API 参考](#6-API-参考)
  




## 1 开发环境工具简介

EVM合约使用 Solidity 语言开发，详见 [Solidity教程](https://solidity-cn.readthedocs.io/zh/develop/index.html)。开发者也可以复用现有的以太坊合约框架开发和部署EVM合约。

### 1.1 Remix

[Remix IDE](https://remix.ethereum.org/#optimize=false&runs=200&evmVersion=null&version=soljson-v0.8.1+commit.df193b15.js) 是一个开源的 Solidity 合约集成开发环境，支持用户进行合约开发、编译。部署、调试等一系列工作。Remix IDE 的官方文档（英文）请见[此链接](https://remix-ide.readthedocs.io/en/latest/)。

下面我们通过一个 Hello World 合约样例来展示如何使用 Remix。 

#### 1.1.1 环境初始化

首次使用 Remix 需要在 PLUGIN MANAGER 里面找到并添加 Solidity Compiler 和 Deploy and Run Transactions 模块到编译器中。

![image-20210526142630046](./image-20210526142630046.png)

然后选择 Solidity 环境，创建新文件并命名为 HelloWorld.sol，再将已经写好的 [Hello World 合约](../contract-demo/helloworlddemo/helloworld.sol)代码复制到该文件中。

![image-20210526143301031](./image-20210526143301031.png)

#### 1.1.2 编译合约

点击 Solidity Compiler 按钮，选择编译器版本为 0.5.10，开始编译 HelloWorld.sol。

#### 1.1.3 部署合约

编译后可以将合约部署到本体网络中。下面以测试网作为范例。

> **注意：** 在部署合约之前，需要将 MetaMask 钱包连接到本体网络。

首先，在 MetaMask 的网络设置中选择"自定义RPC"，然后输入并保存如下配置信息。
- 网络名：ontology testnet
- 节点url："http://polaris1.ont.io:20339"或"http://polaris2.ont.io:20339"或"http://polaris3.ont.io:20339"
- 输入Chain ID：5851
- 输入区块链浏览器url："https://explorer.ont.io/testnet"

![RemixIDE_Step1](./rpc.png)

之后需要在本体[Faucet地址](https://developer.ont.io/)领取测试 ONG 作为手续费。

最后，在 Remix 环境中选择“Injected Web3”，再点击“Deploy”完成合约部署。

![deploy contract](./remix_deploy.jpg)

#### 1.1.4 调用合约

合约部署后就可以调用合约中的方法了。部署示例中的 Hello World 合约时， `hello` 字符串会存入合约，我们可以调用合约的 `message` 方法来查询这个字符串：

![invoke contract](./remix_invoke.jpg)

### 1.2 Truffle

Truffle 是用于辅助以太坊智能合约开发、测试和管理的框架，官方文档（英文）请参考[此链接](https://www.trufflesuite.com/docs/truffle/quickstart)。

下面我们用这段[测试代码](../contract-demo/truffledemo)作为范例介绍 Truffle 的使用。

#### 1.2.1 安装

开发环境初始化，首先安装 Truffle 环境需要的依赖。

- [Node.js v8+ LTS and npm](https://nodejs.org/en/) (comes with Node)
- [Git](https://git-scm.com/)

然后通过以下命令安装 Truffle。

```shell
npm install -g truffle
```

#### 1.2.2 配置 truffle-config

- 首先创建 `.secret` 来存储测试助记词或者私钥（可在 MetaMask 里面找到）
- 然后按照以下内容修改 truffle-config 文件

```
const HDWalletProvider = require('@truffle/hdwallet-provider');
const fs = require('fs');
const mnemonic = fs.readFileSync(".secret").toString().trim();
module.exports = {
  networks: {
    ontology: {
     provider: () => new HDWalletProvider(mnemonic, `http://polaris2.ont.io:20339`),
     network_id: 5851,
     port: 20339,            // Standard Ethereum port (default: none)
     timeoutBlocks: 200,
     gas:800000,
     skipDryRun: true
    }
  },
  compilers: {
    solc: {
      version: "0.5.16",    // Fetch exact version from solc-bin (default: truffle's version)
      docker: false,        // Use "0.5.1" you've installed locally with docker (default: false)
      settings: {          // See the solidity docs for advice about optimization and evmVersion
       optimizer: {
         enabled: true,
         runs: 200
       },
       evmVersion: "byzantium"
      }
    }
  }
};
```

#### 1.2.3 部署合约到本体网络

执行如下的命令部署合约。

```
truffle migrate --network ontology
```

显示如下输出则代表部署成功。

> **注意：** 编写测试脚本时尽量不要使用以太坊代币的单位（如wei，gwei，ether等）。

```
Compiling your contracts...
===========================
> Everything is up to date, there is nothing to compile.

Starting migrations...
======================
> Network name:    'ontology'
> Network id:      12345
> Block gas limit: 0 (0x0)
1_initial_migration.js
======================

   Replacing 'Migrations'
   ----------------------
   > transaction hash:    0x9019551f3d60611e1bc6b323f3cf3020d15c8aeb06833d14ff864e24622884aa
   > Blocks: 0            Seconds: 4
   > contract address:    0x53e137A51CfD1E1b088E0d921eB5dBCF9cFa955E
   > block number:        6264
   > block timestamp:     1624876467
   > account:             0x4e7946D1Ee8f8703E24C6F3fBf032AD4459c4648
   > balance:             0.00001
   > gas used:            172969 (0x2a3a9)
   > gas price:           0 gwei
   > value sent:          0 ETH
   > total cost:          0 ETH


   > Saving migration to chain.
   > Saving artifacts
   -------------------------------------
   > Total cost:                   0 ETH


2_deploy_migration.js
=====================

   Replacing 'HelloWorld'
   ----------------------
   > transaction hash:    0xf8289b96f2496a8c940ca38d736a554a90f64d927b689921781619499906721b
   > Blocks: 0            Seconds: 4
   > contract address:    0xfbff9bd546B0e0D4b40f6f758847b70050d01b37
   > block number:        6266
   > block timestamp:     1624876479
   > account:             0x4e7946D1Ee8f8703E24C6F3fBf032AD4459c4648
   > balance:             0.00001
   > gas used:            243703 (0x3b7f7)
   > gas price:           0 gwei
   > value sent:          0 ETH
   > total cost:          0 ETH

hello contract address: 0xfbff9bd546B0e0D4b40f6f758847b70050d01b37

   > Saving migration to chain.
   > Saving artifacts
   -------------------------------------
   > Total cost:                   0 ETH


Summary
=======
> Total deployments:   2
> Final cost:          0 ETH
```

### 1.3 Hardhat

Hardhat 是一个编译、部署、测试和调试以太坊应用的开发环境。下面我们用这段[测试代码](../contract-demo/hardhatdemo)作为范例介绍 Hardhat 的使用。

#### 1.3.1 安装

请参考此[安装教程](https://hardhat.org/getting-started/)进行安装。

#### 1.3.2 配置 hardhat-config

- 创建 `.secret` 用于存储测试用户的私钥
- 按照如下代码修改 hardhat.config.js 文件

```
require("@nomiclabs/hardhat-waffle");
const fs = require('fs');
const privateKey = fs.readFileSync(".secret").toString().trim();

module.exports = {
    defaultNetwork: "ontology_testnet",
    networks: {
        hardhat: {},
        ontology_testnet: {
            url: "http://polaris2.ont.io:20339",
            chainId: 5851,
            gasPrice:500,
            gas:2000000,
            timeout:10000000,
            accounts: [privateKey]
        }
    },
    solidity: {
        version: "0.8.0",
        settings: {
            optimizer: {
                enabled: true,
                runs: 200
            }
        }
    },
};
```

#### 1.3.3 部署合约

在项目根目录下执行下面的命令，部署合约到本体测试网。

```
$ npx hardhat run scripts/sample-script.js --network ontology_testnet
```

执行结果

```
sss@sss hardhatdemo % npx hardhat run scripts/sample-script.js --network ontology_testnet
RedPacket deployed to: 0xB105388ac7F019557132eD6eA90fB4BAaFde6E81
```

## 2 网络详情

### 2.1 节点网络

#### 主网信息

| 项目           | 描述                                                                                                                    |
| :------------- | :---------------------------------------------------------------------------------------------------------------------- |
| NetworkName    | Ontology Mainnet                                                                                                        |
| chainId        | 58                                                                                                                      |
| Gas Token      | ONG Token                                                                                                               |
| RPC            | http://dappnode1.ont.io:20339,http://dappnode2.ont.io:20339,http://dappnode3.ont.io:20339,http://dappnode4.ont.io:20339 |
| Block Explorer | https://explorer.ont.io/                                                                                                |

#### 测试网信息

| 项目           | 描述                                                                                                                  |
| :------------- | :-------------------------------------------------------------------------------------------------------------------- |
| NetworkName    | Ontology Testnet                                                                                                      |
| chainId        | 5851                                                                                                                  |
| Gas Token      | ONG Token                                                                                                             |
| RPC            | http://polaris1.ont.io:20339, http://polaris2.ont.io:20339, http://polaris3.ont.io:20339,http://polaris4.ont.io:20339 |
| Block Explorer | https://explorer.ont.io/testnet                                                                                       |

### 2.2 本体链上的 EVM 资产

| 资产名 | 资产地址                                    |
| :----- | :------------------------------------------ |
| ONG    | 0x00000000000000000000000000000000000000000 |

### 2.3 OEP-4 资产列表

请参考[此链接](https://explorer.ont.io/tokens/oep4/10/1#)。

### 2.4 手续费

本体链上的 EVM 合约使用 ONG 作为手续费。可以在[此处](https://developer.ont.io/)领取 ONG 测试币。

## 3 使用 MetaMask 管理密钥

本体网络支持开发者使用 MetaMask 插件来管理以太坊钱包私钥。

MetaMask 是一个非托管的钱包，用户的私钥通过助记词加密并储存在本地浏览器，一旦用户丢失私钥将无法恢复对钱包的使用。MetaMask 通过 Infura 接入以太坊。 请在[这里](https://metamask.io/)了解 MetaMask 的详细信息。

### 3.1 初始化 Web3

第一步，在 dApp 内安装 web3 环境:

   ```
   npm install --save web3
   ```

创建一个新的文件，命名为 `web3.js` ，将以下代码复制到该文件:

   ```js
   import Web3 from 'web3';

const getWeb3 = () => new Promise((resolve) => {
    window.addEventListener('load', () => {
        let currentWeb3;

        if (window.ethereum) {
            currentWeb3 = new Web3(window.ethereum);
            try {
                // Request account access if needed
                window.ethereum.enable();
                // Acccounts now exposed
                resolve(currentWeb3);
            } catch (error) {
                // User denied account access...
                alert('Please allow access for the app to work');
            }
        } else if (window.web3) {
            window.web3 = new Web3(web3.currentProvider);
            // Acccounts always exposed
            resolve(currentWeb3);
        } else {
            console.log('Non-Ethereum browser detected. You should consider trying MetaMask!');
        }
    });
});

export default getWeb3;
   ```

简言之，只要在 Chrome 浏览器上安装了 MetaMask 插件，就可以使用该插件注入的`ethereum`全局变量。

第二步，在你的 client 里引入如下代码,

   ```js
   import getWeb3 from '/path/to/web3';
   ```

调用如下函数:

   ```js
     getWeb3()
    .then((result) => {
        this.web3 = result;// we instantiate our contract next
    });
   ```

### 3.2 设置账户

我们需要从以上创建的 web3 实例中获取一个账户来发送交易。

   ```js
     this.web3.eth.getAccounts()
    .then((accounts) => {
        this.account = accounts[0];
    })
   ```

`getAccounts()` 函数返回用户在 MetaMask 中的所有账户。`accounts[0]`是用户当前选择的账户。

### 3.3 合约初始化

完成以上步骤后，对你的合约进行初始化。

### 3.4 调用函数

现在你可以使用你刚才创建的合约实例调用任何你想调用的函数。需要特别说明的是：

函数`call()` 用来完成合约的预执行操作，例如：

```js
  this.myContractInstance.methods.myMethod(myParams)
    .call()
    .then(
        // do stuff with returned values
    )
```

函数`send()` 用来调用合约来改变合约状态，例如：

```
this.myContractInstance.methods.myMethod(myParams)
.send({
from: this.account,gasPrice: 0
}).then (
(receipt) => {
  // returns a transaction receipt}
);
```

## 4 以太坊资产迁移至本体链

开发者可以使用 [PolyBridge](https://bridge.poly.network/) 工具实现资产跨链。

## 5 合约开发流程

下面我们使用 Hardhat 工具来演示开发部署和测试 EVM 合约的完整流程。

### 5.1 环境准备

- 安装 [nodejs](https://nodejs.org/en/) 

- 安装 [Hardhat](https://hardhat.org/getting-started/)

### 5.2 合约设计

#### 5.2.1 合约逻辑

我们将以一个红包合约为例，该合约主要提供以下功能：

- 发红包
- 领红包

每次发红包需要指定红包金额和该红包数量。例如，红包总金额是100个token，红包的数量是10，也就是有10个不同的地址领取红包。为了简单起见，我们设置每个红包金额相等， 也就是每个地址可以领10个token。

根据以上的逻辑我们可以设置如下的存储结构：

```
EIP20Interface public token; // support token address
uint public nextPacketId; // the next redpacket ID

// packetId -> Packet, store all the redpacket
mapping(uint => Packet) public packets;

//packetId -> address -> bool,  store receive redpacket record
mapping(uint => mapping(address => bool)) public receiveRecords;

struct Packet {
    uint[] assetAmounts;// Number of tokens per copy
    uint receivedIndex; // Number of red packets received
}
```

#### 5.2.2 定义合约事件

在合约执行的过程中，我们可以通过添加事件来追溯合约执行流程。

在本例中我们设计以下两个事件：

-  发红包时，合约会生成红包的 ID,该 ID 要通过事件推送给调用者
-  领取红包时，需要推送一个事件用来记录领取的红包 ID 和 token 数量

```
event SendRedPacket(uint packetId, uint amount); 
event ReceiveRedPacket(uint packetId, uint amount);
```

#### 5.2.3 定义函数

**`sendRedPacket`** 

发红包。任何人都可以调用该接口，将一定量的token打给该合约地址，从而其他的地址可以从该合约地址领取红包。

> **注意：** 在调用该方法之前，需要先授权该合约地址能够从用户的地址把 token 转移走，所以需要先调用该 token 的 `approve`方法。

```
function sendRedPacket(uint amount, uint packetNum) public payable returns (uint) {
    require(amount >= packetNum, "amount >= packetNum");
    require(packetNum > 0 && packetNum < 100, "packetNum>0 && packetNum < 100");
    uint before = token.universalBalanceOf(address(this));
    token.universalTransferFrom(address(msg.sender), address(this), amount);
    uint afterValue = token.universalBalanceOf(address(this));
    uint delta = afterValue - before;
    uint id = nextPacketId;
    uint[] memory assetAmounts = new uint[](packetNum);
    for (uint i = 0; i < packetNum; i++) {
        assetAmounts[i] = delta / packetNum;
    }
    packets[id] = Packet({assetAmounts : assetAmounts, receivedIndex : 0});
    nextPacketId = id + 1;
    emit SendRedPacket(id, amount);
    return id;
}
```

**`receivePacket`** 

领取红包。任何地址都可以通过调用该接口领取红包，调用该接口的时候需要指定红包的 ID，也就是指定要领取哪个红包。

```
function receivePacket(uint packetId) public payable returns (bool) {
    require(packetId < nextPacketId, "not the redpacket");
    Packet memory p = packets[packetId];
    if (p.assetAmounts.length < 1) {
        return false;
    }
    require(p.receivedIndex < p.assetAmounts.length - 1, "It's over");
    require(receiveRecords[packetId][address(msg.sender)] == false, "has received");
    p.receivedIndex = p.receivedIndex + 1;
    bool res = token.universalTransfer(msg.sender, p.assetAmounts[p.receivedIndex]);
    require(res, "token transfer failed");
    packets[packetId] = p;
    receiveRecords[packetId][address(msg.sender)] == true;
    emit ReceiveRedPacket(packetId, p.assetAmounts[p.receivedIndex]);
    return true;
}
```

合约完整的代码请参考[此链接](../contract-demo/hardhatdemo/contracts/Redpacket.sol)。

### 5.3 使用 Hardhat 编译和测试合约

#### 5.3.1 创建 Hardhat 项目

```
mkdir hardhatdemo
cd hardhatdemo
npm init
npm install --save-dev hardhat
npx hardhat
```

#### 5.3.2 修改 hardhat.config.js 文件

添加测试网节点配置信息

```
module.exports = {
    defaultNetwork: "ontology_testnet",
    networks: {
        hardhat: {},
        ontology_testnet: {
            url: "http://polaris2.ont.io:20339",
            chainId: 5851,
            gasPrice:500,
            gas:2000000,
            timeout:10000000,
            accounts: ["your private key1","your private key2"]
        }
    },
    solidity: {
        version: "0.8.0",
        settings: {
            optimizer: {
                enabled: true,
                runs: 200
            }
        }
    },
};
```

`accounts`字段指定的私钥数组，对应的地址需要有测试网的ONG来支付交易的手续费，可以在[这里](https://developer.ont.io/)领取测试网ONG。

#### 5.3.3 文件准备

把红包合约代码文件放到 `contracts`文件夹下，为了支持 ERC-20 代币的转账，我们还需要`EIP20Interface.sol`, `UniversalERC20.sol`, 和 `TokenDemo.sol`文件，可以从[此处](../contract-demo/hardhatdemo/contracts)下载相关文件。

#### 5.3.4 在 test 文件夹下添加测试代码

```
describe("RedPacket", function () {
    let tokenDemo, redPacket, owner, acct1, assetAmount, packetAmount;
    beforeEach(async function () {
        const TokenDemo = await ethers.getContractFactory("TokenDemo");
        tokenDemo = await TokenDemo.deploy(10000000, "L Token", 18, "LT");
        await tokenDemo.deployed();
        const RedPacket = await ethers.getContractFactory("RedPacket");
        redPacket = await RedPacket.deploy(tokenDemo.address);
        await redPacket.deployed();
        [owner, acct1] = await ethers.getSigners();
        assetAmount = 1000;
        packetAmount = 10;
    });
    it("token", async function () {
        expect(await redPacket.token()).to.equal(tokenDemo.address);
    });
    it("sendRedPacket", async function () {
        const approveTx = await tokenDemo.approve(redPacket.address, assetAmount);
        await approveTx.wait();

        const sendRedPacketTx = await redPacket.sendRedPacket(assetAmount, packetAmount);
        await sendRedPacketTx.wait();
        let balance = await tokenDemo.balanceOf(redPacket.address);
        expect(balance.toString()).to.equal(assetAmount.toString());

        res = await redPacket.nextPacketId();
        expect(res.toString()).to.equal("1");

        await redPacket.connect(acct1).receivePacket(0);
        balance = await tokenDemo.balanceOf(acct1.address);
        expect(balance.toString()).to.equal((assetAmount / packetAmount).toString());
    });
});
```

#### 5.3.5 编译合约

在项目根目录执行如下命令编译合约

```
$ npx hardhat compile
Compiling 5 files with 0.8.0
Compilation finished successfully
```

该命令执行完成后会生成如下的文件夹

```
.
├── artifacts
├── cache
├── contracts
├── hardhat.config.js
├── node_modules
├── package-lock.json
├── package.json
├── scripts
└── test
```

#### 5.3.6 测试合约

```
npx hardhat test
```

执行结果如下

```
sss@sss hardhatdemo % npx hardhat test
  RedPacket
    ✓ token
    ✓ sendRedPacket (16159ms)


  2 passing (41s)
```

## 6 API 参考

由于以太坊与本体交易的结构体和存储结构存在差异，目前本体只支持了以太坊部分RPC接口，具体如下：

> **注意：** 本体部分接口返回的内容与以太坊返回的有所不同。

### 方法列表

| 方法名                                                                              | 描述                                                  |
| ----------------------------------------------------------------------------------- | ----------------------------------------------------- |
| [net_version](#net_version)                                                         | 返回当前连接网络的 ID                                 |
| [eth_chainId](#eth_chainId)                                                         | 返回当前链的 chainId                                  |
| [eth_blockNumber](#eth_blockNumber)                                                 | 返回最新块的编号                                      |
| [eth_getBalance](#eth_getBalance)                                                   | 返回指定地址账户的余额                                |
| [eth_protocolVersion](#eth_protocolVersion)                                         | 返回当前以太坊协议的版本                              |
| [eth_syncing](#eth_syncing)                                                         | 返回描述客户端同步状态                                |
| [eth_gasPrice](#eth_gasPrice)                                                       | 返回当前 gas 价格                                     |
| [eth_getStorageAt](#eth_getStorageAt)                                               | 返回指定地址存储位置的值                              |
| [eth_getTransactionCount](#eth_getTransactionCount)                                 | 返回指定地址发生的使用EVM虚拟机交易数量               |
| [eth_getBlockTransactionCountByHash](#eth_getBlockTransactionCountByHash)           | 返回指定块内的使用EVM虚拟机交易数量，使用哈希来指定块 |
| [eth_getBlockTransactionCountByNumber](#eth_getBlockTransactionCountByNumber)       | 返回指定块内的交易数量，使用块编号指定块              |
| [eth_getCode](#eth_getCode)                                                         | 返回指定地址的代码                                    |
| [eth_getTransactionLogs](#eth_getTransactionLogs)                                   | 返回交易执行的日志                                    |
| [eth_sendRawTransaction](#eth_sendRawTransaction)                                   | 为签名交易创建一个新的消息调用交易或合约              |
| [eth_call](#eth_call)                                                               | 立刻执行一个新的消息调用                              |
| [eth_estimateGas](#eth_estimateGas)                                                 | 执行并估算一个交易需要的gas用量                       |
| [eth_getBlockByNumber](#eth_getBlockByNumber)                                       | 返回指定编号的块                                      |
| [eth_getBlockByHash](#eth_getBlockByHash)                                           | 返回具有指定哈希的块                                  |
| [eth_getTransactionByHash](#eth_getTransactionByHash)                               | 返回指定哈希对应的交易                                |
| [eth_getTransactionByBlockHashAndIndex](#eth_getTransactionByBlockHashAndIndex)     | 返回指定块内具有指定索引序号的交易                    |
| [eth_getTransactionByBlockNumberAndIndex](#eth_getTransactionByBlockNumberAndIndex) | 返回指定编号的块内具有指定索引序号的交易              |
| [eth_getTransactionReceipt](#eth_getTransactionReceipt)                             | 返回指定交易的收据，使用哈希指定交易                  |
| [eth_pendingTransactions](#eth_pendingTransactions)                                 | 获取所有处于 pending 状态的交易                       |
| [eth_pendingTransactionsByHash](#eth_pendingTransactionsByHash)                     | 根据交易哈希获取处于 pending 状态的交易               |

### net_version

返回当前连接网络的 ID。

#### 请求参数

无

#### 响应参数

`String`，当前连接网络的 ID
- “1”代表本体主网
- “2”代表本体 Polaris 测试网
- “3”代表 solo 节点


#### 请求示例

```shell
curl -X POST --data '{"jsonrpc":"2.0","method":"net_version","params":[],"id":67}'
```

#### 响应示例

```json
{
  "id": 67,
  "jsonrpc": "2.0",
  "result": "1"
}
```

### eth_chainId

返回当前链的 chainId。

#### 请求参数

无

#### 返回值

`String`，当前链的 chainId

#### 请求示例

```shell
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":83}'
```

#### 响应示例

```json
{
  "id": 83,
  "jsonrpc": "2.0",
  "result": "0x00"
}
```

### eth_blockNumber

返回最新块的编号。

#### 请求参数

无

#### 返回值

`QUANTITY`，整数，节点当前块编号

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":83}'
```

#### 响应示例

```
{
  "id":83,
  "jsonrpc": "2.0",
  "result": "0x4b7" // 1207
}
```

### eth_getBalance

返回指定地址账户的余额。

#### 请求参数
1. `DATA`，20字节，要检查余额的地址
2. `QUANTITY|TAG`，整数块编号，或者字符串`"latest"`，`"earliest"` 或 `"pending"`

```
params: [
   '0x407d73d8a49eeb85d32cf465507dd71d507100c1',
   'latest'
]
```

#### 返回值
`QUANTITY`，当前余额，单位：wei

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":["0x407d73d8a49eeb85d32cf465507dd71d507100c1", "latest"],"id":1}'
```

#### 响应示例

```
{
  "id":1,
  "jsonrpc": "2.0",
  "result": "0x0234c8a3397aab58" // 158972490234375000
}
```

### eth_protocolVersion

返回当前以太坊协议的版本。

#### 请求参数

无

#### 返回值

  `String`，当前的以太坊协议版本

#### 请求示例

```shell
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_protocolVersion","params":[],"id":67}'
```

#### 响应示例

```
{
  "id":67,
  "jsonrpc": "2.0",
  "result": "65"
}
```

### eth_syncing

对于已经同步的客户端，该调用返回一个描述同步状态的对象。

#### 请求参数

无

#### 返回值

  `Object|Boolean`，同步状态对象或false。

  同步对象的结构如下：
 - `startingBlock`: `QUANTITY`，开始块
 - `currentBlock`: `QUANTITY`，当前块，同eth_blockNumber
 - `highestBlock`: `QUANTITY`，预估最高块

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}'
```

#### 响应示例

```
{
  "id":1,
  "jsonrpc": "2.0",
  "result": {
    startingBlock: '0',
    currentBlock: '0x386',
    highestBlock: '0x454'
  }
}
```

### eth_gasPrice

返回当前的gas价格，单位：wei。

#### 请求参数

无

#### 返回值

`QUANTITY`，整数，以wei为单位的当前gas价格

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":73}'
```

#### 响应示例

```
{
  "id":73,
  "jsonrpc": "2.0",
  "result": "0x09184e72a000" // 10000000000000
}
```

### eth_getStorageAt

返回指定地址存储位置的值。

#### 请求参数
  1. `DATA`，20字节，存储地址
  2. `QUANTITY`，存储中的位置号
  3. `QUANTITY|TAG`，整数块号，或字符串"latest"、"earliest" 或"pending"（该参数为无效参数）

#### 返回值

  `DATA`，指定存储位置的值

#### 示例

根据要提取的存储计算正确的位置。考虑下面的合约，由`0x391694e7e0b0cce554cb130d723a9d27458f9298` 部署在地址`0x295a70b2de5e3953354a6a8344e616ed314d7251`：

```
contract Storage {
    uint pos0;
    mapping(address => uint) pos1;

    function Storage() {
        pos0 = 1234;
        pos1[msg.sender] = 5678;
    }
}
```

提取pos0的值比较容易：

```
curl -X POST --data '{"jsonrpc":"2.0", "method": "eth_getStorageAt", "params": ["0x295a70b2de5e3953354a6a8344e616ed314d7251", "0x0", "latest"], "id": 1}' localhost:8545
```

响应结果：

```
{"jsonrpc":"2.0","id":1,"result":"0x00000000000000000000000000000000000000000000000000000000000004d2"}
```

要提取映射表中的元素就难一些了。映射表中元素位置的计算如下：

```
keccack(LeftPad32(key, 0), LeftPad32(map position, 0))
```

这意味着为了提取`pos1["0x391694e7e0b0cce554cb130d723a9d27458f9298"]`的值，我们需要按下面的方式计算位置：

```
keccak(decodeHex("000000000000000000000000391694e7e0b0cce554cb130d723a9d27458f9298" + "0000000000000000000000000000000000000000000000000000000000000001"))
```

geth 控制台自带的 web3 库可以用来进行这个计算：

```
> var key = "000000000000000000000000391694e7e0b0cce554cb130d723a9d27458f9298" + "0000000000000000000000000000000000000000000000000000000000000001"
undefined
> web3.sha3(key, {"encoding": "hex"})
"0x6661e9d6d8b923d5bbaab1b96e1dd51ff6ea2a93520fdc9eb75d059238b8c5e9"
```

现在可以提取指定位置的值了：

```
curl -X POST --data '{"jsonrpc":"2.0", "method": "eth_getStorageAt", "params": ["0x295a70b2de5e3953354a6a8344e616ed314d7251", "0x6661e9d6d8b923d5bbaab1b96e1dd51ff6ea2a93520fdc9eb75d059238b8c5e9", "latest"], "id": 1}' localhost:8545
```

响应结果如下：

```
{"jsonrpc":"2.0","id":1,"result":"0x000000000000000000000000000000000000000000000000000000000000162e"}
```

### eth_getTransactionCount

返回指定地址发生的使用 Ontology EVM 虚拟机交易数量。

#### 请求参数
1. `DATA`: 20字节，地址
2. `QUANTITY|TAG`: 整数块编号，或字符串`"latest"`、`"earliest"`或`"pending"`

```
params: [
   '0x407d73d8a49eeb85d32cf465507dd71d507100c1',
   'latest' // state at the latest block
]
```

#### 返回值

`QUANTITY`，从指定地址发出的交易数量，整数

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionCount","params":["0x407d73d8a49eeb85d32cf465507dd71d507100c1","latest"],"id":1}'
```

#### 响应示例

```
{
  "id":1,
  "jsonrpc": "2.0",
  "result": "0x1" // 1
}
```

### eth_getBlockTransactionCountByHash

返回指定块内的使用 Ontology EVM 虚拟机交易数量，使用哈希来指定块。

#### 请求参数
`DATA`，32字节，块哈希

```
params: [
   '0xb903239f8543d04b5dc1ba6579132b143087c68db1b2168786408fcbce568238'
]
```

#### 返回值

`QUANTITY`，指定块内的交易数量，整数

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockTransactionCountByHash","params":["0xb903239f8543d04b5dc1ba6579132b143087c68db1b2168786408fcbce568238"],"id":1}'
```

#### 响应示例

```
{
  "id":1,
  "jsonrpc": "2.0",
  "result": "0xb" // 11
}
```

### eth_getBlockTransactionCountByNumber

返回指定块内的交易数量，使用块编号指定块。

#### 请求参数

`QUANTITY|TAG`，整数块编号，或字符串`"earliest"`、`"latest"`或`"pending"`

```
params: [
   '0xe8', // 232
]
```

#### 返回值

`QUANTITY`，指定块内的交易数量

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockTransactionCountByNumber","params":["0xe8"],"id":1}'
```

#### 响应示例

```
{
  "id":1,
  "jsonrpc": "2.0",
  "result": "0xa" // 10
}
```

### eth_getCode

返回指定地址的代码。

#### 请求参数
1. `DATA`，20字节，地址
2.  `QUANTITY|TAG`，整数块编号，或字符串`"earliest"`、`"latest"`或`"pending"`（无效参数）

```
params: [
   '0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b',
   '0x2'  // 2
]
```

#### 返回值
`DATA`，指定地址处的代码

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getCode","params":["0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b", "0x2"],"id":1}'
```

#### 响应示例

```
{
  "id":1,
  "jsonrpc": "2.0",
  "result": "0x600160008035811a818181146012578301005b601b6001356025565b8060005260206000f25b600060078202905091905056"
}
```

### eth_getTransactionLogs

返回交易执行的日志。

#### 请求参数

`txHash`，交易哈希

#### 返回值 

返回交易执行日志

#### 请求示例

```
curl -X POST --data '{
     "jsonrpc": "2.0",
     "id": 2233,
     "method": "eth_getTransactionLogs",
     "params": [
       "0x4a9e7c5ec484c1cb854d2831ff51f66f2771e8143362aa75c84f0c6544048fba"
     ]
   }'
```

#### 响应示例

```
{
    "jsonrpc": "2.0",
    "id": 2233,
    "result": [
        {
            "address": "0x9ea0eff7153cebbdd18c2ca3bad818e29e556ba7",
            "topics": [
                "0x7ac369dbd14fa5ea3f473ed67cc9d598964a77501540ba6751eb0b3decf5870d"
            ],
            "data": "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f4ffabb197396c7f48c9cd47ec462b54ed9ce84c",
            "blockNumber": "0x25b",
            "transactionHash": "0x4a9e7c5ec484c1cb854d2831ff51f66f2771e8143362aa75c84f0c6544048fba",
            "transactionIndex": "0x0",
            "blockHash": "0x77abadf9e4ad688212a70260244987f6623b54b56ea737a2cfbc7e7a6344eddc",
            "logIndex": "0x0",
            "removed": false
        }
    ]
}
```

### eth_sendRawTransaction

为签名交易创建一个新的消息调用交易或合约。

#### 请求参数

`DATA`，签名的交易数据

#### 返回值

`DATA`，32字节，交易哈希，如果交易未生效则返回全0哈希。

在交易生效后,创建合约时，使用`eth_getTransactionReceipt`获取合约地址。

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":[{see above}],"id":1}'
```

#### 响应示例

```
{
  "id":1,
  "jsonrpc": "2.0",
  "result": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331"
}
```

### eth_call

立刻执行一个新的消息调用，无需在区块链上创建交易。

#### 请求参数

`Object`，交易调用对象。对象结构如下：
- `from`: `DATA`，20 字节 - 发送交易的原地址，可选
- `to`: `DATA`, 20 字节 - 交易目标地址
- `gas`: `QUANTITY` - 交易可用gas量，可选。eth_call不消耗gas，但是某些执行环节需要这个参数
- `gasPrice`：`QUANTITY` - gas价格，可选
- `value`：`QUANTITY` - 交易发送的以太数量，可选
- `data`: `DATA` - 方法签名和编码参数的哈希，可选
- `QUANTITY|TAG` - 整数块编号，或字符串`"latest"`、`"earliest"`或`"pending"`

#### 返回值

`DATA`，所执行合约的返回值

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{see above}],"id":1}'
```

#### 响应示例

```
{
  "id":1,
  "jsonrpc": "2.0",
  "result": "0x"
}
```

### eth_estimateGas

执行并估算一个交易需要的gas用量。该次交易不会写入区块链。注意，由于多种原因，例如EVM的机制 及节点旳性能，估算的数值可能比实际用量大的多。

#### 请求参数

参考`eth_call`调用的参数，所有的属性都是可选的。如果没有指定gas用量上限，geth将使用挂起块的gas上限。 在这种情况下，返回的gas估算量可能不足以执行实际的交易。

#### 返回值

`QUANTITY`，gas 用量估算值

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_estimateGas","params":[{see above}],"id":1}'
```

#### 响应示例

```
{
  "id":1,
  "jsonrpc": "2.0",
  "result": "0x5208" // 21000
}
```

### eth_getBlockByNumber

返回指定编号的块。

#### 请求参数

1. `QUANTITY|TAG`，整数块编号，或字符串"earliest"、"latest" 或"pending"
2. `Boolean`，为true时返回完整的交易对象，否则仅返回交易哈希

#### 返回值

`Object`，匹配的块对象，如果未找到块则返回 null。对象结构如下：

- `number`: `QUANTITY` - 块编号
- `hash`: `DATA` - 32 字节，块哈希
- `parentHash`: `DATA` - 32 字节，父块的哈希
- `nonce`: `DATA` - 8 字节，空
- `sha3Uncles`: `DATA` - 32 字节，空
- `logsBloom`: `DATA` - 256 字节，空
- `transactionsRoot`: `DATA` - 32 字节，块中的交易树根节点
- `stateRoot`: `DATA` - 32 字节，空
- `receiptsRoot`: `DATA` - 32 字节，空
- `miner`: `DATA` - 20 字节，空
- `difficulty`: `QUANTITY` - 空
- `totalDifficulty`: `QUANTITY` - 空
- `extraData`: `DATA` - 空
- `size`: `QUANTITY` - 本块字节数
- `gasLimit`: `QUANTITY` - 本块允许的最大gas用量
- `gasUsed`: `QUANTITY` - 本块中所有交易使用的总gas用量
- `timestamp`: `QUANTITY` - 块时间戳
- `transactions`: `Array` - 交易对象数组，或32字节长的交易哈希数组
- `uncles`: `Array` - 空

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1b4", true],"id":1}'
```

#### 响应示例

```
{
    "jsonrpc": "2.0",
    "id": 2233,
    "result": {
        "difficulty": "0x0",
        "extraData": "0x",
        "gasLimit": "0x0",
        "gasUsed": "0x0",
        "hash": "0x9e539021092397ec631cbb05fa5418e83b5cccb95dd4663180c243425f01d7b2",
        "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "miner": "0x0000000000000000000000000000000000000000",
        "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "nonce": "0x0000000000000000",
        "number": "0x1b4",
        "parentHash": "0xea06f581bb1e1c4a828f149106e697542bb484627e518ab905a67998d9b670dc",
        "receiptsRoot": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "sha3Uncles": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "size": "0xf2",
        "stateRoot": "0x",
        "timestamp": "0x60c04264",
        "totalDifficulty": "0x0",
        "transactions": [],
        "transactionsRoot": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "uncles": []
    }
}
```

### eth_getBlockByHash

返回具有指定哈希的块。

#### 请求参数
1. `DATA`: 32字节，块哈希
2. `Boolean`: 为true时返回完整的交易对象，否则仅返回交易哈希

#### 返回值

参考`eth_getBlockByNumber`的返回值。

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByHash","params":["0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331", true],"id":1}'
```

#### 响应示例

参考`eth_getBlockByNumber`。

### eth_getTransactionByHash

返回指定哈希对应的交易。

#### 请求参数

`DATA`，32 字节，交易哈希

#### 返回值

`Object` - 交易对象，如果没有找到匹配的交易则返回null。对象结构如下：

- `hash`: `DATA` - 32字节，交易哈希
- `nonce`: `QUANTITY` - 本次交易之前发送方已经生成使用evm虚拟机的交易数量
- `blockHash`: `DATA` - 32字节，交易所在块的哈希，对于挂起块，该值为null
- `blockNumber`: `QUANTITY` - 交易所在块的编号，对于挂起块，该值为null
- `transactionIndex`: `QUANTITY` - 交易在块中的索引位置，挂起块该值为null
- `from`: `DATA` - 20字节，交易发送方地址
- `to`: `DATA` - 20字节，交易接收方地址，对于合约创建交易，该值为null
- `value`: `QUANTITY`- 发送的以太数量，单位：wei
- `gasPrice`: `QUANTITY` - 发送方提供的gas价格，单位：wei
- `gas`: `QUANTITY` - 发送方提供的gas可用量
- `input`: `DATA` - 随交易发送的数据

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByHash","params":["0xb903239f8543d04b5dc1ba6579132b143087c68db1b2168786408fcbce568238"],"id":1}'
```

#### 响应示例

```
{
"id":1,
"jsonrpc":"2.0",
"result": {
    "hash":"0xc6ef2fc5426d6ad6fd9e2a26abeab0aa2411b7ab17f30a99d3cb96aed1d1055b",
    "nonce":"0x",
    "blockHash": "0xbeab0aa2411b7ab17f30a99d3cb9c6ef2fc5426d6ad6fd9e2a26a6aed1d1055b",
    "blockNumber": "0x15df", // 5599
    "transactionIndex":  "0x1", // 1
    "from":"0x407d73d8a49eeb85d32cf465507dd71d507100c1",
    "to":"0x85h43d8a49eeb85d32cf465507dd71d507100c1",
    "value":"0x7f110", // 520464
    "gas": "0x7f110", // 520464
    "gasPrice":"0x09184e72a000",
    "input":"0x603880600c6000396000f300603880600c6000396000f3603880600c6000396000f360",
  }
}
```

### eth_getTransactionByBlockHashAndIndex

返回指定块内具有指定索引序号的交易。

#### 请求参数

1. `DATA`，32字节 - 块哈希
2. `QUANTITY`，交易在块内的索引序号

```
params: [
   '0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331',
   '0x0' // 0
]
```

#### 返回值

查阅`eth_getTransactionByHash`的返回值

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByBlockHashAndIndex","params":["0xc6ef2fc5426d6ad6fd9e2a26abeab0aa2411b7ab17f30a99d3cb96aed1d1055b", "0x0"],"id":1}'
```

返回值请参考`eth_getTransactionByHash`的返回值。

#### eth_getTransactionByBlockNumberAndIndex

返回指定编号的块内具有指定索引序号的交易。

#### 请求参数

1. `QUANTITY|TAG`，整数块编号，或字符串"earliest"、"latest" 或"pending"
2. `QUANTITY`，交易索引序号

```
params: [
   '0x29c', // 668
   '0x0' // 0
]
```

#### 返回值

请参考`eth_getTransactionByHash`的返回值。

#### 请求示例

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByBlockNumberAndIndex","params":["0x29c", "0x0"],"id":1}'
```

响应结果请参考`eth_getTransactionByHash`调用。

### eth_getTransactionReceipt

返回指定交易的收据，使用哈希指定交易。

> **注意：** 挂起的交易其收据无效。

#### 请求参数

`DATA`，32字节，交易哈希

```
params: [
   '0xb903239f8543d04b5dc1ba6579132b143087c68db1b2168786408fcbce568238'
]
```

#### 返回值

`Object` - 交易收据对象，如果收据不存在则为null。对象结构如下：

- `transactionHash`: `DATA`, 32字节 - 交易哈希
- `transactionIndex`: `QUANTITY` - 交易在块内的索引序号
- `blockHash`: `DATA`, 32字节 - 交易所在块的哈希
- `blockNumber`: `QUANTITY` - 交易所在块的编号
- `from`: `DATA`, 20字节 - 交易发送方地址- to: DATA, 20字节 - 交易接收方地址，对于合约创建交易该值为null
- `cumulativeGasUsed`:`QUANTITY` - 交易所在块消耗的gas总量
- `gasUsed`: `QUANTITY` - 该次交易消耗的gas用量
- `contractAddress`: `DATA`, 20字节 - 对于合约创建交易，该值为新创建的合约地址，否则为null
- `logs`: `Array` - 本次交易生成的日志对象数组
- `logsBloom`: `DATA`, 256字节 - bloom过滤器，空
- `status`: `QUANTITY` ，1 (成功) 或 0 (失败)

#### 请求示例

```
curl -X POST --data '{    "jsonrpc": "2.0",
    "id": 16661,
    "method": "eth_getTransactionReceipt",
    "params": [
        "0xe15e2c2240dc58dff54f7c4561a3f784b4ac91cefd0b7cf4dad014fd8a0ad70b"
    ]'
```

#### 响应示例

```
{
    "jsonrpc": "2.0",
    "id": 16661,
    "result": {
        "blockHash": "0x747d2b4599a08c423d50ec772897c992b01b1ac1510d487be52d0167014bd063",
        "blockNumber": "0x204",
        "contractAddress": "0xddcb212ce4896bb02f79db726f6bb8588df41a5c",
        "cumulativeGasUsed": "0x13eecbeb0",
        "from": "0x96216849c49358b10257cb55b28ea603c874b05e",
        "gasUsed": "0x20a86c",
        "logs": [
            {
                "address": "0xddcb212ce4896bb02f79db726f6bb8588df41a5c",
                "topics": [
                    "0x7ac369dbd14fa5ea3f473ed67cc9d598964a77501540ba6751eb0b3decf5870d"
                ],
                "data": "0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000055354e90851d79ee31d8f27d94613cf8f5e7f9e8",
                "blockNumber": "0x204",
                "transactionHash": "0xe15e2c2240dc58dff54f7c4561a3f784b4ac91cefd0b7cf4dad014fd8a0ad70b",
                "transactionIndex": "0x0",
                "blockHash": "0x747d2b4599a08c423d50ec772897c992b01b1ac1510d487be52d0167014bd063",
                "logIndex": "0x0",
                "removed": false
            },
            {
                "address": "0xddcb212ce4896bb02f79db726f6bb8588df41a5c",
                "topics": [
                    "0xedffc32e068c7c95dfd4bdfd5c4d939a084d6b11c4199eac8436ed234d72f926"
                ],
                "data": "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a5b9c59f24caa24ddd9a7ef3aec61bb4908ad984",
                "blockNumber": "0x204",
                "transactionHash": "0xe15e2c2240dc58dff54f7c4561a3f784b4ac91cefd0b7cf4dad014fd8a0ad70b",
                "transactionIndex": "0x0",
                "blockHash": "0x747d2b4599a08c423d50ec772897c992b01b1ac1510d487be52d0167014bd063",
                "logIndex": "0x1",
                "removed": false
            },
            {
                "address": "0xddcb212ce4896bb02f79db726f6bb8588df41a5c",
                "topics": [
                    "0xd604de94d45953f9138079ec1b82d533cb2160c906d1076d1f7ed54befbca97a"
                ],
                "data": "0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000056319fd5a22da14daf19394937dd562619ea34ad",
                "blockNumber": "0x204",
                "transactionHash": "0xe15e2c2240dc58dff54f7c4561a3f784b4ac91cefd0b7cf4dad014fd8a0ad70b",
                "transactionIndex": "0x0",
                "blockHash": "0x747d2b4599a08c423d50ec772897c992b01b1ac1510d487be52d0167014bd063",
                "logIndex": "0x2",
                "removed": false
            }
        ],
        "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "status": "0x1",
        "to": null,
        "transactionHash": "0xe15e2c2240dc58dff54f7c4561a3f784b4ac91cefd0b7cf4dad014fd8a0ad70b",
        "transactionIndex": "0x0"
    }
}
```

### eth_pendingTransactions

获取所有处于pending状态的交易

### eth_pendingTransactionsByHash

根据交易哈希获取处于pending状态的交易

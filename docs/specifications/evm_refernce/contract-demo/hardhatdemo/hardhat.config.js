require("@nomiclabs/hardhat-waffle");

module.exports = {
    defaultNetwork: "ontology_testnet",
    networks: {
        hardhat: {},
        ontology_testnet: {
            url: "http://127.0.0.1:20339",
            chainId: 12345,
            gasPrice:500,
            gas:2000000,
            timeout:10000000,
            accounts: ["59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d","6b98389b8bb98ccaa68876efdcbacc4ae923600023be6b4091e69daa59ba9a9d"]
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


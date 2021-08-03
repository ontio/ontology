const HelloWorld = artifacts.require("HelloWorld");

module.exports = function (deployer, network, accounts) {
    deployer.then(async () => {
        let hello = await deployer.deploy(HelloWorld, "hello ontology!");
        console.log("hello contract address:", hello.address);
    });
}

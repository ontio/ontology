const {expect} = require("chai");

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

//SPDX-License-Identifier: Unlicense
pragma solidity ^0.8.0;

import "hardhat/console.sol";
import "./EIP20Interface.sol";
import "./UniversalERC20.sol";

contract RedPacket {
    using UniversalERC20 for EIP20Interface;
    EIP20Interface public token;
    uint public nextPacketId;

    // packetId -> Packet
    mapping(uint => Packet) public packets;

    //packetId -> address -> bool
    mapping(uint => mapping(address => bool)) public receiveRecords;

    struct Packet {
        uint[] assetAmounts;
        uint receivedIndex;
    }

    event SendRedPacket(uint packetId, uint amount);
    event ReceiveRedPacket(uint packetId, uint amount);
    event Test(uint index, uint data);

    constructor(EIP20Interface token_) {
        token = token_;
    }

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
}

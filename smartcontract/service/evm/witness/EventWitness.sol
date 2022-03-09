// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.10;

contract EventWitness {
    event EventWitnessed(address indexed sender, bytes32 indexed hash);

    function witness(bytes calldata evt) external {
        bytes memory witnessData = abi.encodePacked(msg.sender, evt);
        bytes32 hash = sha256(witnessData);
        emit EventWitnessed(msg.sender, hash);
    }
}


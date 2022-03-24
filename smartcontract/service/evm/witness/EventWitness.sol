// SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

contract EventWitness {
    event EventWitnessed(address indexed sender, bytes32 indexed hash);

    function witness(bytes calldata evt) external {
        // ugly hack: append byte1(0) since we are hashing a merkle leaf. see: ontology/merkle/merkle_hasher.go
        bytes memory leafData = abi.encodePacked(bytes1(0), msg.sender, evt);
        bytes32 merkleLeafHash = sha256(leafData);
        emit EventWitnessed(msg.sender, merkleLeafHash);
    }
}


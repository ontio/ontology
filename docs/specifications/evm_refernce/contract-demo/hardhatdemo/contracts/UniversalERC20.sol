// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./EIP20Interface.sol";


library UniversalERC20 {

    address internal constant ZERO_ADDRESS = address(0x0000000000000000000000000000000000000000);

    function universalTransfer(
        EIP20Interface token,
        address to,
        uint256 amount
    ) internal returns (bool) {
        if (amount == 0) {
            return true;
        }
        if (isETH(token)) {
            payable(to).transfer(amount);
        } else {
            token.transfer(to, amount);
        }
        return true;
    }

    function universalTransferFrom(
        EIP20Interface token,
        address from,
        address to,
        uint256 amount
    ) internal {
        if (amount == 0) {
            return;
        }
        if (isETH(token)) {
            require(from == msg.sender && msg.value >= amount, "Wrong useage of ETH.universalTransferFrom()");
            if (to != address(this)) {
                address payable toAddr = payable(to);
                toAddr.transfer(amount);
            }
            if (msg.value > amount) {
                payable(msg.sender).transfer(msg.value - amount);
            }
        } else {
            token.transferFrom(from, to, amount);
        }
    }

    function universalApprove(
        EIP20Interface token,
        address to,
        uint256 amount
    ) internal {
        if (!isETH(token)) {
            if (amount == 0) {
                token.approve(to, 0);
                return;
            }

            uint256 allowance = token.allowance(address(this), to);
            if (allowance < amount) {
                if (allowance > 0) {
                    token.approve(to, 0);
                }
                token.approve(to, amount);
            }
        }
    }

    function universalBalanceOf(EIP20Interface token, address who) internal view returns (uint256) {
        if (isETH(token)) {
            return who.balance;
        } else {
            return token.balanceOf(who);
        }
    }

    function universalDecimals(EIP20Interface token) internal view returns (uint256) {
        if (isETH(token)) {
            return 18;
        }

        (bool success, bytes memory data) = address(token).staticcall{gas: 10000}(abi.encodeWithSignature("decimals()"));
        if (!success || data.length == 0) {
            (success, data) = address(token).staticcall{gas: 10000}(abi.encodeWithSignature("DECIMALS()"));
        }

        return (success && data.length > 0) ? abi.decode(data, (uint256)) : 18;
    }

    function isETH(EIP20Interface token) internal pure returns (bool) {
        return address(token) == ZERO_ADDRESS;
    }
}

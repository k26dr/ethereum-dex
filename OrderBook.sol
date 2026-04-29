pragma solidity ^0.8.1;

import "../libraries/SafeERC20.sol";

using SafeERC20 for IERC20;

contract OrderBook {
	address public BASE_TOKEN;
	address public QUOTE_TOKEN;
	address public ORDERBOOK_ROUTER;

	constructor(address _baseToken, address _quoteToken, address _router) {
		BASE_TOKEN = _baseToken;
		QUOTE_TOKEN = _quoteToken;
		ORDERBOOK_ROUTER = _router;
	}

        enum Side { BUY, SELL }
        struct Order {
		address user;
                uint baseQuantity;
                uint quoteQuantity;
                Side side;
        }
        mapping(uint => Order) public orders;

	function placeOrder (uint orderId, address user, Side side, uint baseQuantity, uint quoteQuantity) public payable {
		require(msg.sender == ORDERBOOK_ROUTER, "only router can send orders");
		require(baseQuantity > 0 && quoteQuantity > 0, "zero quantity orders not permitted");
		if (side == Side.SELL) {
			if (msg.value > 0) {
				require(BASE_TOKEN == address(0), "base token should be 0x0 when selling ETH");
				require(baseQuantity == msg.value, "mismatch between provided baseQuantity and amount of ETH sent");
			}
			else {
				uint beforeBalance = IERC20(BASE_TOKEN).balanceOf(address(this));
				IERC20(BASE_TOKEN).safeTransferFrom(user, address(this), baseQuantity);
				uint afterBalance = IERC20(BASE_TOKEN).balanceOf(address(this));
				require(afterBalance - beforeBalance == baseQuantity, "token error: tokens that charge transfer fees are not permitted");
			}
		}
		else if (side == Side.BUY) {
			if (msg.value > 0) {
				require(QUOTE_TOKEN == address(0), "quote token should be 0x0 when buying with ETH");
				require(quoteQuantity == msg.value, "mismatch between provided quoteQuantity and amount of ETH sent");
			}
			else {
				uint beforeBalance = IERC20(QUOTE_TOKEN).balanceOf(address(this));
				IERC20(QUOTE_TOKEN).safeTransferFrom(user, address(this), quoteQuantity);
				uint afterBalance = IERC20(QUOTE_TOKEN).balanceOf(address(this));
				require(afterBalance - beforeBalance == quoteQuantity, "token error: tokens that charge transfer fees are not permitted");
			}
		}
                orders[orderId] = Order(user, baseQuantity, quoteQuantity, side);
        }

	function cancelOrder (address user, uint orderId) public {
		require(msg.sender == ORDERBOOK_ROUTER, "only router can send orders");
		Order memory order = orders[orderId];
		require(user == order.user, "users can only cancel their own order");
                delete orders[orderId];
		if (order.side == Side.SELL) {
			if (BASE_TOKEN == address(0)) {
				payable(order.user).transfer(order.baseQuantity);
			} 
			else {
				IERC20(BASE_TOKEN).safeTransfer(user, order.baseQuantity);
			}
		}
		else if (order.side == Side.BUY) {
			if (QUOTE_TOKEN == address(0)) {
				payable(order.user).transfer(order.quoteQuantity);
			} 
			else {
				IERC20(QUOTE_TOKEN).safeTransfer(user, order.quoteQuantity);
			}
		}
        }
	
	function fillOrder (address user, uint orderId, uint baseQuantity) public payable {
		require(msg.sender == ORDERBOOK_ROUTER, "only router can send orders");
                Order memory order = orders[orderId];
		uint quoteQuantity = baseQuantity * order.quoteQuantity / order.baseQuantity;
		if (msg.value > 0) {
			if (order.side == Side.SELL) {
				require(QUOTE_TOKEN == address(0), "quote token should be 0x0");
				require(quoteQuantity == msg.value, "mismatch between quoteQuantity and amount of ETH sent");
			}
			else if (order.side == Side.BUY) {
				require(BASE_TOKEN == address(0), "base token should be 0x0");
				require(baseQuantity == msg.value, "mismatch between provided baseQuantity and amount of ETH sent");
			}
		}
                require(baseQuantity > 0, "zero quantity fills not permitted");
                require(baseQuantity <= order.baseQuantity, "trying to fill more than order size");
                require(quoteQuantity > 0, "calculated quote quantity is zero");
		orders[orderId].baseQuantity -= baseQuantity;
		orders[orderId].quoteQuantity -= quoteQuantity;
		if (orders[orderId].baseQuantity == 0) {
			delete orders[orderId];
		}
		if (order.side == Side.SELL) {
			if (QUOTE_TOKEN == address(0)) {
				payable(order.user).transfer(quoteQuantity);
			}
			else {
				IERC20(QUOTE_TOKEN).safeTransferFrom(user, order.user, quoteQuantity);
			}
			if (BASE_TOKEN == address(0)) {
				payable(user).transfer(baseQuantity);
			}
			else {
				IERC20(BASE_TOKEN).safeTransfer(user, baseQuantity);
			}
		}
		else if (order.side == Side.BUY) {
			if (BASE_TOKEN == address(0)) {
				payable(order.user).transfer(baseQuantity);
			}
			else {
				IERC20(BASE_TOKEN).safeTransferFrom(user, order.user, baseQuantity);
			}
			if (QUOTE_TOKEN == address(0)) {
				payable(user).transfer(quoteQuantity);
			}
			else {
				IERC20(QUOTE_TOKEN).safeTransfer(user, quoteQuantity);
			}
		}
	}

}


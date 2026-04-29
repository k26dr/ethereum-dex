pragma solidity ^0.8.1;

import "./OrderBook.sol";

contract OrderBookFactory {
	mapping(bytes32 => OrderBook) public orderbooks;

	event OrderPlaced(uint orderId, address indexed user, address indexed baseToken, address indexed quoteToken, OrderBook.Side side, uint baseQuantity, uint quoteQuantity);
	event OrderCanceled(uint indexed orderId);
	event OrderFill(uint indexed orderId, uint baseQuantity);

        uint public orderCounter = 0; 
	
	function deploy(address baseToken, address quoteToken) public {
		orderbooks[keccak256(abi.encodePacked(baseToken, quoteToken))] = new OrderBook(baseToken, quoteToken, address(this));
	}

	function placeOrder (uint orderId, address baseToken, address quoteToken, OrderBook.Side side, uint baseQuantity, uint quoteQuantity) public payable {
		(bool success, bytes memory data) = address(orderbooks[keccak256(abi.encodePacked(baseToken, quoteToken))]).call{
		    value: msg.value
		}(abi.encodeWithSignature("placeOrder(uint,address,uint256,uint256,uint256)", ++orderCounter, msg.sender, side, baseQuantity, quoteQuantity));
		require(success);
		emit OrderPlaced(orderId, msg.sender, baseToken, quoteToken, side, baseQuantity, quoteQuantity);
	}

	function cancelOrder (address baseToken, address quoteToken, uint orderId) public {
		orderbooks[keccak256(abi.encodePacked(baseToken, quoteToken))].cancelOrder(msg.sender, orderId);
		emit OrderCanceled(orderId);
	}

	function fillOrder (address baseToken, address quoteToken, uint orderId, uint baseQuantity) public payable {
		(bool success, bytes memory data) = address(orderbooks[keccak256(abi.encodePacked(baseToken, quoteToken))]).call{
		    value: msg.value
		}(abi.encodeWithSignature("fillOrder(address,uint256,uint256)", msg.sender, orderId, baseQuantity));
		require(success);
		emit OrderFill(orderId, baseQuantity);
	}
}

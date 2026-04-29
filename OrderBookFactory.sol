import "./Orderbook.sol"

contract OrderBookFactory {
	mapping(bytes32 => address) public orderbooks;

	event OrderPlaced(uint orderId, address indexed user, address indexed baseToken, address indexed quoteToken, Side side, uint baseQuantity, uint quoteQuantity);
	event OrderCanceled(uint indexed orderId);
	event OrderFill(uint indexed orderId, uint baseQuantity);
	
	function deploy(address baseToken, address quoteToken) {
		orderbooks[keccak256(abi.encodePacked(baseToken, quoteToken))] = OrderBook(baseToken, quoteToken);
	}

	function placeOrder (address baseToken, address quoteToken, Side side, uint baseQuantity, uint quoteQuantity) public payable {
		uint orderId = orderbooks[keccak256(abi.encodePacked(baseToken, quoteToken))].placeOrder(side, baseQuantity, quoteQuantity);
		emit OrderPlaced(orderId, msg.sender, baseToken, quoteToken, side, baseQuantity, quoteQuantity);
	}
}

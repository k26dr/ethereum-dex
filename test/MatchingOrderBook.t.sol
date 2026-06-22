// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../MatchingOrderBook.sol"; 
import "../src/MockERC20.sol"; 

contract MatchingOrderBookTest is Test {
	MatchingOrderBook orderBook;
	MockERC20 USDC;
	MockERC20 EURT;
	address owner = address(0x1);
	address user1 = address(0x2);
	address user2 = address(0x3);
	address user3 = address(0x4);
	address burnAddress = address(0x6);

	function setUp() public {
		USDC = new MockERC20("USDC", "USDC", 6);
		EURT = new MockERC20("EURT", "EURT", 6);
		orderBook = new MatchingOrderBook();
		USDC.mint(user1, 10000e6);
		vm.prank(user1);
		USDC.approve(address(orderBook), 100e18);
	}

	function testPlaceOrderBeforeMarketCreation() public {
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		vm.expectRevert("createMarket before placing an order on it"); 
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 115 * 1e6, 1e6);
	}

	function testCreateMarket() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
	}

	function testCreateMultipleMarketsForSamePair() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		orderBook.createMarket(address(EURT), address(USDC), 10e6, 10e6);
		orderBook.createMarket(address(EURT), address(USDC), 0, 100e6);
	}
	
	function testHackBankFailWithExternalAddress() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		MatchingOrderBook.MarketDetails memory marketDetails = orderBook.getMarketDetails(marketId);
		vm.expectRevert("only owner can withdraw funds");
		Bank(marketDetails.bankAddress).withdrawTo(user1, address(0), 1e18);
	}

	function testCreateMarketTwiceAndFail() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		vm.expectRevert("market has already been created");
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
	}
	
	function testPlaceOneOrderSell() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		EURT.approve(address(orderBook), 100e18);
		EURT.mint(user1, 1000*1e18);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		uint128 orderId = orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 115e6, 1e6);
	        MatchingOrderBook.Order memory order = orderBook.getOrder(marketId, MatchingOrderBook.Side.SELL, orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 115e6, "Base Quantity should match");
		assertEq(order.price, 1e6, "Price should match");
		MatchingOrderBook.MarketDetails memory marketDetails = orderBook.getMarketDetails(marketId);
		uint bankEurtBalance = EURT.balanceOf(marketDetails.bankAddress);
		assertEq(bankEurtBalance, 115e6);
		MatchingOrderBook.Order[] memory orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.SELL, 1);
		assertEq(orders[0].user, user1);
		assertEq(orders[0].baseQuantity, 115e6);
		assertEq(orders[0].price, 1e6);
		assertEq(orders[0].nextOrderId, 0);
	}

	function testPlaceOneOrderBuy() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		USDC.approve(address(orderBook), 100e18);
		USDC.mint(user1, 1000*1e18);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		uint128 orderId = orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 115e6, 1e6);
	        MatchingOrderBook.Order memory order = orderBook.getOrder(marketId, MatchingOrderBook.Side.BUY, orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 115e6, "Base Quantity should match");
		assertEq(order.price, 1e6, "Price should match");
		MatchingOrderBook.MarketDetails memory marketDetails = orderBook.getMarketDetails(marketId);
		uint bankUsdcBalance = USDC.balanceOf(marketDetails.bankAddress);
		assertEq(bankUsdcBalance, 115e6);
		MatchingOrderBook.Order[] memory orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.BUY, 1);
		assertEq(orders[0].user, user1);
		assertEq(orders[0].baseQuantity, 115e6);
		assertEq(orders[0].price, 1e6);
		assertEq(orders[0].nextOrderId, 0);
	}

	function testPlaceTwoOrdersToCheckGas() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		USDC.approve(address(orderBook), 100e18);
		USDC.mint(user1, 1000*1e18);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 115e6, 1e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 115e6, 11e5);
	}

	function testPlace99SellOrders() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		EURT.approve(address(orderBook), 100e18);
		EURT.mint(user1, 1000*1e18);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		for (uint i=1; i < 100; i++) {
			vm.prank(user1);
			orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 115e6, i * 1e6);
		}
		MatchingOrderBook.Order[] memory orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.SELL, 99);
		assertEq(orders[0].previousOrderId, 0, "order pointers are wrong");
		assertEq(orders[0].nextOrderId, 3, "order pointers are wrong");
		assertEq(orders[33].previousOrderId, 34, "order pointers are wrong");
		assertEq(orders[33].nextOrderId, 36, "order pointers are wrong");
	}

	function testPlace100SellOrders() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		EURT.approve(address(orderBook), 100e18);
		EURT.mint(user1, 1000*1e18);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		for (uint i=1; i < 101; i++) {
			vm.prank(user1);
			orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 115e6, i * 1e6);
		}
	}

	function testPlace99BuyOrders() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		USDC.approve(address(orderBook), 100e18);
		USDC.mint(user1, 1000*1e18);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		for (uint i=1; i < 100; i++) {
			vm.prank(user1);
			orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 115e6, i * 1e6);
		}
		MatchingOrderBook.Order[] memory orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.SELL, 99);
		assertNotEq(orders[33].previousOrderId, 0, "previousOrderId pointer is wrong");
		assertNotEq(orders[33].nextOrderId, 0, "nextOrderId pointer is wrong");
	}

	//function testPlace999SellOrders() public {
	//	orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
	//	vm.prank(user1);
	//	EURT.approve(address(orderBook), 100e18);
	//	EURT.mint(user1, 1000*1e18);
	//	bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
	//	for (uint i=1; i < 1000; i++) {
	//		vm.prank(user1);
	//		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 115e6, i * 1e6);
	//	}
	//}

	//function testPlace1000SellOrders() public {
	//	orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
	//	vm.prank(user1);
	//	EURT.approve(address(orderBook), 100e18);
	//	EURT.mint(user1, 1000*1e18);
	//	bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
	//	for (uint i=1; i < 1001; i++) {
	//		vm.prank(user1);
	//		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 115e6, i * 1e6);
	//	}
	//}

	function testFillOneOrder() public {
		// zero all balances to start
		uint user1eurt = EURT.balanceOf(user1);
		uint user1usdc = USDC.balanceOf(user1);
		uint user2eurt = EURT.balanceOf(user2);
		uint user2usdc = USDC.balanceOf(user2);
		if (user1eurt != 0) {
			vm.prank(user1);
			EURT.transfer(burnAddress, user1eurt);
		}
		if (user1usdc != 0) {
			vm.prank(user1);
			USDC.transfer(burnAddress, user1usdc);
		}
		if (user2eurt != 0) {
			vm.prank(user1);
			EURT.transfer(burnAddress, user2eurt);
		}
		if (user2usdc != 0) {
			vm.prank(user1);
			USDC.transfer(burnAddress, user2usdc);
		}

		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		EURT.approve(address(orderBook), 100e18);
		vm.prank(user2);
		USDC.approve(address(orderBook), 100e18);
		EURT.mint(user1, 1000e6);
		USDC.mint(user2, 1000e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 1e6);
		vm.prank(user2);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 1000e6, 1e6);

		// verify post-trade balances
		user1eurt = EURT.balanceOf(user1);
		user1usdc = USDC.balanceOf(user1);
		user2eurt = EURT.balanceOf(user2);
		user2usdc = USDC.balanceOf(user2);
		assertEq(user1eurt, 0);
		assertEq(user1usdc, 1000e6);
		assertEq(user2eurt, 1000e6);
		assertEq(user2usdc, 0);
	}

	function testFill3SellsWithPartialMakerFill() public {
		// zero all balances to start
		uint user1eurt = EURT.balanceOf(user1);
		uint user1usdc = USDC.balanceOf(user1);
		uint user2eurt = EURT.balanceOf(user2);
		uint user2usdc = USDC.balanceOf(user2);
		if (user1eurt != 0) {
			vm.prank(user1);
			EURT.transfer(burnAddress, user1eurt);
		}
		if (user1usdc != 0) {
			vm.prank(user1);
			USDC.transfer(burnAddress, user1usdc);
		}
		if (user2eurt != 0) {
			vm.prank(user1);
			EURT.transfer(burnAddress, user2eurt);
		}
		if (user2usdc != 0) {
			vm.prank(user1);
			USDC.transfer(burnAddress, user2usdc);
		}

		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		USDC.approve(address(orderBook), 100e18);
		vm.prank(user2);
		EURT.approve(address(orderBook), 100e18);
		USDC.mint(user1, 3300e6);
		EURT.mint(user2, 2500e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 1000e6, 1e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 1000e6, 11e5);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 1000e6, 12e5);
		vm.prank(user2);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 2500e6, 1e6);

		// verify post-trade balances
		user1eurt = EURT.balanceOf(user1);
		user1usdc = USDC.balanceOf(user1);
		user2eurt = EURT.balanceOf(user2);
		user2usdc = USDC.balanceOf(user2);
		assertEq(user1eurt, 2500e6);
		assertEq(user1usdc, 0);
		assertEq(user2eurt, 0);
		assertEq(user2usdc, 2800e6);
		MatchingOrderBook.MarketDetails memory marketDetails = orderBook.getMarketDetails(marketId);
		uint bankEurtBalance = EURT.balanceOf(marketDetails.bankAddress);
		uint bankUsdcBalance = USDC.balanceOf(marketDetails.bankAddress);
		assertEq(bankEurtBalance, 0);
		assertEq(bankUsdcBalance, 500e6);
		MatchingOrderBook.Order[] memory orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.BUY, 2);
		assertEq(orders[0].user, user1);
		assertEq(orders[0].baseQuantity, 500e6);
		assertEq(orders[0].price, 1e6);
		assertEq(orders[0].nextOrderId, 0);
		assertEq(orders[1].user, address(0));
		assertEq(orders[1].baseQuantity, 0);
	}

	function testFill3SellsWithLeftover() public {
		// zero all balances to start
		uint user1eurt = EURT.balanceOf(user1);
		uint user1usdc = USDC.balanceOf(user1);
		uint user2eurt = EURT.balanceOf(user2);
		uint user2usdc = USDC.balanceOf(user2);
		if (user1eurt != 0) {
			vm.prank(user1);
			EURT.transfer(burnAddress, user1eurt);
		}
		if (user1usdc != 0) {
			vm.prank(user1);
			USDC.transfer(burnAddress, user1usdc);
		}
		if (user2eurt != 0) {
			vm.prank(user1);
			EURT.transfer(burnAddress, user2eurt);
		}
		if (user2usdc != 0) {
			vm.prank(user1);
			USDC.transfer(burnAddress, user2usdc);
		}

		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		EURT.approve(address(orderBook), 100e18);
		vm.prank(user2);
		USDC.approve(address(orderBook), 100e18);
		EURT.mint(user1, 3000e6);
		USDC.mint(user2, 4200e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 1e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 11e5);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 12e5);
		MatchingOrderBook.Order[] memory orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.SELL, 4);
		assertEq(orders[0].baseQuantity, 1000e6);
		assertEq(orders[0].price, 1e6);
		assertEq(orders[1].baseQuantity, 1000e6);
		assertEq(orders[1].price, 11e5);
		assertEq(orders[2].baseQuantity, 1000e6);
		assertEq(orders[2].price, 12e5);
		assertEq(orders[3].baseQuantity, 0, "order 4 should be empty");
		vm.prank(user2);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 3500e6, 12e5);

		// verify post-trade balances
		user1eurt = EURT.balanceOf(user1);
		user1usdc = USDC.balanceOf(user1);
		user2eurt = EURT.balanceOf(user2);
		user2usdc = USDC.balanceOf(user2);
		assertEq(user1eurt, 0);
		assertEq(user1usdc, 3300e6);
		assertEq(user2eurt, 3000e6);
		assertEq(user2usdc, 300e6, "user2 usdc balance is not right");
		MatchingOrderBook.MarketDetails memory marketDetails = orderBook.getMarketDetails(marketId);
		uint bankEurtBalance = EURT.balanceOf(marketDetails.bankAddress);
		uint bankUsdcBalance = USDC.balanceOf(marketDetails.bankAddress);
		assertEq(bankEurtBalance, 0);
		assertEq(bankUsdcBalance, 600e6, "bank usdc balance is wrong");
		orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.BUY, 1);
		assertEq(orders[0].user, user2);
		assertEq(orders[0].baseQuantity, 500e6);
		assertEq(orders[0].price, 12e5);
		assertEq(orders[0].nextOrderId, 0);
	}

	function testFill3SellsWithPartialFill() public {
		// zero all balances to start
		uint user1eurt = EURT.balanceOf(user1);
		uint user1usdc = USDC.balanceOf(user1);
		uint user2eurt = EURT.balanceOf(user2);
		uint user2usdc = USDC.balanceOf(user2);
		if (user1eurt != 0) {
			vm.prank(user1);
			EURT.transfer(burnAddress, user1eurt);
		}
		if (user1usdc != 0) {
			vm.prank(user1);
			USDC.transfer(burnAddress, user1usdc);
		}
		if (user2eurt != 0) {
			vm.prank(user1);
			EURT.transfer(burnAddress, user2eurt);
		}
		if (user2usdc != 0) {
			vm.prank(user1);
			USDC.transfer(burnAddress, user2usdc);
		}
		
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		MatchingOrderBook.MarketDetails memory marketDetails = orderBook.getMarketDetails(marketId);

		vm.prank(user1);
		EURT.approve(address(orderBook), 100e18);
		vm.prank(user2);
		USDC.approve(address(orderBook), 100e18);
		EURT.mint(user1, 3000e6);
		USDC.mint(user2, 3000e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 1e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 11e5);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 12e5);
		vm.prank(user2);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 2500e6, 12e5);

		// verify post-trade balances
		user1eurt = EURT.balanceOf(user1);
		user1usdc = USDC.balanceOf(user1);
		user2eurt = EURT.balanceOf(user2);
		user2usdc = USDC.balanceOf(user2);
		assertEq(user1eurt, 0);
		assertEq(user1usdc, 2700e6);
		assertEq(user2eurt, 2500e6);
		assertEq(user2usdc, 300e6);
		uint bankEurtBalance = EURT.balanceOf(marketDetails.bankAddress);
		uint bankUsdcBalance = USDC.balanceOf(marketDetails.bankAddress);
		assertEq(bankEurtBalance, 500e6);
		assertEq(bankUsdcBalance, 0);
		MatchingOrderBook.Order[] memory orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.SELL, 1);
		assertEq(orders[0].user, user1);
		assertEq(orders[0].baseQuantity, 500e6);
		assertEq(orders[0].price, 12e5);
		assertEq(orders[0].nextOrderId, 0);
	}

	function testLotsOfOrders() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);

		vm.prank(user1);
		EURT.approve(address(orderBook), 100e18);
		vm.prank(user1);
		USDC.approve(address(orderBook), 100e18);
		vm.prank(user2);
		USDC.approve(address(orderBook), 100e18);
		vm.prank(user2);
		EURT.approve(address(orderBook), 100e18);
		vm.prank(user3);
		EURT.approve(address(orderBook), 100e18);
		vm.prank(user3);
		USDC.approve(address(orderBook), 100e18);
		EURT.mint(user1, 3000e10);
		EURT.mint(user2, 3000e10);
		EURT.mint(user3, 3000e10);
		USDC.mint(user1, 3000e10);
		USDC.mint(user2, 3000e10);
		USDC.mint(user3, 3000e10);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 1e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 11e5);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 12e5);
		vm.prank(user2);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 2500e6, 12e5);
		vm.prank(user2);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 2500e6, 12e5);
		vm.prank(user3);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 2500e6, 11e5);
		vm.prank(user3);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 2500e6, 11e5);
		vm.prank(user2);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 2500e6, 12e5);
	}


	function testCancelOrder() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		EURT.approve(address(orderBook), 300e18);
		EURT.mint(user1, 1000e6);
		vm.prank(user1);
		uint128 orderId = orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 1e6);
	        MatchingOrderBook.Order memory order = orderBook.getOrder(marketId, MatchingOrderBook.Side.SELL, orderId);
		vm.prank(user1);
		orderBook.cancelOrder(marketId, MatchingOrderBook.Side.SELL, orderId);
	        order = orderBook.getOrder(marketId, MatchingOrderBook.Side.SELL, orderId);
		assertEq(order.user, address(0), "User should be deleted");
		assertEq(order.baseQuantity, 0, "Base Quantity should be deleted");
		assertEq(order.price, 0, "Price should be deleted");
	}

	function testDoubleCancelOrderFail() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		EURT.approve(address(orderBook), 300e18);
		EURT.mint(user1, 1000e6);
		vm.prank(user1);
		uint128 orderId = orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 1e6);
	        MatchingOrderBook.Order memory order = orderBook.getOrder(marketId, MatchingOrderBook.Side.SELL, orderId);
		vm.prank(user1);
		orderBook.cancelOrder(marketId, MatchingOrderBook.Side.SELL, orderId);
	        order = orderBook.getOrder(marketId, MatchingOrderBook.Side.SELL, orderId);
		vm.prank(user1);
		vm.expectRevert("users can only cancel their own order / order may not exist");
		orderBook.cancelOrder(marketId, MatchingOrderBook.Side.SELL, orderId);
	}
	
	function testCancelOrderWrongUserFail() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		EURT.approve(address(orderBook), 300e18);
		EURT.mint(user1, 1000e6);
		vm.prank(user1);
		uint128 orderId = orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 1e6);
		vm.prank(user2);
		vm.expectRevert("users can only cancel their own order / order may not exist");
		orderBook.cancelOrder(marketId, MatchingOrderBook.Side.SELL, orderId);
	}

	function testCancelMiddleOrderThenFill() public {
		// zero all balances to start
		uint user1eurt = EURT.balanceOf(user1);
		uint user1usdc = USDC.balanceOf(user1);
		uint user2eurt = EURT.balanceOf(user2);
		uint user2usdc = USDC.balanceOf(user2);
		if (user1eurt != 0) {
			vm.prank(user1);
			EURT.transfer(burnAddress, user1eurt);
		}
		if (user1usdc != 0) {
			vm.prank(user1);
			USDC.transfer(burnAddress, user1usdc);
		}
		if (user2eurt != 0) {
			vm.prank(user2);
			EURT.transfer(burnAddress, user2eurt);
		}
		if (user2usdc != 0) {
			vm.prank(user2);
			USDC.transfer(burnAddress, user2usdc);
		}

		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		MatchingOrderBook.MarketDetails memory marketDetails = orderBook.getMarketDetails(marketId);

		vm.prank(user1);
		EURT.approve(address(orderBook), 300e18);
		EURT.mint(user1, 3000e6);
		vm.prank(user2);
		USDC.approve(address(orderBook), 300e18);
		USDC.mint(user2, 2400e6);
		vm.prank(user1);
		uint128 order1Id = orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 1e6);
		vm.prank(user1);
		uint128 orderIdToCancel = orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 11e5);
		vm.prank(user1);
		uint128 order3Id = orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 12e5);
		vm.prank(user1);
		orderBook.cancelOrder(marketId, MatchingOrderBook.Side.SELL, orderIdToCancel);
		MatchingOrderBook.Order[] memory orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.SELL, 2);
		assertEq(orders[1].price, 12e5);
		assertEq(orders[1].previousOrderId, order1Id);
		assertEq(orders[0].nextOrderId, order3Id);
		vm.prank(user2);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 2000e6, 12e5);

		// verify post-trade balances
		user1eurt = EURT.balanceOf(user1);
		user1usdc = USDC.balanceOf(user1);
		user2eurt = EURT.balanceOf(user2);
		user2usdc = USDC.balanceOf(user2);
		assertEq(user1eurt, 1000e6);
		assertEq(user1usdc, 2200e6, "user1 post-trade usdc balance is wrong");
		assertEq(user2eurt, 2000e6);
		assertEq(user2usdc, 200e6);
		uint bankEurtBalance = EURT.balanceOf(marketDetails.bankAddress);
		uint bankUsdcBalance = USDC.balanceOf(marketDetails.bankAddress);
		assertEq(bankEurtBalance, 0);
		assertEq(bankUsdcBalance, 0);
		orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.SELL, 1);
		assertEq(orders[0].user, address(0));
		assertEq(orders[0].baseQuantity, 0);
		assertEq(orders[0].price, 0);
		assertEq(orders[0].nextOrderId, 0);
		assertEq(orders[0].previousOrderId, 0);
	}

	function testFillOrKillTooSmallFill() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 1000e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 1000e6);
		vm.prank(user1);
		EURT.approve(address(orderBook), 300e18);
		EURT.mint(user1, 1000e6);
		vm.prank(user2);
		USDC.approve(address(orderBook), 300e18);
		USDC.mint(user2, 1000e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 1e6);
		MatchingOrderBook.Order[] memory orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.SELL, 2);
		assertEq(orders[0].user, user1);
		assertEq(orders[0].baseQuantity, 1000e6);
		assertEq(orders[0].previousOrderId, 0);
		assertEq(orders[0].nextOrderId, 0);
		assertEq(orders[1].baseQuantity, 0);
		vm.prank(user2);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 500e6, 1e6);
	}

	function testFillThenTooSmallToPost() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 1000e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 1000e6);
		vm.prank(user1);
		EURT.approve(address(orderBook), 300e18);
		EURT.mint(user1, 1000e6);
		vm.prank(user2);
		USDC.approve(address(orderBook), 300e18);
		USDC.mint(user2, 1100e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 1e6);
		MatchingOrderBook.Order[] memory orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.SELL, 2);
		assertEq(orders[0].user, user1);
		assertEq(orders[0].baseQuantity, 1000e6);
		assertEq(orders[0].previousOrderId, 0);
		assertEq(orders[0].nextOrderId, 0);
		assertEq(orders[1].baseQuantity, 0);
		vm.prank(user2);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 1100e6, 1e6);
		uint user2usdc = USDC.balanceOf(user2);
		assertEq(user2usdc, 100e6, "remainder was not refunded");
	}

	function testInsertBestOffer() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		EURT.approve(address(orderBook), 300e18);
		EURT.mint(user1, 2000e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 1e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 9e5);
		MatchingOrderBook.Order[] memory orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.SELL, 2);
		assertEq(orders[0].price, 9e5);
		assertEq(orders[1].price, 1e6);
	}

	function testInsertBestBid() public {
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		vm.prank(user1);
		EURT.approve(address(orderBook), 300e18);
		EURT.mint(user1, 2000e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 1000e6, 1e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 1000e6, 11e5);
		MatchingOrderBook.Order[] memory orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.BUY, 2);
		assertEq(orders[0].price, 11e5);
		assertEq(orders[1].price, 1e6);
	}

	function testFillDifferentBaseSizes() public {
		// zero all balances to start
		uint user1eurt = EURT.balanceOf(user1);
		uint user1usdc = USDC.balanceOf(user1);
		uint user2eurt = EURT.balanceOf(user2);
		uint user2usdc = USDC.balanceOf(user2);
		if (user1eurt != 0) {
			vm.prank(user1);
			EURT.transfer(burnAddress, user1eurt);
		}
		if (user1usdc != 0) {
			vm.prank(user1);
			USDC.transfer(burnAddress, user1usdc);
		}
		if (user2eurt != 0) {
			vm.prank(user1);
			EURT.transfer(burnAddress, user2eurt);
		}
		if (user2usdc != 0) {
			vm.prank(user1);
			USDC.transfer(burnAddress, user2usdc);
		}
		
		orderBook.createMarket(address(EURT), address(USDC), 0, 10e6);
		bytes32 marketId = orderBook.getMarketId(address(EURT), address(USDC), 0, 10e6);
		MatchingOrderBook.MarketDetails memory marketDetails = orderBook.getMarketDetails(marketId);

		vm.prank(user1);
		EURT.approve(address(orderBook), 100e18);
		vm.prank(user2);
		USDC.approve(address(orderBook), 100e18);
		EURT.mint(user1, 4000e6);
		USDC.mint(user2, 4000e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1000e6, 1e6);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1500e6, 11e5);
		vm.prank(user1);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.SELL, 1200e6, 12e5);
		vm.prank(user2);
		orderBook.placeOrder(marketId, MatchingOrderBook.Side.BUY, 3000e6, 12e5);

		// verify post-trade balances
		user1eurt = EURT.balanceOf(user1);
		user1usdc = USDC.balanceOf(user1);
		user2eurt = EURT.balanceOf(user2);
		user2usdc = USDC.balanceOf(user2);
		assertEq(user1eurt, 300e6);
		assertEq(user1usdc, 3250e6);
		assertEq(user2eurt, 3000e6);
		assertEq(user2usdc, 750e6);
		uint bankEurtBalance = EURT.balanceOf(marketDetails.bankAddress);
		uint bankUsdcBalance = USDC.balanceOf(marketDetails.bankAddress);
		assertEq(bankEurtBalance, 700e6);
		assertEq(bankUsdcBalance, 0);
		MatchingOrderBook.Order[] memory orders = orderBook.getOrderBook(marketId, MatchingOrderBook.Side.SELL, 1);
		assertEq(orders[0].user, user1);
		assertEq(orders[0].baseQuantity, 700e6);
		assertEq(orders[0].price, 12e5);
		assertEq(orders[0].nextOrderId, 0);
	}

}

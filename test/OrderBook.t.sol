// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../OrderBook.sol"; 
import "../libraries/MockERC20.sol"; 

contract OrderBookTest is Test {
	OrderBook orderBook;
	MockERC20 USDC;
	MockERC20 EURT;
	address owner = address(0x1);
	address user1 = address(0x2);
	address user2 = address(0x3);

	function setUp() public {
		USDC = new MockERC20("USDC", "USDC", 6);
		EURT = new MockERC20("EURT", "EURT", 6);
		orderBook = new OrderBook();
		USDC.mint(user1, 10000e6);
		vm.prank(user1);
		USDC.approve(address(orderBook), 100e18);
	}

	function testPlaceOrderBeforeMarketCreation() public {
		vm.expectRevert("createMarket before placing an order on it"); 
		vm.prank(user1);
		orderBook.placeOrder(address(EURT), address(USDC), OrderBook.Side.SELL, 115 * 1e6, 115 * 1e6);
	}

	function testCreateMarket() public {
		orderBook.createMarket(address(EURT), address(USDC));
	}
	
	function testBankWithdrawToFail() public {
		orderBook.createMarket(address(EURT), address(USDC));
		address payable bank = orderBook.getBankAddress(address(EURT), address(USDC));
		vm.expectRevert("only owner can withdraw funds");
		Bank(bank).withdrawTo(user1, address(0), 1e18);
	}

	function testCreateMarketTwiceAndFail() public {
		orderBook.createMarket(address(USDC), address(EURT));
		vm.expectRevert("market has already been created");
		orderBook.createMarket(address(USDC), address(EURT));
	}
	
	function testPlaceOrder() public {
		orderBook.createMarket(address(USDC), address(EURT));
		vm.prank(user1);
		USDC.approve(address(orderBook), 100e18);
		USDC.mint(user1, 1000*1e18);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 1e6);
	        OrderBook.Order memory order = orderBook.getorder(address(USDC), address(EURT), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 100 * 1e18, "Base Quantity should match");
		assertEq(order.price, 1e6, "Price should match");
		address payable bank = orderBook.getBankAddress(address(USDC), address(EURT));
		uint bankUsdcBalance = USDC.balanceOf(bank);
		assertEq(bankUsdcBalance, 100e18);
	}

	function testPlaceETHOrder() public {
		orderBook.createMarket(address(0), address(USDC));
		vm.deal(user1, 2e18);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder{value: 1e18}(address(0), address(USDC), OrderBook.Side.SELL, 1e18, 2150e6);
	        OrderBook.Order memory order = orderBook.getorder(address(0), address(USDC), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 1 * 1e18, "Base Quantity should match");
		assertEq(order.price, 2150 * 1e6, "Price should match");
	}

	function testPlace2ETHOrdersForGasDifference() public {
		orderBook.createMarket(address(0), address(USDC));
		vm.deal(user1, 2e18);
		vm.prank(user1);
		orderBook.placeOrder{value: 1e18}(address(0), address(USDC), OrderBook.Side.SELL, 1e18, 2150e6);
		orderBook.placeOrder{value: 1e18}(address(0), address(USDC), OrderBook.Side.SELL, 1e18, 2150e6);
	}

	function testPlace2OrdersForGasDifference() public {
		orderBook.createMarket(address(USDC), address(EURT));
		vm.prank(user1);
		USDC.approve(address(orderBook), 300e18);
		USDC.mint(user1, 1000*1e18);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 87e4);
	        OrderBook.Order memory order = orderBook.getorder(address(USDC), address(EURT), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 100 * 1e18, "Base Quantity should match");
		assertEq(order.price, 87e4, "Price should match");
		vm.prank(user1);
		orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 88e4);
	}

	function testCancelOrder() public {
		orderBook.createMarket(address(USDC), address(EURT));
		vm.prank(user1);
		USDC.approve(address(orderBook), 300e18);
		USDC.mint(user1, 1000*1e18);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 1e6);
	        OrderBook.Order memory order = orderBook.getorder(address(USDC), address(EURT), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 100 * 1e18, "Base Quantity should match");
		assertEq(order.price, 1e6, "Price should match");
		vm.prank(user1);
		orderBook.cancelOrder(address(USDC), address(EURT), orderId);
	        order = orderBook.getorder(address(USDC), address(EURT), orderId);
		assertEq(order.user, address(0), "User should be deleted");
		assertEq(order.baseQuantity, 0, "Base Quantity should be deleted");
		assertEq(order.price, 0, "Price should be deleted");
	}

	function testDoubleCancelOrderFail() public {
		orderBook.createMarket(address(USDC), address(EURT));
		vm.prank(user1);
		USDC.approve(address(orderBook), 300e18);
		USDC.mint(user1, 1000*1e18);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 1e6);
	        OrderBook.Order memory order = orderBook.getorder(address(USDC), address(EURT), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 100 * 1e18, "Base Quantity should match");
		assertEq(order.price, 1e6, "Price should match");
		vm.prank(user1);
		orderBook.cancelOrder(address(USDC), address(EURT), orderId);
	        order = orderBook.getorder(address(USDC), address(EURT), orderId);
		assertEq(order.user, address(0), "User should be deleted");
		assertEq(order.baseQuantity, 0, "Base Quantity should be deleted");
		assertEq(order.price, 0, "Price should be deleted");
		vm.expectRevert("users can only cancel their own order / order may not exist");
		orderBook.cancelOrder(address(USDC), address(EURT), orderId);
	}
	
	function testCancelOrderWrongUserFail() public {
		orderBook.createMarket(address(USDC), address(EURT));
		vm.prank(user1);
		USDC.approve(address(orderBook), 300e18);
		USDC.mint(user1, 1000*1e18);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 1e6);
	        OrderBook.Order memory order = orderBook.getorder(address(USDC), address(EURT), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 100 * 1e18, "Base Quantity should match");
		assertEq(order.price, 1e6, "Price should match");
		vm.prank(user2);
		vm.expectRevert("users can only cancel their own order / order may not exist");
		orderBook.cancelOrder(address(USDC), address(EURT), orderId);
	}

	function testPartialFillOrder() public {
		orderBook.createMarket(address(EURT), address(USDC));
		vm.prank(user1);
		EURT.approve(address(orderBook), 300e6);
		EURT.mint(user1, 1000*1e6);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(EURT), address(USDC), OrderBook.Side.SELL, 115 * 1e6, 115e4);
	        OrderBook.Order memory order = orderBook.getorder(address(EURT), address(USDC), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 115 * 1e6, "Base Quantity should match");
		assertEq(order.price, 115e4, "Price should match");
		vm.prank(user2);
		USDC.approve(address(orderBook), 400e6);
		USDC.mint(user2, 1000e6);
		vm.prank(user2);
		orderBook.fillOrder(orderId, address(EURT), address(USDC), 100e6);
	        order = orderBook.getorder(address(EURT), address(USDC), orderId);
		assertEq(order.user, user1, "User should be user who placed order");
		assertEq(order.baseQuantity, 15e6, "Remaining Base Quantity is wrong");
		assertEq(order.price, 115e4, "Price shouldn't change");
	}

	function testBalances() public {
		// burn all coins
		vm.prank(user1);
		assertTrue(EURT.transfer(address(1), EURT.balanceOf(user1)), "EURT burn transfer user1 failed");
		assertEq(USDC.balanceOf(user1), 100e8);
		uint USDCuser1bal = USDC.balanceOf(user1);
		vm.prank(user1);
		assertTrue(USDC.transfer(address(1), USDCuser1bal), "USDC burn transfer user1 failed");
		vm.prank(user2);
		assertTrue(EURT.transfer(address(1), EURT.balanceOf(user2)), "EURT burn transfer user2 failed");
		vm.prank(user2);
		assertTrue(USDC.transfer(address(1), USDC.balanceOf(user2)), "USDC burn transfer user2 failed");
		assertEq(EURT.balanceOf(user2), 0);
		assertEq(EURT.balanceOf(user1), 0);
		assertEq(USDC.balanceOf(user1), 0);
		assertEq(USDC.balanceOf(user2), 0);

		orderBook.createMarket(address(EURT), address(USDC));
		vm.prank(user1);
		EURT.approve(address(orderBook), 300e6);
		EURT.mint(user1, 1000*1e6);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(EURT), address(USDC), OrderBook.Side.SELL, 100e6, 1e6);
		vm.prank(user2);
		USDC.approve(address(orderBook), 400e6);
		USDC.mint(user2, 1000e6);
		vm.prank(user2);
		orderBook.fillOrder(orderId, address(EURT), address(USDC), 60e6);
		address bankAddress = orderBook.getBankAddress(address(EURT), address(USDC));
		assertEq(EURT.balanceOf(user2), 60e6);
		assertEq(EURT.balanceOf(user1), 900e6);
		assertEq(EURT.balanceOf(bankAddress), 40e6);
		assertEq(USDC.balanceOf(user1), 60e6);
		assertEq(USDC.balanceOf(user2), 940e6);
		assertEq(USDC.balanceOf(bankAddress), 0);
	}

	function testPartialThenFullFill() public {
		orderBook.createMarket(address(EURT), address(USDC));
		vm.prank(user1);
		EURT.approve(address(orderBook), 300e6);
		EURT.mint(user1, 1000*1e6);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(EURT), address(USDC), OrderBook.Side.SELL, 115 * 1e6, 1e6);
	        OrderBook.Order memory order = orderBook.getorder(address(EURT), address(USDC), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 115 * 1e6, "Base Quantity should match");
		assertEq(order.price, 1e6, "Price should match");
		vm.prank(user2);
		USDC.approve(address(orderBook), 400e6);
		USDC.mint(user2, 1000e6);
		vm.prank(user2);
		orderBook.fillOrder(orderId, address(EURT), address(USDC), 100e6);
	        order = orderBook.getorder(address(EURT), address(USDC), orderId);
		assertEq(order.user, user1, "User should be user who placed order");
		assertEq(order.baseQuantity, 15e6, "Remaining Base Quantity is wrong");
		assertEq(order.price, 1e6, "Price shouldn't change");

		vm.prank(user2);
		orderBook.fillOrder(orderId, address(EURT), address(USDC), 15e6);
	        order = orderBook.getorder(address(EURT), address(USDC), orderId);
		assertEq(order.user, address(0), "User should be deleted after order fill");
		assertEq(order.baseQuantity, 0, "Remaining Base Quantity is wrong");
		assertEq(order.price, 0, "Order should be deleted after fill");
	}

	function testFullFillOrder() public {
		orderBook.createMarket(address(EURT), address(USDC));
		vm.prank(user1);
		EURT.approve(address(orderBook), 300e6);
		EURT.mint(user1, 1000*1e6);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(EURT), address(USDC), OrderBook.Side.SELL, 115 * 1e6, 1e6);
	        OrderBook.Order memory order = orderBook.getorder(address(EURT), address(USDC), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 115 * 1e6, "Base Quantity should match");
		assertEq(order.price, 1e6, "Price should match");
		vm.prank(user2);
		USDC.approve(address(orderBook), 400e6);
		USDC.mint(user2, 1000e6);
		vm.prank(user2);
		orderBook.fillOrder(orderId, address(EURT), address(USDC), 115e6);
	        order = orderBook.getorder(address(EURT), address(USDC), orderId);
		assertEq(order.user, address(0), "Order should be deleted after fill");
		assertEq(order.baseQuantity, 0, "Order should be deleted after fill");
		assertEq(order.price, 0, "Order should be deleted after fill");
	}

	function testFillTooMuch() public {
		orderBook.createMarket(address(EURT), address(USDC));
		vm.prank(user1);
		EURT.approve(address(orderBook), 300e6);
		EURT.mint(user1, 1000*1e6);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(EURT), address(USDC), OrderBook.Side.SELL, 115 * 1e6, 1e6);
	        OrderBook.Order memory order = orderBook.getorder(address(EURT), address(USDC), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 115 * 1e6, "Base Quantity should match");
		assertEq(order.price, 1e6, "Price should match");
		vm.prank(user2);
		USDC.approve(address(orderBook), 400e6);
		USDC.mint(user2, 1000e6);
		vm.prank(user2);
		vm.expectRevert("trying to fill more than order size");
		orderBook.fillOrder(orderId, address(EURT), address(USDC), 120e6);
	}

	function testPartialFillThenOverfill() public {
		orderBook.createMarket(address(EURT), address(USDC));
		vm.prank(user1);
		EURT.approve(address(orderBook), 300e6);
		EURT.mint(user1, 1000*1e6);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(EURT), address(USDC), OrderBook.Side.SELL, 115 * 1e6, 1e6);
	        OrderBook.Order memory order = orderBook.getorder(address(EURT), address(USDC), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 115 * 1e6, "Base Quantity should match");
		assertEq(order.price, 1e6, "Price should match");
		vm.prank(user2);
		USDC.approve(address(orderBook), 400e6);
		USDC.mint(user2, 1000e6);
		vm.prank(user2);
		orderBook.fillOrder(orderId, address(EURT), address(USDC), 100e6);
	        order = orderBook.getorder(address(EURT), address(USDC), orderId);
		assertEq(order.user, user1, "User should be user who placed order");
		assertEq(order.baseQuantity, 15e6, "Remaining Base Quantity is wrong");
		assertEq(order.price, 1e6, "Remaining Quote Quantity is wrong");
		vm.prank(user2);
		vm.expectRevert("trying to fill more than order size");
		orderBook.fillOrder(orderId, address(EURT), address(USDC), 20e6);
	}
	
}

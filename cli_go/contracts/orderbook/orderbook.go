// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package orderbook

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// OrderBookOrder is an auto generated low-level Go binding around an user-defined struct.
type OrderBookOrder struct {
	User         common.Address
	BaseQuantity *big.Int
	Price        *big.Int
	Side         uint8
}

// OrderBookMetaData contains all meta data concerning the OrderBook contract.
var OrderBookMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"cancelOrder\",\"inputs\":[{\"name\":\"baseToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"quoteToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"createMarket\",\"inputs\":[{\"name\":\"baseToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"quoteToken\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"fillOrder\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"baseToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"quoteToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"baseQuantityToFill\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"getBankAddress\",\"inputs\":[{\"name\":\"baseToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"quoteToken\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"addresspayable\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getorder\",\"inputs\":[{\"name\":\"baseToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"quoteToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structOrderBook.Order\",\"components\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"baseQuantity\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"side\",\"type\":\"uint8\",\"internalType\":\"enumOrderBook.Side\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"orderCounter\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"placeOrder\",\"inputs\":[{\"name\":\"baseToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"quoteToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"side\",\"type\":\"uint8\",\"internalType\":\"enumOrderBook.Side\"},{\"name\":\"baseQuantity\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"payable\"},{\"type\":\"event\",\"name\":\"OrderCanceled\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OrderFill\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"baseQuantity\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OrderPlaced\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"baseToken\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"quoteToken\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"markethash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"side\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumOrderBook.Side\"},{\"name\":\"baseQuantity\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"price\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]}]",
}

// OrderBookABI is the input ABI used to generate the binding from.
// Deprecated: Use OrderBookMetaData.ABI instead.
var OrderBookABI = OrderBookMetaData.ABI

// OrderBook is an auto generated Go binding around an Ethereum contract.
type OrderBook struct {
	OrderBookCaller     // Read-only binding to the contract
	OrderBookTransactor // Write-only binding to the contract
	OrderBookFilterer   // Log filterer for contract events
}

// OrderBookCaller is an auto generated read-only Go binding around an Ethereum contract.
type OrderBookCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OrderBookTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OrderBookTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OrderBookFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OrderBookFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OrderBookSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OrderBookSession struct {
	Contract     *OrderBook        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OrderBookCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OrderBookCallerSession struct {
	Contract *OrderBookCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// OrderBookTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OrderBookTransactorSession struct {
	Contract     *OrderBookTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// OrderBookRaw is an auto generated low-level Go binding around an Ethereum contract.
type OrderBookRaw struct {
	Contract *OrderBook // Generic contract binding to access the raw methods on
}

// OrderBookCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OrderBookCallerRaw struct {
	Contract *OrderBookCaller // Generic read-only contract binding to access the raw methods on
}

// OrderBookTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OrderBookTransactorRaw struct {
	Contract *OrderBookTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOrderBook creates a new instance of OrderBook, bound to a specific deployed contract.
func NewOrderBook(address common.Address, backend bind.ContractBackend) (*OrderBook, error) {
	contract, err := bindOrderBook(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OrderBook{OrderBookCaller: OrderBookCaller{contract: contract}, OrderBookTransactor: OrderBookTransactor{contract: contract}, OrderBookFilterer: OrderBookFilterer{contract: contract}}, nil
}

// NewOrderBookCaller creates a new read-only instance of OrderBook, bound to a specific deployed contract.
func NewOrderBookCaller(address common.Address, caller bind.ContractCaller) (*OrderBookCaller, error) {
	contract, err := bindOrderBook(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OrderBookCaller{contract: contract}, nil
}

// NewOrderBookTransactor creates a new write-only instance of OrderBook, bound to a specific deployed contract.
func NewOrderBookTransactor(address common.Address, transactor bind.ContractTransactor) (*OrderBookTransactor, error) {
	contract, err := bindOrderBook(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OrderBookTransactor{contract: contract}, nil
}

// NewOrderBookFilterer creates a new log filterer instance of OrderBook, bound to a specific deployed contract.
func NewOrderBookFilterer(address common.Address, filterer bind.ContractFilterer) (*OrderBookFilterer, error) {
	contract, err := bindOrderBook(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OrderBookFilterer{contract: contract}, nil
}

// bindOrderBook binds a generic wrapper to an already deployed contract.
func bindOrderBook(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OrderBookMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OrderBook *OrderBookRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OrderBook.Contract.OrderBookCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OrderBook *OrderBookRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OrderBook.Contract.OrderBookTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OrderBook *OrderBookRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OrderBook.Contract.OrderBookTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OrderBook *OrderBookCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OrderBook.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OrderBook *OrderBookTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OrderBook.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OrderBook *OrderBookTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OrderBook.Contract.contract.Transact(opts, method, params...)
}

// GetBankAddress is a free data retrieval call binding the contract method 0x6ae6adb6.
//
// Solidity: function getBankAddress(address baseToken, address quoteToken) view returns(address)
func (_OrderBook *OrderBookCaller) GetBankAddress(opts *bind.CallOpts, baseToken common.Address, quoteToken common.Address) (common.Address, error) {
	var out []interface{}
	err := _OrderBook.contract.Call(opts, &out, "getBankAddress", baseToken, quoteToken)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetBankAddress is a free data retrieval call binding the contract method 0x6ae6adb6.
//
// Solidity: function getBankAddress(address baseToken, address quoteToken) view returns(address)
func (_OrderBook *OrderBookSession) GetBankAddress(baseToken common.Address, quoteToken common.Address) (common.Address, error) {
	return _OrderBook.Contract.GetBankAddress(&_OrderBook.CallOpts, baseToken, quoteToken)
}

// GetBankAddress is a free data retrieval call binding the contract method 0x6ae6adb6.
//
// Solidity: function getBankAddress(address baseToken, address quoteToken) view returns(address)
func (_OrderBook *OrderBookCallerSession) GetBankAddress(baseToken common.Address, quoteToken common.Address) (common.Address, error) {
	return _OrderBook.Contract.GetBankAddress(&_OrderBook.CallOpts, baseToken, quoteToken)
}

// Getorder is a free data retrieval call binding the contract method 0xaede72ff.
//
// Solidity: function getorder(address baseToken, address quoteToken, uint256 orderId) view returns((address,uint256,uint256,uint8))
func (_OrderBook *OrderBookCaller) Getorder(opts *bind.CallOpts, baseToken common.Address, quoteToken common.Address, orderId *big.Int) (OrderBookOrder, error) {
	var out []interface{}
	err := _OrderBook.contract.Call(opts, &out, "getorder", baseToken, quoteToken, orderId)

	if err != nil {
		return *new(OrderBookOrder), err
	}

	out0 := *abi.ConvertType(out[0], new(OrderBookOrder)).(*OrderBookOrder)

	return out0, err

}

// Getorder is a free data retrieval call binding the contract method 0xaede72ff.
//
// Solidity: function getorder(address baseToken, address quoteToken, uint256 orderId) view returns((address,uint256,uint256,uint8))
func (_OrderBook *OrderBookSession) Getorder(baseToken common.Address, quoteToken common.Address, orderId *big.Int) (OrderBookOrder, error) {
	return _OrderBook.Contract.Getorder(&_OrderBook.CallOpts, baseToken, quoteToken, orderId)
}

// Getorder is a free data retrieval call binding the contract method 0xaede72ff.
//
// Solidity: function getorder(address baseToken, address quoteToken, uint256 orderId) view returns((address,uint256,uint256,uint8))
func (_OrderBook *OrderBookCallerSession) Getorder(baseToken common.Address, quoteToken common.Address, orderId *big.Int) (OrderBookOrder, error) {
	return _OrderBook.Contract.Getorder(&_OrderBook.CallOpts, baseToken, quoteToken, orderId)
}

// OrderCounter is a free data retrieval call binding the contract method 0xb789bf52.
//
// Solidity: function orderCounter() view returns(uint256)
func (_OrderBook *OrderBookCaller) OrderCounter(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OrderBook.contract.Call(opts, &out, "orderCounter")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// OrderCounter is a free data retrieval call binding the contract method 0xb789bf52.
//
// Solidity: function orderCounter() view returns(uint256)
func (_OrderBook *OrderBookSession) OrderCounter() (*big.Int, error) {
	return _OrderBook.Contract.OrderCounter(&_OrderBook.CallOpts)
}

// OrderCounter is a free data retrieval call binding the contract method 0xb789bf52.
//
// Solidity: function orderCounter() view returns(uint256)
func (_OrderBook *OrderBookCallerSession) OrderCounter() (*big.Int, error) {
	return _OrderBook.Contract.OrderCounter(&_OrderBook.CallOpts)
}

// CancelOrder is a paid mutator transaction binding the contract method 0x9a3bc1d7.
//
// Solidity: function cancelOrder(address baseToken, address quoteToken, uint256 orderId) returns()
func (_OrderBook *OrderBookTransactor) CancelOrder(opts *bind.TransactOpts, baseToken common.Address, quoteToken common.Address, orderId *big.Int) (*types.Transaction, error) {
	return _OrderBook.contract.Transact(opts, "cancelOrder", baseToken, quoteToken, orderId)
}

// CancelOrder is a paid mutator transaction binding the contract method 0x9a3bc1d7.
//
// Solidity: function cancelOrder(address baseToken, address quoteToken, uint256 orderId) returns()
func (_OrderBook *OrderBookSession) CancelOrder(baseToken common.Address, quoteToken common.Address, orderId *big.Int) (*types.Transaction, error) {
	return _OrderBook.Contract.CancelOrder(&_OrderBook.TransactOpts, baseToken, quoteToken, orderId)
}

// CancelOrder is a paid mutator transaction binding the contract method 0x9a3bc1d7.
//
// Solidity: function cancelOrder(address baseToken, address quoteToken, uint256 orderId) returns()
func (_OrderBook *OrderBookTransactorSession) CancelOrder(baseToken common.Address, quoteToken common.Address, orderId *big.Int) (*types.Transaction, error) {
	return _OrderBook.Contract.CancelOrder(&_OrderBook.TransactOpts, baseToken, quoteToken, orderId)
}

// CreateMarket is a paid mutator transaction binding the contract method 0x207fd126.
//
// Solidity: function createMarket(address baseToken, address quoteToken) returns()
func (_OrderBook *OrderBookTransactor) CreateMarket(opts *bind.TransactOpts, baseToken common.Address, quoteToken common.Address) (*types.Transaction, error) {
	return _OrderBook.contract.Transact(opts, "createMarket", baseToken, quoteToken)
}

// CreateMarket is a paid mutator transaction binding the contract method 0x207fd126.
//
// Solidity: function createMarket(address baseToken, address quoteToken) returns()
func (_OrderBook *OrderBookSession) CreateMarket(baseToken common.Address, quoteToken common.Address) (*types.Transaction, error) {
	return _OrderBook.Contract.CreateMarket(&_OrderBook.TransactOpts, baseToken, quoteToken)
}

// CreateMarket is a paid mutator transaction binding the contract method 0x207fd126.
//
// Solidity: function createMarket(address baseToken, address quoteToken) returns()
func (_OrderBook *OrderBookTransactorSession) CreateMarket(baseToken common.Address, quoteToken common.Address) (*types.Transaction, error) {
	return _OrderBook.Contract.CreateMarket(&_OrderBook.TransactOpts, baseToken, quoteToken)
}

// FillOrder is a paid mutator transaction binding the contract method 0x590f6f68.
//
// Solidity: function fillOrder(uint256 orderId, address baseToken, address quoteToken, uint256 baseQuantityToFill) payable returns()
func (_OrderBook *OrderBookTransactor) FillOrder(opts *bind.TransactOpts, orderId *big.Int, baseToken common.Address, quoteToken common.Address, baseQuantityToFill *big.Int) (*types.Transaction, error) {
	return _OrderBook.contract.Transact(opts, "fillOrder", orderId, baseToken, quoteToken, baseQuantityToFill)
}

// FillOrder is a paid mutator transaction binding the contract method 0x590f6f68.
//
// Solidity: function fillOrder(uint256 orderId, address baseToken, address quoteToken, uint256 baseQuantityToFill) payable returns()
func (_OrderBook *OrderBookSession) FillOrder(orderId *big.Int, baseToken common.Address, quoteToken common.Address, baseQuantityToFill *big.Int) (*types.Transaction, error) {
	return _OrderBook.Contract.FillOrder(&_OrderBook.TransactOpts, orderId, baseToken, quoteToken, baseQuantityToFill)
}

// FillOrder is a paid mutator transaction binding the contract method 0x590f6f68.
//
// Solidity: function fillOrder(uint256 orderId, address baseToken, address quoteToken, uint256 baseQuantityToFill) payable returns()
func (_OrderBook *OrderBookTransactorSession) FillOrder(orderId *big.Int, baseToken common.Address, quoteToken common.Address, baseQuantityToFill *big.Int) (*types.Transaction, error) {
	return _OrderBook.Contract.FillOrder(&_OrderBook.TransactOpts, orderId, baseToken, quoteToken, baseQuantityToFill)
}

// PlaceOrder is a paid mutator transaction binding the contract method 0xeebe0f3d.
//
// Solidity: function placeOrder(address baseToken, address quoteToken, uint8 side, uint256 baseQuantity, uint256 price) payable returns(uint256 orderId)
func (_OrderBook *OrderBookTransactor) PlaceOrder(opts *bind.TransactOpts, baseToken common.Address, quoteToken common.Address, side uint8, baseQuantity *big.Int, price *big.Int) (*types.Transaction, error) {
	return _OrderBook.contract.Transact(opts, "placeOrder", baseToken, quoteToken, side, baseQuantity, price)
}

// PlaceOrder is a paid mutator transaction binding the contract method 0xeebe0f3d.
//
// Solidity: function placeOrder(address baseToken, address quoteToken, uint8 side, uint256 baseQuantity, uint256 price) payable returns(uint256 orderId)
func (_OrderBook *OrderBookSession) PlaceOrder(baseToken common.Address, quoteToken common.Address, side uint8, baseQuantity *big.Int, price *big.Int) (*types.Transaction, error) {
	return _OrderBook.Contract.PlaceOrder(&_OrderBook.TransactOpts, baseToken, quoteToken, side, baseQuantity, price)
}

// PlaceOrder is a paid mutator transaction binding the contract method 0xeebe0f3d.
//
// Solidity: function placeOrder(address baseToken, address quoteToken, uint8 side, uint256 baseQuantity, uint256 price) payable returns(uint256 orderId)
func (_OrderBook *OrderBookTransactorSession) PlaceOrder(baseToken common.Address, quoteToken common.Address, side uint8, baseQuantity *big.Int, price *big.Int) (*types.Transaction, error) {
	return _OrderBook.Contract.PlaceOrder(&_OrderBook.TransactOpts, baseToken, quoteToken, side, baseQuantity, price)
}

// OrderBookOrderCanceledIterator is returned from FilterOrderCanceled and is used to iterate over the raw logs and unpacked data for OrderCanceled events raised by the OrderBook contract.
type OrderBookOrderCanceledIterator struct {
	Event *OrderBookOrderCanceled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OrderBookOrderCanceledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrderBookOrderCanceled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OrderBookOrderCanceled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OrderBookOrderCanceledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrderBookOrderCanceledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrderBookOrderCanceled represents a OrderCanceled event raised by the OrderBook contract.
type OrderBookOrderCanceled struct {
	OrderId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterOrderCanceled is a free log retrieval operation binding the contract event 0xc41f4ceb2938876c35e61b705e9d2f18a02c4a26ce5e049a6308a943d46851b3.
//
// Solidity: event OrderCanceled(uint256 indexed orderId)
func (_OrderBook *OrderBookFilterer) FilterOrderCanceled(opts *bind.FilterOpts, orderId []*big.Int) (*OrderBookOrderCanceledIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _OrderBook.contract.FilterLogs(opts, "OrderCanceled", orderIdRule)
	if err != nil {
		return nil, err
	}
	return &OrderBookOrderCanceledIterator{contract: _OrderBook.contract, event: "OrderCanceled", logs: logs, sub: sub}, nil
}

// WatchOrderCanceled is a free log subscription operation binding the contract event 0xc41f4ceb2938876c35e61b705e9d2f18a02c4a26ce5e049a6308a943d46851b3.
//
// Solidity: event OrderCanceled(uint256 indexed orderId)
func (_OrderBook *OrderBookFilterer) WatchOrderCanceled(opts *bind.WatchOpts, sink chan<- *OrderBookOrderCanceled, orderId []*big.Int) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _OrderBook.contract.WatchLogs(opts, "OrderCanceled", orderIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrderBookOrderCanceled)
				if err := _OrderBook.contract.UnpackLog(event, "OrderCanceled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOrderCanceled is a log parse operation binding the contract event 0xc41f4ceb2938876c35e61b705e9d2f18a02c4a26ce5e049a6308a943d46851b3.
//
// Solidity: event OrderCanceled(uint256 indexed orderId)
func (_OrderBook *OrderBookFilterer) ParseOrderCanceled(log types.Log) (*OrderBookOrderCanceled, error) {
	event := new(OrderBookOrderCanceled)
	if err := _OrderBook.contract.UnpackLog(event, "OrderCanceled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrderBookOrderFillIterator is returned from FilterOrderFill and is used to iterate over the raw logs and unpacked data for OrderFill events raised by the OrderBook contract.
type OrderBookOrderFillIterator struct {
	Event *OrderBookOrderFill // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OrderBookOrderFillIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrderBookOrderFill)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OrderBookOrderFill)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OrderBookOrderFillIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrderBookOrderFillIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrderBookOrderFill represents a OrderFill event raised by the OrderBook contract.
type OrderBookOrderFill struct {
	OrderId      *big.Int
	BaseQuantity *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterOrderFill is a free log retrieval operation binding the contract event 0x6faa89edce0f05e917d8a9d83dec3380c7664b1540c810e63d2d37304fece236.
//
// Solidity: event OrderFill(uint256 indexed orderId, uint256 baseQuantity)
func (_OrderBook *OrderBookFilterer) FilterOrderFill(opts *bind.FilterOpts, orderId []*big.Int) (*OrderBookOrderFillIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _OrderBook.contract.FilterLogs(opts, "OrderFill", orderIdRule)
	if err != nil {
		return nil, err
	}
	return &OrderBookOrderFillIterator{contract: _OrderBook.contract, event: "OrderFill", logs: logs, sub: sub}, nil
}

// WatchOrderFill is a free log subscription operation binding the contract event 0x6faa89edce0f05e917d8a9d83dec3380c7664b1540c810e63d2d37304fece236.
//
// Solidity: event OrderFill(uint256 indexed orderId, uint256 baseQuantity)
func (_OrderBook *OrderBookFilterer) WatchOrderFill(opts *bind.WatchOpts, sink chan<- *OrderBookOrderFill, orderId []*big.Int) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _OrderBook.contract.WatchLogs(opts, "OrderFill", orderIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrderBookOrderFill)
				if err := _OrderBook.contract.UnpackLog(event, "OrderFill", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOrderFill is a log parse operation binding the contract event 0x6faa89edce0f05e917d8a9d83dec3380c7664b1540c810e63d2d37304fece236.
//
// Solidity: event OrderFill(uint256 indexed orderId, uint256 baseQuantity)
func (_OrderBook *OrderBookFilterer) ParseOrderFill(log types.Log) (*OrderBookOrderFill, error) {
	event := new(OrderBookOrderFill)
	if err := _OrderBook.contract.UnpackLog(event, "OrderFill", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrderBookOrderPlacedIterator is returned from FilterOrderPlaced and is used to iterate over the raw logs and unpacked data for OrderPlaced events raised by the OrderBook contract.
type OrderBookOrderPlacedIterator struct {
	Event *OrderBookOrderPlaced // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OrderBookOrderPlacedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrderBookOrderPlaced)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OrderBookOrderPlaced)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OrderBookOrderPlacedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrderBookOrderPlacedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrderBookOrderPlaced represents a OrderPlaced event raised by the OrderBook contract.
type OrderBookOrderPlaced struct {
	OrderId      *big.Int
	User         common.Address
	BaseToken    common.Address
	QuoteToken   common.Address
	Markethash   [32]byte
	Side         uint8
	BaseQuantity *big.Int
	Price        *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterOrderPlaced is a free log retrieval operation binding the contract event 0x94a398e9c4c7cdd1b360f86f5759cd9e2c94d2991c809b56003158d30626ec17.
//
// Solidity: event OrderPlaced(uint256 indexed orderId, address indexed user, address baseToken, address quoteToken, bytes32 indexed markethash, uint8 side, uint256 baseQuantity, uint256 price)
func (_OrderBook *OrderBookFilterer) FilterOrderPlaced(opts *bind.FilterOpts, orderId []*big.Int, user []common.Address, markethash [][32]byte) (*OrderBookOrderPlacedIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	var markethashRule []interface{}
	for _, markethashItem := range markethash {
		markethashRule = append(markethashRule, markethashItem)
	}

	logs, sub, err := _OrderBook.contract.FilterLogs(opts, "OrderPlaced", orderIdRule, userRule, markethashRule)
	if err != nil {
		return nil, err
	}
	return &OrderBookOrderPlacedIterator{contract: _OrderBook.contract, event: "OrderPlaced", logs: logs, sub: sub}, nil
}

// WatchOrderPlaced is a free log subscription operation binding the contract event 0x94a398e9c4c7cdd1b360f86f5759cd9e2c94d2991c809b56003158d30626ec17.
//
// Solidity: event OrderPlaced(uint256 indexed orderId, address indexed user, address baseToken, address quoteToken, bytes32 indexed markethash, uint8 side, uint256 baseQuantity, uint256 price)
func (_OrderBook *OrderBookFilterer) WatchOrderPlaced(opts *bind.WatchOpts, sink chan<- *OrderBookOrderPlaced, orderId []*big.Int, user []common.Address, markethash [][32]byte) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	var markethashRule []interface{}
	for _, markethashItem := range markethash {
		markethashRule = append(markethashRule, markethashItem)
	}

	logs, sub, err := _OrderBook.contract.WatchLogs(opts, "OrderPlaced", orderIdRule, userRule, markethashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrderBookOrderPlaced)
				if err := _OrderBook.contract.UnpackLog(event, "OrderPlaced", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOrderPlaced is a log parse operation binding the contract event 0x94a398e9c4c7cdd1b360f86f5759cd9e2c94d2991c809b56003158d30626ec17.
//
// Solidity: event OrderPlaced(uint256 indexed orderId, address indexed user, address baseToken, address quoteToken, bytes32 indexed markethash, uint8 side, uint256 baseQuantity, uint256 price)
func (_OrderBook *OrderBookFilterer) ParseOrderPlaced(log types.Log) (*OrderBookOrderPlaced, error) {
	event := new(OrderBookOrderPlaced)
	if err := _OrderBook.contract.UnpackLog(event, "OrderPlaced", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

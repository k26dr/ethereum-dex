// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;
 
import {Script} from "forge-std/Script.sol";
import {MatchingOrderBook} from "../MatchingOrderBook.sol";
 
contract DeployScript is Script {
    function run() public {
        vm.startBroadcast();
        
        MatchingOrderBook book = new MatchingOrderBook();
        
        vm.stopBroadcast();
    }
}

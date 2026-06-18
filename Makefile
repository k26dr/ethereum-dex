CLI_DIR := cli_go
CLI_BIN := ./dex
ANVIL_STATE_FILE := .anvil.env

ANVIL_HOST ?= 127.0.0.1
ANVIL_PORT ?= 8545
ANVIL_RPC_URL ?= http://$(ANVIL_HOST):$(ANVIL_PORT)
ANVIL_CHAIN_ID ?= 31337
ANVIL_MNEMONIC ?= test test test test test test test test test test test junk
ANVIL_BALANCE ?= 10000
ANVIL_PRIVATE_KEY ?= 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
ANVIL_DEPLOYER_ADDRESS ?= 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266

TOKEN_A_NAME ?= TokenA
TOKEN_A_SYMBOL ?= TKA
TOKEN_B_NAME ?= TokenB
TOKEN_B_SYMBOL ?= TKB
TOKEN_DECIMALS ?= 18
MINT_AMOUNT ?= 1000000000000000000000000

MARKET_WALLET ?= $(ANVIL_DEPLOYER_ADDRESS)
DEX_WALLET_PASSWORD ?= dev-password
NEW_WALLET_PASSWORD ?= dev
ETH_FUND_AMOUNT ?= 1000000000000000000
FORGE_CREATE_FLAGS ?= --broadcast
ANVIL_RUN_DIR := .anvil
ANVIL_LOG_FILE := $(ANVIL_RUN_DIR)/anvil.log
ANVIL_PID_FILE := $(ANVIL_RUN_DIR)/anvil.pid

-include $(ANVIL_STATE_FILE)

.PHONY: help anvil check-tools check-rpc ensure-anvil stop terminate init-state build-cli build_cli build-forge deploy deploy-and-configure-cli import-deployer-wallet create-contract create-market create-tokens mint add-account testnet show-state

help:
	@echo "Anvil + DEX local workflow"
	@echo ""
	@echo "make anvil                         # start local Anvil node"
	@echo "make build_cli                     # rebuild CLI binary (alias of build-cli)"
	@echo "make deploy                        # deploy OrderBook and persist ORDERBOOK_ADDRESS"
	@echo "make deploy-and-configure-cli      # build Go CLI, deploy, set rpc/chain-id/contract in CLI config"
	@echo "make create-contract               # create a market for TOKEN_A/TOKEN_B via CLI"
	@echo "make create-tokens                 # deploy two mock ERC20s and persist token addresses"
	@echo "make mint TO=0x...                 # mint both tokens to TO"
	@echo "make add-account                   # create CLI wallet, fund ETH, mint both tokens, persist wallet address"
	@echo "make testnet                        # full local testnet flow using a newly created CLI wallet"
	@echo "make stop                          # stop background anvil started by Makefile"
	@echo "make terminate                     # stop anvil and remove local anvil env/state for a clean start"
	@echo "make show-state                    # print persisted .anvil.env values"

check-tools:
	@command -v anvil >/dev/null || { echo "anvil not found in PATH"; exit 1; }
	@command -v forge >/dev/null || { echo "forge not found in PATH"; exit 1; }
	@command -v cast >/dev/null || { echo "cast not found in PATH"; exit 1; }

check-rpc: check-tools
	@cast chain-id --rpc-url "$(ANVIL_RPC_URL)" >/dev/null 2>&1 || { \
		echo "RPC not reachable at $(ANVIL_RPC_URL)"; \
		echo "Start Anvil first with: make anvil"; \
		echo "Or run the full setup flow with: make testnet"; \
		exit 1; \
	}

init-state:
	@touch $(ANVIL_STATE_FILE)
	@grep -q '^ANVIL_RPC_URL=' $(ANVIL_STATE_FILE) || echo "ANVIL_RPC_URL=$(ANVIL_RPC_URL)" >> $(ANVIL_STATE_FILE)
	@grep -q '^ANVIL_CHAIN_ID=' $(ANVIL_STATE_FILE) || echo "ANVIL_CHAIN_ID=$(ANVIL_CHAIN_ID)" >> $(ANVIL_STATE_FILE)
	@grep -q '^ANVIL_PRIVATE_KEY=' $(ANVIL_STATE_FILE) || echo "ANVIL_PRIVATE_KEY=$(ANVIL_PRIVATE_KEY)" >> $(ANVIL_STATE_FILE)
	@grep -q '^ANVIL_DEPLOYER_ADDRESS=' $(ANVIL_STATE_FILE) || echo "ANVIL_DEPLOYER_ADDRESS=$(ANVIL_DEPLOYER_ADDRESS)" >> $(ANVIL_STATE_FILE)

anvil: check-tools
	@anvil --host $(ANVIL_HOST) --port $(ANVIL_PORT) --chain-id $(ANVIL_CHAIN_ID) --mnemonic "$(ANVIL_MNEMONIC)" --balance $(ANVIL_BALANCE)

ensure-anvil: check-tools
	@mkdir -p $(ANVIL_RUN_DIR)
	@if cast chain-id --rpc-url "$(ANVIL_RPC_URL)" >/dev/null 2>&1; then \
		echo "Anvil already running at $(ANVIL_RPC_URL)"; \
	else \
		nohup anvil --host $(ANVIL_HOST) --port $(ANVIL_PORT) --chain-id $(ANVIL_CHAIN_ID) --mnemonic "$(ANVIL_MNEMONIC)" --balance $(ANVIL_BALANCE) > "$(ANVIL_LOG_FILE)" 2>&1 & \
		echo $$! > "$(ANVIL_PID_FILE)"; \
		echo "Started Anvil in background (pid=$$(cat "$(ANVIL_PID_FILE)"))"; \
		sleep 1; \
	fi
	@$(MAKE) check-rpc

stop:
	@if [ -f "$(ANVIL_PID_FILE)" ]; then \
		pid="$$(cat "$(ANVIL_PID_FILE)")"; \
		if kill "$$pid" >/dev/null 2>&1; then \
			echo "Stopped Anvil pid $$pid"; \
		else \
			echo "Anvil process $$pid was not running"; \
		fi; \
		rm -f "$(ANVIL_PID_FILE)"; \
	else \
		echo "No $(ANVIL_PID_FILE) found"; \
	fi

terminate: stop
	@rm -f "$(ANVIL_STATE_FILE)"
	@rm -f "$(ANVIL_LOG_FILE)"
	@rm -f "$(ANVIL_PID_FILE)"
	@rmdir "$(ANVIL_RUN_DIR)" 2>/dev/null || true
	@echo "Cleaned local state: $(ANVIL_STATE_FILE), $(ANVIL_RUN_DIR)"

build-cli:
	@cd $(CLI_DIR) && go build -o ../$(CLI_BIN) .
	@echo "Built CLI: $(CLI_BIN)"

build_cli: build-cli

build-forge: check-tools
	@forge clean
	@forge build

deploy: check-tools ensure-anvil build-forge init-state
	@set -eu; \
	deploy_output="$$(forge create ./OrderBook.sol:OrderBook $(FORGE_CREATE_FLAGS) --rpc-url "$(ANVIL_RPC_URL)" --private-key "$(ANVIL_PRIVATE_KEY)" 2>&1)" || { printf '%s\n' "$$deploy_output"; exit 1; }; \
	echo "$$deploy_output"; \
	orderbook_address="$$(printf '%s\n' "$$deploy_output" | awk '/Deployed to:|Contract Address:/ {print $$NF}' | tail -n1)"; \
	if [ -z "$$orderbook_address" ]; then \
		orderbook_address="$$(printf '%s\n' "$$deploy_output" | grep -Eo '"deployedTo"[[:space:]]*:[[:space:]]*"0x[0-9a-fA-F]{40}"' | head -n1 | grep -Eo '0x[0-9a-fA-F]{40}' || true)"; \
	fi; \
	[ -n "$$orderbook_address" ] || { echo "Failed to parse deployed OrderBook address"; exit 1; }; \
	upsert_env() { \
		key="$$1"; value="$$2"; file="$(ANVIL_STATE_FILE)"; \
		if grep -q "^$${key}=" "$$file"; then \
			awk -v k="$$key" -v v="$$value" 'index($$0, k "=")==1 {print k "=" v; next} {print}' "$$file" > "$$file.tmp"; \
			mv "$$file.tmp" "$$file"; \
		else \
			printf '%s=%s\n' "$$key" "$$value" >> "$$file"; \
		fi; \
	}; \
	upsert_env ORDERBOOK_ADDRESS "$$orderbook_address"; \
	echo "ORDERBOOK_ADDRESS=$$orderbook_address"

deploy-and-configure-cli: build-cli deploy
	@set -eu; \
	. ./$(ANVIL_STATE_FILE); \
	[ -n "$${ORDERBOOK_ADDRESS:-}" ] || { echo "ORDERBOOK_ADDRESS missing in $(ANVIL_STATE_FILE). Run: make deploy"; exit 1; }; \
	$(CLI_BIN) config rpc "$(ANVIL_RPC_URL)"; \
	$(CLI_BIN) config chain-id "$(ANVIL_CHAIN_ID)"; \
	$(CLI_BIN) config contract "$$ORDERBOOK_ADDRESS"; \
	$(CLI_BIN) config show

import-deployer-wallet: build-cli init-state
	@set -eu; \
	if $(CLI_BIN) wallet list | grep -qi "$(ANVIL_DEPLOYER_ADDRESS)"; then \
		echo "Deployer wallet already imported: $(ANVIL_DEPLOYER_ADDRESS)"; \
	else \
		$(CLI_BIN) wallet import --private-key "$(ANVIL_PRIVATE_KEY)" --password "$(DEX_WALLET_PASSWORD)"; \
	fi

create-contract: create-market

create-market: build-cli ensure-anvil init-state import-deployer-wallet
	@set -eu; \
	. ./$(ANVIL_STATE_FILE); \
	[ -n "$${ORDERBOOK_ADDRESS:-}" ] || { echo "ORDERBOOK_ADDRESS missing in $(ANVIL_STATE_FILE). Run: make deploy-and-configure-cli"; exit 1; }; \
	base="$${BASE:-$${TOKEN_A_ADDRESS:-}}"; \
	quote="$${QUOTE:-$${TOKEN_B_ADDRESS:-}}"; \
	wallet="$${WALLET:-$(MARKET_WALLET)}"; \
	[ -n "$$base" ] || { echo "Missing base token address. Set BASE=0x... or run make create-tokens first."; exit 1; }; \
	[ -n "$$quote" ] || { echo "Missing quote token address. Set QUOTE=0x... or run make create-tokens first."; exit 1; }; \
	DEX_WALLET_PASSWORD="$(DEX_WALLET_PASSWORD)" $(CLI_BIN) trade market create --yes --contract "$$ORDERBOOK_ADDRESS" --wallet "$$wallet" --base "$$base" --quote "$$quote"

create-tokens: check-tools ensure-anvil build-cli build-forge init-state
	@set -eu; \
	deploy_a="$$(forge create src/MockERC20.sol:MockERC20 $(FORGE_CREATE_FLAGS) --rpc-url "$(ANVIL_RPC_URL)" --private-key "$(ANVIL_PRIVATE_KEY)" --constructor-args "$(TOKEN_A_NAME)" "$(TOKEN_A_SYMBOL)" "$(TOKEN_DECIMALS)" 2>&1)" || { printf '%s\n' "$$deploy_a"; exit 1; }; \
	echo "$$deploy_a"; \
	token_a="$$(printf '%s\n' "$$deploy_a" | awk '/Deployed to:|Contract Address:/ {print $$NF}' | tail -n1)"; \
	if [ -z "$$token_a" ]; then \
		token_a="$$(printf '%s\n' "$$deploy_a" | grep -Eo '"deployedTo"[[:space:]]*:[[:space:]]*"0x[0-9a-fA-F]{40}"' | head -n1 | grep -Eo '0x[0-9a-fA-F]{40}' || true)"; \
	fi; \
	[ -n "$$token_a" ] || { echo "Failed to parse TOKEN_A address"; exit 1; }; \
	deploy_b="$$(forge create src/MockERC20.sol:MockERC20 $(FORGE_CREATE_FLAGS) --rpc-url "$(ANVIL_RPC_URL)" --private-key "$(ANVIL_PRIVATE_KEY)" --constructor-args "$(TOKEN_B_NAME)" "$(TOKEN_B_SYMBOL)" "$(TOKEN_DECIMALS)" 2>&1)" || { printf '%s\n' "$$deploy_b"; exit 1; }; \
	echo "$$deploy_b"; \
	token_b="$$(printf '%s\n' "$$deploy_b" | awk '/Deployed to:|Contract Address:/ {print $$NF}' | tail -n1)"; \
	if [ -z "$$token_b" ]; then \
		token_b="$$(printf '%s\n' "$$deploy_b" | grep -Eo '"deployedTo"[[:space:]]*:[[:space:]]*"0x[0-9a-fA-F]{40}"' | head -n1 | grep -Eo '0x[0-9a-fA-F]{40}' || true)"; \
	fi; \
	[ -n "$$token_b" ] || { echo "Failed to parse TOKEN_B address"; exit 1; }; \
	upsert_env() { \
		key="$$1"; value="$$2"; file="$(ANVIL_STATE_FILE)"; \
		if grep -q "^$${key}=" "$$file"; then \
			awk -v k="$$key" -v v="$$value" 'index($$0, k "=")==1 {print k "=" v; next} {print}' "$$file" > "$$file.tmp"; \
			mv "$$file.tmp" "$$file"; \
		else \
			printf '%s=%s\n' "$$key" "$$value" >> "$$file"; \
		fi; \
	}; \
	upsert_env TOKEN_A_ADDRESS "$$token_a"; \
	upsert_env TOKEN_B_ADDRESS "$$token_b"; \
	$(CLI_BIN) config token add "$$token_a" "$(TOKEN_A_SYMBOL)" --decimals "$(TOKEN_DECIMALS)"; \
	$(CLI_BIN) config token add "$$token_b" "$(TOKEN_B_SYMBOL)" --decimals "$(TOKEN_DECIMALS)"; \
	echo "TOKEN_A_ADDRESS=$$token_a"; \
	echo "TOKEN_B_ADDRESS=$$token_b"; \
	echo "Registered tracked tokens in CLI config."

mint: check-tools ensure-anvil init-state
	@set -eu; \
	. ./$(ANVIL_STATE_FILE); \
	to="$${TO:-}"; \
	[ -n "$$to" ] || { echo "Missing recipient. Usage: make mint TO=0x..."; exit 1; }; \
	[ -n "$${TOKEN_A_ADDRESS:-}" ] || { echo "TOKEN_A_ADDRESS missing in $(ANVIL_STATE_FILE)"; exit 1; }; \
	[ -n "$${TOKEN_B_ADDRESS:-}" ] || { echo "TOKEN_B_ADDRESS missing in $(ANVIL_STATE_FILE)"; exit 1; }; \
	cast send "$$TOKEN_A_ADDRESS" "mint(address,uint256)" "$$to" "$(MINT_AMOUNT)" --rpc-url "$(ANVIL_RPC_URL)" --private-key "$(ANVIL_PRIVATE_KEY)"; \
	cast send "$$TOKEN_B_ADDRESS" "mint(address,uint256)" "$$to" "$(MINT_AMOUNT)" --rpc-url "$(ANVIL_RPC_URL)" --private-key "$(ANVIL_PRIVATE_KEY)"; \
	echo "Minted $(MINT_AMOUNT) of both tokens to $$to"

add-account: ensure-anvil build-cli init-state
	@set -eu; \
	. ./$(ANVIL_STATE_FILE); \
	[ -n "$${TOKEN_A_ADDRESS:-}" ] || { echo "TOKEN_A_ADDRESS missing in $(ANVIL_STATE_FILE). Run: make create-tokens"; exit 1; }; \
	[ -n "$${TOKEN_B_ADDRESS:-}" ] || { echo "TOKEN_B_ADDRESS missing in $(ANVIL_STATE_FILE). Run: make create-tokens"; exit 1; }; \
	new_wallet_out="$$($(CLI_BIN) wallet create --yes --no-mnemonic --password "$(NEW_WALLET_PASSWORD)")"; \
	echo "$$new_wallet_out"; \
	new_wallet="$$(printf '%s\n' "$$new_wallet_out" | awk '/Created wallet:/ {print $$3}' | tail -n1)"; \
	[ -n "$$new_wallet" ] || { echo "Failed to parse newly created wallet address"; exit 1; }; \
	cast send "$$new_wallet" --value "$(ETH_FUND_AMOUNT)" --rpc-url "$(ANVIL_RPC_URL)" --private-key "$(ANVIL_PRIVATE_KEY)"; \
	cast send "$$TOKEN_A_ADDRESS" "mint(address,uint256)" "$$new_wallet" "$(MINT_AMOUNT)" --rpc-url "$(ANVIL_RPC_URL)" --private-key "$(ANVIL_PRIVATE_KEY)"; \
	cast send "$$TOKEN_B_ADDRESS" "mint(address,uint256)" "$$new_wallet" "$(MINT_AMOUNT)" --rpc-url "$(ANVIL_RPC_URL)" --private-key "$(ANVIL_PRIVATE_KEY)"; \
	upsert_env() { \
		key="$$1"; value="$$2"; file="$(ANVIL_STATE_FILE)"; \
		if grep -q "^$${key}=" "$$file"; then \
			awk -v k="$$key" -v v="$$value" 'index($$0, k "=")==1 {print k "=" v; next} {print}' "$$file" > "$$file.tmp"; \
			mv "$$file.tmp" "$$file"; \
		else \
			printf '%s=%s\n' "$$key" "$$value" >> "$$file"; \
		fi; \
	}; \
	upsert_env ADDED_WALLET_ADDRESS "$$new_wallet"; \
	upsert_env CLI_NEW_WALLET_ADDRESS "$$new_wallet"; \
	echo "CLI_NEW_WALLET_ADDRESS=$$new_wallet"; \
	echo "Minted $(MINT_AMOUNT) of both tokens to $$new_wallet"

testnet: ensure-anvil deploy-and-configure-cli create-tokens add-account
	@set -eu; \
	. ./$(ANVIL_STATE_FILE); \
	[ -n "$${ORDERBOOK_ADDRESS:-}" ] || { echo "ORDERBOOK_ADDRESS missing in $(ANVIL_STATE_FILE)"; exit 1; }; \
	wallet="$${CLI_NEW_WALLET_ADDRESS:-$${ADDED_WALLET_ADDRESS:-}}"; \
	[ -n "$$wallet" ] || { echo "No new wallet found in $(ANVIL_STATE_FILE). Run: make add-account"; exit 1; }; \
	[ -n "$${TOKEN_A_ADDRESS:-}" ] || { echo "TOKEN_A_ADDRESS missing in $(ANVIL_STATE_FILE)"; exit 1; }; \
	[ -n "$${TOKEN_B_ADDRESS:-}" ] || { echo "TOKEN_B_ADDRESS missing in $(ANVIL_STATE_FILE)"; exit 1; }; \
	DEX_WALLET_PASSWORD="$(NEW_WALLET_PASSWORD)" $(CLI_BIN) trade market create --yes --contract "$$ORDERBOOK_ADDRESS" --wallet "$$wallet" --base "$$TOKEN_A_ADDRESS" --quote "$$TOKEN_B_ADDRESS"; \
	echo ""; \
	echo "Quick summary:"; \
	echo "- RPC: $(ANVIL_RPC_URL)"; \
	echo "- OrderBook: $$ORDERBOOK_ADDRESS"; \
	echo "- Wallet: $$wallet"; \
	echo "- CLI binary: $(CLI_BIN)"; \
	echo "- Example: DEX_WALLET_PASSWORD=\"$(NEW_WALLET_PASSWORD)\" $(CLI_BIN) trade order count --contract $$ORDERBOOK_ADDRESS"

show-state: init-state
	@cat $(ANVIL_STATE_FILE)

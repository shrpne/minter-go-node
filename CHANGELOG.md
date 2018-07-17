# Changelog

## 0.0.6
*Jule 16th, 2018*

BREAKING CHANGES

- [core] Change commissions
- [testnet] New testnet id
- [core] Fix transaction decoding issue
- [core] Remove transaction ConvertCoin, add SellCoin and BuyCoin. For details see the docs.
- [core] Coin name is now limited to max 64 bytes
- [api] Update estimate exchange endpoint

IMPROVEMENT

- [api] Update transaction api
- [api] Add transaction result to block api
- [mempool] Mempool cache is disabled
- [tendermint] Updated to v0.22.4
- [versioning] Adapt Semantic Versioning https://semver.org/
- [client] Add --disable-api flag to client

## 0.0.5
*Jule 4rd, 2018*

BREAKING CHANGES

- [core] Remove Reserve Coin from coin object. All coins should be reserved with base coin
- [core] Limit tx payload and service data to 128 bytes
- [core] Fix critical issue with instant convert of 2 custom coins 
- [testnet] New testnet chain id (minter-test-network-9)
- [tendermint] Switched to v0.22.0

IMPROVEMENT

- [api] Fix issue with not found coins

BUG FIXES

- [api] Fix transaction endpoint

## 0.0.4

*June 24th, 2018*

BREAKING CHANGES

- [validators] Reward now is payed each 12 blocks
- [validators] Change total "validators' power" to 100 mln
- [tendermint] Switched to v0.21.0
- [testnet] New testnet chain id
- [api] Changed */api/block* response format

IMPROVEMENT

- [docs] Updated docs

BUG FIXES

- [validators] Fixed issue with incorrect pubkey length
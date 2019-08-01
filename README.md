[![Build Status](https://travis-ci.org/tokend/stellar-deposit-svc.svg?branch=master)](https://travis-ci.org/tokend/stellar-deposit-svc)

# Stellar deposit integration module
Stellar deposit service is a bridge between TokenD and Stellar, which allows 
to deposit tokens into TokenD from Stellar. It listens for token transfers
specific address. Memo generated on TokenD side must be present in payment in order
for payment to be considered valid.

## Usage

Enviromental variable `KV_VIPER_FILE` must be set and contain path to desired config file.

```bash
stellar-deposit-svc run deposit
```

## Watchlist

In order for service to start watching for specific asset in stellar network, asset details in TokenD must have entry of the following form: 
```json5
{
//...
"stellar": {
   "deposit": true,
   "asset_code": "USD", // Omit for asset type "native"
   "asset_type": "AlphaNum4",
   },
//...
}
```

## Config

```yaml
stellar:
  is_testnet: true # set true for stellar testnet

payment:
  target_address: "G_STELLAR_DEPOSIT_ADDRESS" # address to payments to

deposit:
  asset_owner: "G_SOME_VALID_ADDRESS"
  asset_issuer: "S_SOME_VALID_SECRET_KEY" # Issuer of assets

horizon:
  endpoint:
  signer: "S_SOME_VALID_SECRET_KEY" # Issuer of assets

log:
  level: info
  disable_sentry: true
```

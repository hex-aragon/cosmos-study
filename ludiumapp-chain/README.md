# Ludium App Chain

Ludium App Chain is a application chain based on Cosmos-SDK & CometBFT(Tendermint). This chain will be used for only education purpose.

<!-- 기존 구성과 많이 다르게함. 이해하기 편하게 하기 위함 -->
<!-- https://github.com/cosmos/gaia/blob/v7.0.2/cmd/gaiad/cmd/root.go -->

### Setup

```bash
# install chain
make install

# init chain
./scripts/init.sh

# start chain
./scripts/start.sh
```

### Example

1. buy-name transaction

```bash
# send tx
ludiumappd tx nameservice buy-name foo 20stake --from alice -y --output json | jq .

# check tx
ludiumappd q tx <tx-hash>
```

2. query whois messages by index

```bash
# query
ludiumappd q nameservice show-whois foo -o json | jq

# expected result
# {
#   "whois": {
#     "index": "foo",
#     "name": "foo",
#     "value": "",
#     "price": "20stake",
#     "owner": "cosmos1d2zgeskrvrvxjdsledwt5e0r26pyz5hgyhu63s"
#   }
# }
```

3. query all whois messages

```bash
# query
ludiumappd q nameservice list-whois -o json | jq .

# expected result
# {
#   "whois": [
#     {
#       "index": "foo",
#       "name": "foo",
#       "value": "",
#       "price": "20stake",
#       "owner": "cosmos1d2zgeskrvrvxjdsledwt5e0r26pyz5hgyhu63s"
#     }
#   ],
#   "pagination": {
#     "next_key": null,
#     "total": "0"
#   }
# }
```

4. more examples: https://docs.ignite.com/v0.25/guide/nameservice/play

### References

1. https://github.com/cosmos/cosmos-sdk/tree/v0.45.4/simapp
2. https://github.com/cosmos/gaia/tree/v7.0.2
3. https://github.com/Jeongseup/jeongseupchain
4. https://github.com/cosmosregistry/chain-minimal
5. https://gitlab.onechain.game/cosmos/cosmos-sdk/-/blob/5d32ed615210d9f88914dc78b842a9c107cc2ae7/scripts/protocgen.sh

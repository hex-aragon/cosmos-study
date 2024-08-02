#!/bin/bash

# check bond denom for preparing alice validatror
grep bond_denom ./private/.simapp/config/genesis.json

# "bond_denom": "stake"

# add initial tokens into genesis
./build/simd add-genesis-account alice 100000000stake \
    --home ./private/.simapp \
    --keyring-backend test 
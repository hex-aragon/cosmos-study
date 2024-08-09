#!/bin/bash

simd genesis add-genesis-account alice 5000000000stake --keyring-backend test && simd genesis add-genesis-account bob 5000000000stake --keyring-backend test && simd genesis gentx alice 1000000stake --chain-id gov-demo && simd genesis collect-gentxs
#!/bin/sh

set -exuf -o pipefail

while true; do

  make install
  ssd init local --chain-id statechain

  echo "password" | sscli keys add jack
  echo "password" | sscli keys add alice

  ssd add-genesis-account $(sscli keys show jack -a) 1000thor
  ssd add-genesis-account $(sscli keys show alice -a) 1000thor

  sscli config chain-id statechain
  sscli config output json
  sscli config indent true
  sscli config trust-node true

  if [ -z "${NET:-}" ]; then
    echo "ProdNetwork"
  fi
  if [ -z "${POOL_ADDRESS:-}" ]; then
    echo "empty pool address"
    POOL_ADDRESS=bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6
  fi
  # add jack as a trusted account
  {
    jq --arg VERSION "$(sscli query swapservice version | jq -r .version)" --arg POOL_ADDRESS "$POOL_ADDRESS" --arg VALIDATOR "$(ssd tendermint show-validator)" --arg NODE_ADDRESS "$(sscli keys show jack -a)" --arg OBSERVER_ADDRESS "$(sscli keys show jack -a)" '.app_state.swapservice.node_accounts[0] = {"node_address": $NODE_ADDRESS, "version": $VERSION, "status":"active","bond_address":$POOL_ADDRESS,"accounts":{"bnb_signer_acc": $POOL_ADDRESS, "bepv_validator_acc": $VALIDATOR, "bep_observer_acc": $OBSERVER_ADDRESS}} | .app_state.swapservice.pool_addresses.rotate_at="28800" | .app_state.swapservice.pool_addresses.rotate_window_open_at="27800" | .app_state.swapservice.pool_addresses.current = $POOL_ADDRESS'
  } <~/.ssd/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json ~/.ssd/config/genesis.json
  cat ~/.ssd/config/genesis.json
  ssd validate-genesis
  break

done

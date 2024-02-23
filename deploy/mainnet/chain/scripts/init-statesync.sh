#!/bin/bash
set -eu

echo "Setting statesync config"

# Onomy home dir
ONOMY_HOME=$HOME/.onomy
# Config directories for onomy node
ONOMY_HOME_CONFIG="$ONOMY_HOME/config"
# Config file for onomy node
ONOMY_NODE_CONFIG="$ONOMY_HOME_CONFIG/config.toml"
# App config file for onomy node
# Statysync servers default IPs
ONOMY_STATESYNC_SERVERS_DEFAULT_IPS="rpc-mainnet.onomy.io:443,35.224.118.71:26657"

statesync_nodes=
blk_height=
blk_hash=


read -r -p "Enter IPs of statesync nodes (at least 2) [$ONOMY_STATESYNC_SERVERS_DEFAULT_IPS]:" statesync_ips
statesync_ips=${statesync_ips:-$ONOMY_STATESYNC_SERVERS_DEFAULT_IPS}
for statesync_ip in ${statesync_ips//,/ } ; do
  latest_height=$(curl -s http://$statesync_ip/block | jq -r .result.block.header.height);
  trusted_height=$((latest_height - 2000));

  blk_details=$(curl -s http://$statesync_ip/block?height=$trusted_height | jq -r '.result.block.header.height + "\n" + .result.block_id.hash')
  blk_height=$(echo $blk_details | cut -d$' ' -f1)

  blk_hash=$(echo $blk_details | cut -d$' ' -f2)
  statesync_nodes="$statesync_nodes$statesync_ip,"
done

echo "Setting up trusted block number $blk_height and hash $blk_hash"

# Change statesync settings
crudini --set $ONOMY_NODE_CONFIG statesync enable true
crudini --set $ONOMY_NODE_CONFIG statesync rpc_servers "\"$statesync_nodes\""
crudini --set $ONOMY_NODE_CONFIG statesync trust_height $blk_height
crudini --set $ONOMY_NODE_CONFIG statesync trust_hash "\"$blk_hash\""

echo "Setup for statesync is complete"

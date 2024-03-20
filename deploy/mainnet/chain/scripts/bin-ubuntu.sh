#Use Ubuntu Latest

#Setting up constants
ONOMY_HOME=$HOME/.onomy
ONOMY_SRC=$ONOMY_HOME/src/onomy
COSMOVISOR_SRC=$ONOMY_HOME/src/cosmovisor

ONOMY_VERSION="v1.1.4"
NODE_EXPORTER_VERSION="0.18.1"
COSMOVISOR_VERSION="cosmovisor-v1.0.1"

mkdir -p $ONOMY_HOME
mkdir -p $ONOMY_HOME/bin
mkdir -p $ONOMY_HOME/contracts
mkdir -p $ONOMY_HOME/logs
mkdir -p $ONOMY_HOME/cosmovisor/genesis/bin
mkdir -p $ONOMY_HOME/cosmovisor/upgrades/

echo "----------------------installing onomy---------------"
curl -LO https://github.com/onomyprotocol/onomy/releases/download/$ONOMY_VERSION/onomyd
mv onomyd $ONOMY_HOME/cosmovisor/genesis/bin/onomyd

echo "----------------------installing cosmovisor---------------"
curl -LO https://github.com/onomyprotocol/onomy-sdk/releases/download/$COSMOVISOR_VERSION/cosmovisor
mv cosmovisor $ONOMY_HOME/bin/cosmovisor

# echo "----------------installing eth bridge gbt-------------"
# curl -LO https://github.com/onomyprotocol/arc/releases/download/$ETH_BRIDGE_VERSION/gbt
# mv gbt $ONOMY_HOME/bin/gbt

echo "-------------------installing node_exporter-----------------------"
curl -LO "https://github.com/prometheus/node_exporter/releases/download/v$NODE_EXPORTER_VERSION/node_exporter-$NODE_EXPORTER_VERSION.linux-amd64.tar.gz"
tar -xvf "node_exporter-$NODE_EXPORTER_VERSION.linux-amd64.tar.gz"
mv "node_exporter-$NODE_EXPORTER_VERSION.linux-amd64/node_exporter" $ONOMY_HOME/bin/node_exporter
rm -r "node_exporter-$NODE_EXPORTER_VERSION.linux-amd64"
rm "node_exporter-$NODE_EXPORTER_VERSION.linux-amd64.tar.gz"

echo "-------------------adding binaries to path-----------------------"
chmod +x $ONOMY_HOME/bin/*
export PATH=$PATH:$ONOMY_HOME/bin
chmod +x $ONOMY_HOME/cosmovisor/genesis/bin/*
export PATH=$PATH:$ONOMY_HOME/cosmovisor/genesis/bin

# set the cosmovisor environments
echo "export DAEMON_HOME=$ONOMY_HOME/" >> ~/.profile
echo "export PATH=\$PATH:\$DAEMON_HOME/cosmovisor/genesis/bin:\$DAEMON_HOME/bin" >> ~/.profile
echo "export DAEMON_NAME=onomyd" >> ~/.profile
echo "export DAEMON_RESTART_AFTER_UPGRADE=true" >> ~/.profile

source $HOME/.profile
ulimit -S -n 65536

echo "Onomy binaries are installed successfully."

---
title: Cosmos Hub 3 Upgrade
order: 1
---

<!-- markdown-link-check-disable -->

The following document describes the necessary steps involved that validators and full node operators
must take in order to upgrade from `cosmoshub-3` to `cosmoshub-4`. The Cosmos teams
will post an official `cosmoshub-4` genesis file, but it is recommended that validators
execute the following instructions in order to verify the resulting genesis file.

There is a strong social consensus around proposal `Cosmos Hub 4 Upgrade Proposal`
on `cosmoshub-3`. Following proposals #[27](https://www.mintscan.io/cosmos/proposals/27), #[35](https://www.mintscan.io/cosmos/proposals/35) and #[36](https://www.mintscan.io/cosmos/proposals/36).
This indicates that the upgrade procedure should be performed on `February 18, 2021 at 06:00 UTC`.

- [Summary](#summary)
- [Migrations](#migrations)
- [Preliminary](#preliminary)
- [Major Updates](#major-updates)
- [Risks](#risks)
- [Recovery](#recovery)
- [Upgrade Procedure](#upgrade-procedure)
- [Guidance for Full Node Operators](#guidance-for-full-node-operators)
- [Notes for Service Providers](#notes-for-service-providers)

# Summary

The Cosmoshub-3 will undergo a scheduled upgrade to Cosmoshub-4 on Feb 18, 2021 at 6 UTC.

The following is a short summary of the upgrade steps:
    1. Stopping the running Onomy v2.0.x instance
    1. Backing up configs, data, and keys used for running Cosmoshub-3
    1. Resetting state to clear the local Cosmoshub-3 state
    1. Copying the cosmoshub-4 genesis file to the Onomy config folder (either after migrating an existing cosmoshub-3 genesis export, or downloading the cosmoshub-4 genesis from the mainnet github)
    1. Installing the Onomy v4.0.x release
    1. Starting the Onomy v4.0.x instance to resume the Cosmos hub chain at a height of <cosmoshub3 height> + 1.

Specific instructions for validators are available in [Upgrade Procedure](#upgrade-procedure),
and specific instructions for full node operators are available in [Guidance for Full Node Operators](#guidance-for-full-node-operators).

Upgrade coordination and support for validators will be available on the #validators-verified channel of the [Cosmos Discord](https://discord.gg/cosmosnetwork).

The network upgrade can take the following potential pathways:

1. Happy path: Validator successfully migrates the cosmoshub-3 genesis file to a cosmoshub-4 genesis file, and the validator can successfully start Onomy v4 with the cosmoshub-4 genesis within 1-2 hours of the scheduled upgrade.
1. Not-so-happy path: Validators have trouble migrating the cosmoshub-3 genesis to a cosmoshub-4 genesis, but can obtain the genesis file from the Cosmos mainnet github repo and can successfully start Onomy v4 within 1-2 hours of the scheduled upgrade.  
1. Abort path: In the rare event that the team becomes aware of critical issues, which result in an unsuccessful migration within a few hours, the upgrade will be announced as aborted
   on the #validators-verified channel of [Discord](https://discord.gg/cosmosnetwork), and validators will need to resume running cosmoshub-3 network without any updates or changes.
   A new governance proposal for the upgrade will need to be issued and voted on by the community.

# Migrations

These chapters contain all the migration guides to update your app and modules to Cosmos v0.40 Stargate.

If you’re running a block explorer, wallet, exchange, validator, or any other service (eg. custody provider) that depends upon the Cosmos Hub or Cosmos ecosystem, you’ll want to pay attention, because this upgrade will involve substantial changes.

1. [App and Modules Migration](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/app_and_modules.md)
2. [REST Endpoints Migration](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/rest.md)
3. [Inter-Blockchain Communication (IBC)– cross-chain transactions](https://figment.io/resources/cosmos-stargate-upgrade-overview/#ibc)
4. [Protobuf Migration – blockchain performance & dev acceleration](https://figment.io/resources/cosmos-stargate-upgrade-overview/#proto)
5. [State Sync – minutes to sync new nodes](https://figment.io/resources/cosmos-stargate-upgrade-overview/#sync)
6. [Full-Featured Light Clients](https://figment.io/resources/cosmos-stargate-upgrade-overview/#light)
7. [Chain Upgrade Module – upgrade automation](https://figment.io/resources/cosmos-stargate-upgrade-overview/#upgrade)

If you want to test the procedure before the update happens on 18th of February, please see this post accordingly:

<https://github.com/onomyprotocol/onomy-rebuild/issues/569#issuecomment-767910963>

## Preliminary

Many changes have occurred to the Cosmos SDK and the Onomy application since the latest
major upgrade (`cosmoshub-3`). These changes notably consist of many new features,
protocol changes, and application structural changes that favor developer ergonomics
and application development.

First and foremost, [IBC](https://docs.cosmos.network/main/ibc/overview.html) following
the [Interchain Standads](https://github.com/cosmos/ics#ibc-quick-references) will be enabled.
This upgrade comes with several improvements in efficiency, node synchronization and following blockchain upgrades.
More details on the [Stargate Website](https://stargate.cosmos.network/).

__[Onomy](https://github.com/onomyprotocol/onomy-rebuild) application v4.0.2 is
what full node operators will upgrade to and run in this next major upgrade__.
Following Cosmos SDK version v0.41.2 and Tendermint v0.34.7.

Validators should expect that at least 16GB of RAM needs to be provisioned to process the first new block on cosmoshub-4.

## Major Updates

There are many notable features and changes in the upcoming release of the SDK. Many of these
are discussed at a high level
[here](https://github.com/cosmos/stargate).

Some of the biggest changes to take note on when upgrading as a developer or client are the following:

- __Protocol Buffers__: Initially the Cosmos SDK used Amino codecs for nearly all encoding and decoding.
In this version a major upgrade to Protocol Buffers have been integrated. It is expected that with Protocol Buffers
applications gain in speed, readability, convenience and interoperability with many programming languages.
[Read more](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/app_and_modules.md#protocol-buffers)
- __CLI__: The CLI and the daemon for a blockchain were separated in previous versions of the Cosmos SDK. This
led to a `onomyd` and `onomycli` binary which were separated and could be used for different interactions with the
blockchain. Both of these have been merged into one `onomyd` which now supports the commands the `onomycli` previously
supported.
- __Node Configuration__: Previously blockchain data and node configuration was stored in `~/.onomy/`, these will
now reside in `~/.onomy/`, if you use scripts that make use of the configuration or blockchain data, make sure to update the path.

## Risks

As a validator performing the upgrade procedure on your consensus nodes carries a heightened risk of
double-signing and being slashed. The most important piece of this procedure is verifying your
software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade
procedure again during the network startup. If you discover a mistake in the process, the best thing
to do is wait for the network to start before correcting it. If the network is halted and you have
started with a different genesis file than the expected one, seek advice from a Tendermint developer
before resetting your validator.

## Recovery

Prior to exporting `cosmoshub-3` state, validators are encouraged to take a full data snapshot at the
export height before proceeding. Snapshotting depends heavily on infrastructure, but generally this
can be done by backing up the `.onomy` directory.

It is critically important to back-up the `.onomy/data/priv_validator_state.json` file after stopping your onomyd process. This file is updated every block as your validator participates in a consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to
onomy v2.0.15 with v0.37.15 of the _Cosmos SDK_ and restore to their latest snapshot before restarting their nodes.

## Upgrade Procedure

__Note__: It is assumed you are currently operating a full-node running onomy v2.0.15 with v0.37.15 of the _Cosmos SDK_.

The version/commit hash of Onomy v2.0.15: `89cf7e6fc166eaabf47ad2755c443d455feda02e`

1. Verify you are currently running the correct version (v2.0.15) of _onomyd_:

   ```bash
    $ onomyd version --long
    name: onomy
    server_name: onomyd
    client_name: onomycli
    version: 2.0.15
    commit: 89cf7e6fc166eaabf47ad2755c443d455feda02e
    build_tags: netgo,ledger
    go: go version go1.15 darwin/amd64
   ```

1. Make sure your chain halts at the right time and date:
    February 18, 2021 at 06:00 UTC is in UNIX seconds: `1613628000`

    ```bash
    perl -i -pe 's/^halt-time =.*/halt-time = 1613628000/' ~/.onomy/config/app.toml
    ```

1. After the chain has halted, make a backup of your `.onomy` directory

    ```bash
    mv ~/.onomy ./onomyd_backup
    ```

    __NOTE__: It is recommended for validators and operators to take a full data snapshot at the export
   height before proceeding in case the upgrade does not go as planned or if not enough voting power
   comes online in a sufficient and agreed upon amount of time. In such a case, the chain will fallback
   to continue operating `cosmoshub-3`. See [Recovery](#recovery) for details on how to proceed.

1. Export existing state from `cosmoshub-3`:

   Before exporting state via the following command, the `onomyd` binary must be stopped!
   As a validator, you can see the last block height created in the
   `~/.onomy/data/priv_validator_state.json` - or now residing in `onomyd_backup` when you made
    a backup as in the last step - and obtain it with

   ```bash
   cat ~/.onomy/data/priv_validator_state.json | jq '.height'
   ```

   ```bash
   onomyd export --height=<height> > cosmoshub_3_genesis_export.json
   ```

   _this might take a while, you can expect an hour for this step_

1. Verify the SHA256 of the (sorted) exported genesis file:

    Compare this value with other validators / full node operators of the network.
    Going forward it will be important that all parties can create the same genesis file export.

   ```bash
   $ jq -S -c -M '' cosmoshub_3_genesis_export.json | shasum -a 256
   [SHA256_VALUE]  cosmoshub_3_genesis_export.json
   ```

1. At this point you now have a valid exported genesis state! All further steps now require
v4.0.2 of [Onomy](https://github.com/onomyprotocol/onomy-rebuild).
Cross check your genesis hash with other peers (other validators) in the chat rooms.

   __NOTE__: Go [1.15+](https://golang.org/dl/) is required!

   ```bash
   git clone https://github.com/onomyprotocol/onomy-rebuild.git && cd onomy && git checkout v4.0.2; make install
   ```

1. Verify you are currently running the correct version (v4.0.2) of the _Onomy_:

   ```bash
    name: onomy
    server_name: onomyd
    version: 4.0.2
    commit: 6d46572f3273423ad9562cf249a86ecc8206e207
    build_tags: netgo,ledger
    ...
   ```

    The version/commit hash of Onomy v4.0.2: `6d46572f3273423ad9562cf249a86ecc8206e207`

1. Migrate exported state from the current v2.0.15 version to the new v4.0.2 version:

   ```bash
   onomyd migrate cosmoshub_3_genesis_export.json --chain-id=cosmoshub-4 --initial-height [last_cosmoshub-3_block+1] > genesis.json
   ```

   This will migrate our exported state into the required `genesis.json` file to start the cosmoshub-4.

1. Verify the SHA256 of the final genesis JSON:

   ```bash
   $ jq -S -c -M '' genesis.json | shasum -a 256
   [SHA256_VALUE]  genesis.json
   ```

    Compare this value with other validators / full node operators of the network.
    It is important that each party can reproduce the same genesis.json file from the steps accordingly.

1. Reset state:

   __NOTE__: Be sure you have a complete backed up state of your node before proceeding with this step.
   See [Recovery](#recovery) for details on how to proceed.

   ```bash
   onomyd unsafe-reset-all
   ```

1. Move the new `genesis.json` to your `.onomy/config/` directory

    ```bash
    cp genesis.json ~/.onomy/config/
    ```

1. Start your blockchain

    ```bash
    onomyd start
    ```

    Automated audits of the genesis state can take 30-120 min using the crisis module. This can be disabled by
    `onomyd start --x-crisis-skip-assert-invariants`.

# Guidance for Full Node Operators

1. Verify you are currently running the correct version (v2.0.15) of _onomyd_:

   ```bash
    $ onomyd version --long
    name: onomy
    server_name: onomyd
    client_name: onomycli
    version: 2.0.15
    commit: 89cf7e6fc166eaabf47ad2755c443d455feda02e
    build_tags: netgo,ledger
    go: go version go1.15 darwin/amd64
   ```

1. Stop your Onomy v2.0.15 instance.

1. After the chain has halted, make a backup of your `.onomy` directory

   ```bash
   mv ~/.onomy ./onomyd_backup
   ```

   __NOTE__: It is recommended for validators and operators to take a full data snapshot at the export
   height before proceeding in case the upgrade does not go as planned or if not enough voting power
   comes online in a sufficient and agreed upon amount of time. That means the backup of `.onomy` should
   only take place once the chain has halted at UNIX time `1613628000`.
   In such a case, the chain will fallback
   to continue operating `cosmoshub-3`. See [Recovery](#recovery) for details on how to proceed.

1. Download the cosmoshub-4 genesis file from the [Cosmos Mainnet Github](https://github.com/cosmos/mainnet).
   This file will be generated by a validator that is migrating from cosmoshub-3 to cosmoshub-4.
   The cosmoshub-4 genesis file will be validated by community participants, and
   the hash of the file will be shared on the #validators-verified channel of the [Cosmos Discord](https://discord.gg/cosmosnetwork).

1. Install v4.0.2 of [Onomy](https://github.com/onomyprotocol/onomy-rebuild).

   __NOTE__: Go [1.15+](https://golang.org/dl/) is required!

   ```bash
   git clone https://github.com/onomyprotocol/onomy-rebuild.git && cd onomy && git checkout v4.0.2; make install
   ```

1. Verify you are currently running the correct version (v4.0.2) of the _Onomy_:

   ```bash
    name: onomy
    server_name: onomyd
    version: 4.0.2
    commit: 6d46572f3273423ad9562cf249a86ecc8206e207
    build_tags: netgo,ledger
    ...
   ```

   The version/commit hash of Onomy v4.0.2: `6d46572f3273423ad9562cf249a86ecc8206e207`

1. Reset state:

   __NOTE__: Be sure you have a complete backed up state of your node before proceeding with this step.
   See [Recovery](#recovery) for details on how to proceed.

   ```bash
   onomyd unsafe-reset-all
   ```

1. Move the new `genesis.json` to your `.onomy/config/` directory

    ```bash
    cp genesis.json ~/.onomy/config/
    ```

1. Start your blockchain

    ```bash
    onomyd start
    ```

   Automated audits of the genesis state can take 30-120 min using the crisis module. This can be disabled by
   `onomyd start --x-crisis-skip-assert-invariants`.

## Notes for Service Providers

# REST server

In case you have been running REST server with the command `onomycli rest-server` previously, running this command will not be necessary anymore.
API server is now in-process with daemon and can be enabled/disabled by API configuration in your `.onomy/config/app.toml`:

```
[api]
# Enable defines if the API server should be enabled.
enable = false
# Swagger defines if swagger documentation should automatically be registered.
swagger = false
```

`swagger` setting refers to enabling/disabling swagger docs API, i.e, /swagger/ API endpoint.

# gRPC Configuration

gRPC configuration in your `.onomy/config/app.toml`

```yaml
[grpc]
# Enable defines if the gRPC server should be enabled.
enable = true
# Address defines the gRPC server address to bind to.
address = "0.0.0.0:9090"
```

# State Sync

State Sync Configuration in your `.onomy/config/app.toml`

```yaml
# State sync snapshots allow other nodes to rapidly join the network without replaying historical
# blocks, instead downloading and applying a snapshot of the application state at a given height.
[state-sync]
# snapshot-interval specifies the block interval at which local state sync snapshots are
# taken (0 to disable). Must be a multiple of pruning-keep-every.
snapshot-interval = 0
# snapshot-keep-recent specifies the number of recent snapshots to keep and serve (0 to keep all).
snapshot-keep-recent = 2
```

<!-- markdown-link-check-enable -->
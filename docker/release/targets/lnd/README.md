# Running a dockerized BTCD/LND Simnet

This guide will show you how to setup a btcd with lnd simnet environment.

The environment consists of four nodes, one btcd node, and three lnd nodes (miner, alice and bob).

- The btcd instance will mine coins to the 'miner' lnd node.
- The 'alice' and 'bob' lnd nodes will be bootstrapped with 100 sBTC each, as part of the setup.

## Requirements

The following services are required in order to run this environment:

| Library        | Download                                        |
|----------------|-------------------------------------------------|
| docker         | https://www.docker.com/products/docker-desktop/ |
| docker-compose | https://docs.docker.com/compose/install/        |
| jq             | https://stedolan.github.io/jq/download/         |

Just to note:

- We will also assume that you've added your current user to the docker group. You can find a guide on how to do this at https://docs.docker.com/engine/install/linux-postinstall/.
- This will ensure that you can run docker under your current user.

## Setting up the environment

Bootstrapping is pretty straightforward, assuming that you have all of the requirements above installed. It can be done the following way:

### Running Setup

Navigate to your indra project root directory, and run the following: 

- We will assume in the following example the directory is contained in `/opt/indra-labs/indra`.
- BEWARE: This script must be run from the project root directory!

``` 
    docker/release/targets/lnd/scripts/setup.sh
```

When complete, it will produce an environment configuration file, located at `docker/release/targets/lnd/.env`. It will be in the following format:

```
    MINER_PUBKEY=<lightning_public_key>
    MINER_ADDRESS=<bitcoin_address>
    ALICE_PUBKEY=<lightning_public_key>
    ALICE_ADDRESS=<bitcoin_address>
    BOB_PUBKEY=<lightning_public_key>
    BOB_ADDRESS=<bitcoin_address>
```

### Using the environment config

The config file has two functions:
- The MINER_ADDRESS is passed to the docker-compose.yml file, on start. This will ensure that any block rewards will be send to the 'miner' node.   
- The rest of the environment variables can be used by the user for constructing transactions. See the use-cases below.

```
    export $(grep -vE "^(#.*|\s*)$" /opt/indra-labs/indra/docker/release/targets/lnd/.env)
```

This will take around 30 seconds to complete. Once complete, we can move on to using the simnet.

### Adding the bin folder to your $PATH (recommended)

For example: (this will not persist, and assumes your indra project root as at /opt/indra-labs/indra):

```
    export PATH=$PATH:/opt/indra-labs/indra/docker/release/targets/lnd/bin
```

#### Adding $PATH persistence

If you would like to persist your path, check out this tutorial: https://linuxhint.com/add-path-permanently-linux/

## Starting / Stopping the network

The following section will show you how to start and stop the network.

### Starting
To start the environment, run the following from your indra project root directory:

```
    docker-compose --file=docker/release/targets/lnd/docker-compose.yml up
```

#### Running in the background

To start the environment *as a background process*, run the following (from your indra project root directory):

```
    docker-compose --file=docker/release/targets/lnd/docker-compose.yml up --detach
```

### Stopping

To stop the environment, run the following (from your indra project root directory):

```
    docker-compose --file=docker/release/targets/lnd/docker-compose.yml down
```

## Running commands on the network

Assuming we've already started the environment, we can run commands for all nodes in the simnet.

### Generating blocks

To generate a new block, run the following:

```
    lnsim-btcctl generate 1
```

If you would like to generate many blocks, you can use the second argument. Here's an example to generate 100 blocks:

```
    lnsim-btcctl generate 100
```

#### A small rule of thumb

- In order for a lightning channel to be opened, there is a requirement that 6 blocks must be generated to confirm the open.

### Interacting with the LND nodes

The lncli interface for any of the nodes can be accessed with the following commands

#### Getting the wallet balance:

To check the balance of the 'miner' node
```
    lnsim-lncli-miner walletbalance
```
To check alice's balance:
```
    lnsim-lncli-alice walletbalance
```
To check bob's balance:
```
    lnsim-lncli-bob walletbalance
```

## Some useful use-cases

### Sending alice and bob 1 sBTC from the miner

``` 
    lnsim-lncli-miner sendcoins --addr $LNSIM_ALICE_ADDRESS --amt 100000000
    lnsim-lncli-miner sendcoins --addr $LNSIM_BOB_ADDRESS --amt 100000000
    
    lnsim-btcctl generate 1
```

# fin; happy simulating!

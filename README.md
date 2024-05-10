# cosmos-node-finder
A cli tool to find cosmos chains public rpc nodes and live peers.

## Description

cosmos-node-finder is a cli tool to find [cosmos chains](https://cosmos.network/) public rpc nodes and live peers.
You can get detailed info of these public rpc nodes(latest height,if it is an archive node, validator...)

Thanks to strangelove's https://github.com/strangelove-ventures/tmp2p, now the peer found will be validated.

## Warning and Disclaimer
cosmos-node-finder is intended for educational and informational purposes only. Users must not use this tool for illegal activities, such as network attacks or unauthorized data access. Compliance with all applicable laws in your region is required. The developer is not responsible for any misuse or consequences arising from the use of this tool. Use responsibly.

## Prerequisites

Install go 1.20 or higher

## Build

```bash
go mod tidy
go build 
```

## Usage
### Find public rpcs by one rpc
```bash
./cosmos-node-finder rpcfinder <one public rpc> <chain-id>
```
for example
```bash
./cosmos-node-finder rpcfinder https://rpc.evmos.silentvalidator.com:443 evmos_9001-2
```
### Find live peers by one rpc 
```bash
./cosmos-node-finder peerfinder <one public rpc> <chain-id>
```
for example
```bash
./cosmos-node-finder peerfinder https://rpc.evmos.silentvalidator.com:443 evmos_9001-2
```
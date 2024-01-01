/*
Copyright © 2024 silent silentvalidator@gmail.com
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"sync"
)

// peerfinderCmd represents the peerfinder command
var peerfinderCmd = &cobra.Command{
	Use:   "peerfinder <rpc> <chain-id>",
	Short: "Find cosmos chain live peers by one rpc and chain id",
	Long: `Find cosmos chain live peers by one rpc and chain id.
This command may take up to 1 minute to complete.`,
	Example: "cosmos-node-finder peerfinder https://rpc.evmos.silentvalidator.com:443 evmos_9001-2",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		timeout, err := cmd.Flags().GetFloat64("timeout")
		if err != nil {
			fmt.Println("Invalid timeout value")
			return
		}
		rpc := args[0]
		chainid := args[1]
		peers, err := GetLivePeersInOneString(rpc, chainid, timeout, 3)
		if err != nil {
			fmt.Println(err)
		}
		if peers == "" {
			fmt.Println("No live peers found, pls check your rpc and chain-id")
			return
		}
		fmt.Println(peers)

	},
}

func init() {
	rootCmd.AddCommand(peerfinderCmd)
	peerfinderCmd.Flags().Float64P("timeout", "t", 10.0, "Timeout in seconds. Set less timeout if you want to get result faster.(You may get less peers)")
}

// GetLivePeers get live peers from []PeerInfo
func GetLivePeers(peersinfo []PeerInfo) []string {
	var LivePeers []string
	for _, peer := range peersinfo {
		LivePeers = append(LivePeers, peer.Id+"@"+peer.ListenAddr)
	}
	return LivePeers
}

// GetLivePeersByOneRPC filter all peers by query all reachable rpc
func GetLivePeersByOneRPC(rpcaddr string, chainid string, timeout float64, maxRetries int) ([]string, error) {
	netinfo, err := GetNetInfo(rpcaddr, timeout)
	if err != nil {
		return []string{}, fmt.Errorf("Error to connect the given rpc: %w", err)
	}
	//rpcs := GetAllOpenRPCPeers(netinfo, chainid, timeout, make(map[string]bool), 0)
	rpcs, err := GetRpcsInfoByOneRpc(rpcaddr, chainid, timeout, maxRetries)
	if err != nil {
		return []string{}, fmt.Errorf("Error to connect the given rpc: %w", err)
	}
	var LivePeers []string
	//save current rpc's peers
	LivePeers = GetLivePeers(GetPeersInfo(netinfo, chainid))
	var wg sync.WaitGroup
	var mutex sync.Mutex
	for _, rpc := range rpcs {
		wg.Add(1)
		go func(rpcaddr string) {
			defer wg.Done()
			retries := 0
			for retries <= maxRetries {
				rpcnetinfo, err := GetNetInfo(rpcaddr, timeout)
				if err != nil {
					retries++
					continue
				}
				peers := GetLivePeers(GetPeersInfo(rpcnetinfo, chainid))
				mutex.Lock()
				LivePeers = append(LivePeers, peers...)
				mutex.Unlock()
				break
			}
		}(rpc.RpcAddress)
	}
	wg.Wait()
	// filter out duplicate peers
	uniquePeers := make(map[string]bool)
	var uniquePeersList []string

	for _, peer := range LivePeers {
		if !uniquePeers[peer] {
			uniquePeers[peer] = true
			uniquePeersList = append(uniquePeersList, peer)
		}
	}
	return uniquePeersList, nil

}

// GetLivePeersInOneString get live peers in one string
func GetLivePeersInOneString(rpcaddr string, chainid string, timeout float64, maxRetries int) (string, error) {
	peers, err := GetLivePeersByOneRPC(rpcaddr, chainid, timeout, maxRetries)
	if err != nil {
		return "", err
	}
	var peersString string
	var peersInOneString string
	if len(peers) == 0 {
		return "", nil
	}
	for _, peer := range peers {
		peersString = peersString + peer + ","

	}
	peersInOneString = peersString[:len(peersString)-1]
	return peersInOneString, nil
}

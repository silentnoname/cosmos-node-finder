/*
Copyright © 2024 silent silentvalidator@gmail.com
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// rpcfinderCmd represents the rpcfinder command
var rpcfinderCmd = &cobra.Command{
	Use:     "rpcfinder",
	Short:   "Find cosmos chain public rpcs by one rpc and chain id",
	Long:    `Find cosmos chain public rpcs by one rpc and chain id. You can get detail info of these public rpc nodes(latest height,if it is an archive node, ...)`,
	Example: "cosmos-node-finder rpcfinder https://rpc.evmos.silentvalidator.com:443 evmos_9001-2",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		timeout, err := cmd.Flags().GetFloat64("timeout")
		if err != nil {
			fmt.Println("Invalid timeout value")
			return
		}
		rpc := args[0]
		chainid := args[1]
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			rpcInfos, err := GetRpcsInfoByOneRpc(rpc, chainid, timeout, 3)
			if err != nil {
				fmt.Println(err)
				return
			}
			jsonBytes, err := json.Marshal(rpcInfos)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(jsonBytes))
			return
		}
		err = GetRpcsInfoTableByOneRpc(rpc, chainid, timeout, 3)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(rpcfinderCmd)
	rpcfinderCmd.Flags().Float64P("timeout", "t", 10.0, "Timeout in seconds. Set less timeout if you want to get result faster.(You may get less peers)")
	rpcfinderCmd.Flags().Bool("json", false, "Output in JSON format. Use --json to enable it.")
}

type NetInfo struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Listening bool     `json:"listening"`
		Listeners []string `json:"listeners"`
		NPeers    string   `json:"n_peers"`
		Peers     []struct {
			NodeInfo struct {
				ProtocolVersion struct {
					P2P   string `json:"p2p"`
					Block string `json:"block"`
					App   string `json:"app"`
				} `json:"protocol_version"`
				ID         string `json:"id"`
				ListenAddr string `json:"listen_addr"`
				Network    string `json:"network"`
				Version    string `json:"version"`
				Channels   string `json:"channels"`
				Moniker    string `json:"moniker"`
				Other      struct {
					TxIndex    string `json:"tx_index"`
					RPCAddress string `json:"rpc_address"`
				} `json:"other"`
			} `json:"node_info"`
			IsOutbound       bool `json:"is_outbound"`
			ConnectionStatus struct {
				Duration    string `json:"Duration"`
				SendMonitor struct {
					Start    time.Time `json:"Start"`
					Bytes    string    `json:"Bytes"`
					Samples  string    `json:"Samples"`
					InstRate string    `json:"InstRate"`
					CurRate  string    `json:"CurRate"`
					AvgRate  string    `json:"AvgRate"`
					PeakRate string    `json:"PeakRate"`
					BytesRem string    `json:"BytesRem"`
					Duration string    `json:"Duration"`
					Idle     string    `json:"Idle"`
					TimeRem  string    `json:"TimeRem"`
					Progress int       `json:"Progress"`
					Active   bool      `json:"Active"`
				} `json:"SendMonitor"`
				RecvMonitor struct {
					Start    time.Time `json:"Start"`
					Bytes    string    `json:"Bytes"`
					Samples  string    `json:"Samples"`
					InstRate string    `json:"InstRate"`
					CurRate  string    `json:"CurRate"`
					AvgRate  string    `json:"AvgRate"`
					PeakRate string    `json:"PeakRate"`
					BytesRem string    `json:"BytesRem"`
					Duration string    `json:"Duration"`
					Idle     string    `json:"Idle"`
					TimeRem  string    `json:"TimeRem"`
					Progress int       `json:"Progress"`
					Active   bool      `json:"Active"`
				} `json:"RecvMonitor"`
				Channels []struct {
					ID                int    `json:"ID"`
					SendQueueCapacity string `json:"SendQueueCapacity"`
					SendQueueSize     string `json:"SendQueueSize"`
					Priority          string `json:"Priority"`
					RecentlySent      string `json:"RecentlySent"`
				} `json:"Channels"`
			} `json:"connection_status"`
			RemoteIP string `json:"remote_ip"`
		} `json:"peers"`
	} `json:"result"`
}

type PeerInfo struct {
	Id            string
	ListenAddr    string
	Moniker       string
	RpcIsOpen     bool
	Rpc           string
	RpcReachable  bool
	TxIndexIsOpen bool
}

// GetNetInfo get netinfo from rpc
func GetNetInfo(rpc string, timeout float64) (NetInfo, error) {

	req, err := http.NewRequest("GET", rpc+"/net_info?", nil)
	if err != nil {
		return NetInfo{}, err
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return NetInfo{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return NetInfo{}, err
	}

	var result NetInfo
	decoder := json.NewDecoder(bytes.NewReader(body))
	err = decoder.Decode(&result)
	return result, nil
}

// GetPeersInfo get peers info( []PeerInfo) from netinfo
func GetPeersInfo(Netinfo NetInfo, ChainId string) []PeerInfo {
	var peers []PeerInfo
	for _, peer := range Netinfo.Result.Peers {
		if peer.NodeInfo.Version == "0.0.1" {
			continue
		}
		if peer.NodeInfo.Network != ChainId {
			continue
		}
		//if remote_ip is ipv6, we need to remove it
		if strings.Count(peer.RemoteIP, ":") >= 2 {
			continue
		}
		//listenaddr have 2 situation, one is like "tcp://0.0.0.0:26656", another is like  "65.108.72.253:29656", we need change the first one to latter one
		part := strings.Split(peer.NodeInfo.ListenAddr, ":")
		if part[0] == "tcp" {
			peer.NodeInfo.ListenAddr = peer.RemoteIP + ":" + part[2]
		}
		if part[0] == "http" {
			peer.NodeInfo.ListenAddr = peer.RemoteIP + ":" + part[2]
		}
		if part[0] == "0.0.0.0" {
			peer.NodeInfo.ListenAddr = peer.RemoteIP + ":" + part[1]
		}
		//rpc_address should be "tcp://0.0.0.0:xxxx if rpc is open
		var RpcIsOpen = false
		var RpcPort string
		rpcpart := strings.Split(peer.NodeInfo.Other.RPCAddress, ":")
		if len(rpcpart) == 3 && rpcpart[0] == "tcp" && rpcpart[1] == "//0.0.0.0" {
			RpcIsOpen = true
			RpcPort = rpcpart[2]
		} else if len(rpcpart) == 3 && rpcpart[0] == "tcp" && rpcpart[1] == "//127.0.0.1" {
			RpcIsOpen = false
			RpcPort = rpcpart[2]
		} else {
			continue
		}
		peerinfo := PeerInfo{
			Id:            peer.NodeInfo.ID,
			ListenAddr:    peer.NodeInfo.ListenAddr,
			Moniker:       peer.NodeInfo.Moniker,
			RpcIsOpen:     RpcIsOpen,
			Rpc:           "http://" + peer.RemoteIP + ":" + RpcPort,
			TxIndexIsOpen: peer.NodeInfo.Other.TxIndex == "on",
		}

		peers = append(peers, peerinfo)
	}
	return peers
}

// GetAllOpenRPCPeers filter all reachable public rpc by recursion
func GetAllOpenRPCPeers(netinfo NetInfo, chainid string, timeout float64, visited *sync.Map, currentDepth int) []PeerInfo {
	peersinfo := GetPeersInfo(netinfo, chainid)

	var peers []PeerInfo

	sem := make(chan struct{}, 30) // Number of concurrences
	var wg sync.WaitGroup

	for _, peer := range peersinfo {
		if peer.RpcIsOpen {
			_, loaded := visited.LoadOrStore(peer.Id, peer)
			if !loaded {
				wg.Add(1)
				go func(peer PeerInfo) {
					defer wg.Done()
					sem <- struct{}{}
					defer func() { <-sem }()

					rNetInfo, err := GetNetInfo(peer.Rpc, timeout)
					if err != nil {
						return
					}

					peer.RpcReachable = true
					visited.Store(peer.Id, peer)

					recursionPeers := GetAllOpenRPCPeers(rNetInfo, chainid, timeout, visited, currentDepth+1)
					for _, rpcPeer := range recursionPeers {
						visited.Store(rpcPeer.Id, rpcPeer)
					}
				}(peer)
			}
		}
	}
	wg.Wait()

	visited.Range(func(key, value interface{}) bool {
		peer, ok := value.(PeerInfo)
		if ok && peer.RpcIsOpen && peer.RpcReachable {
			peers = append(peers, peer)
		}
		return true
	})

	return peers
}

type RpcStatus struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		NodeInfo struct {
			ProtocolVersion struct {
				P2P   string `json:"p2p"`
				Block string `json:"block"`
				App   string `json:"app"`
			} `json:"protocol_version"`
			ID         string `json:"id"`
			ListenAddr string `json:"listen_addr"`
			Network    string `json:"network"`
			Version    string `json:"version"`
			Channels   string `json:"channels"`
			Moniker    string `json:"moniker"`
			Other      struct {
				TxIndex    string `json:"tx_index"`
				RPCAddress string `json:"rpc_address"`
			} `json:"other"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHash     string    `json:"latest_block_hash"`
			LatestAppHash       string    `json:"latest_app_hash"`
			LatestBlockHeight   string    `json:"latest_block_height"`
			LatestBlockTime     time.Time `json:"latest_block_time"`
			EarliestBlockHash   string    `json:"earliest_block_hash"`
			EarliestAppHash     string    `json:"earliest_app_hash"`
			EarliestBlockHeight string    `json:"earliest_block_height"`
			EarliestBlockTime   time.Time `json:"earliest_block_time"`
			CatchingUp          bool      `json:"catching_up"`
		} `json:"sync_info"`
		ValidatorInfo struct {
			Address string `json:"address"`
			PubKey  struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"pub_key"`
			VotingPower string `json:"voting_power"`
		} `json:"validator_info"`
	} `json:"result"`
}
type RpcInfo struct {
	RpcAddress          string
	Reachable           bool
	Moniker             string
	ChainId             string
	TxIndex             bool `json:"tx_index"`
	EarliestBlockHeight string
	EarliestBlockTime   time.Time
	LatestBlockHeight   string
	LatestBlockTime     time.Time
	CatchingUp          bool
	IsArchive           bool
	IsValidator         bool
}

func GetRpcStatus(rpc string, timeout float64) (RpcStatus, error) {
	req, err := http.NewRequest("GET", rpc+"/status", nil)
	if err != nil {
		return RpcStatus{}, fmt.Errorf("error connect rpc to get netinfo: %w", err)
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return RpcStatus{}, fmt.Errorf("error connect rpc to get netinfo: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var result RpcStatus
	decoder := json.NewDecoder(bytes.NewReader(body))
	err = decoder.Decode(&result)
	return result, nil
}

func GetRpcsInfoByOneRpc(rpcaddr string, chainid string, timeout float64, maxRetries int) ([]RpcInfo, error) {
	var rpcInfos []RpcInfo
	var wg sync.WaitGroup
	var mutex sync.Mutex
	rpcnetinfo, err := GetNetInfo(rpcaddr, timeout)
	if err != nil {
		return nil, fmt.Errorf("Error to connect the given rpc: %w", err)
	}
	visited := &sync.Map{}
	rpcs := GetAllOpenRPCPeers(rpcnetinfo, chainid, timeout, visited, 0)
	for _, rpc := range rpcs {
		wg.Add(1)
		go func(rpc PeerInfo) {
			defer wg.Done()
			retries := 0
			var rpcInfo RpcInfo

			for retries <= maxRetries {
				rpcStatus, err := GetRpcStatus(rpc.Rpc, timeout)
				if err != nil {
					retries++
					continue
				}
				if rpcStatus.Result.NodeInfo.Network != chainid {
					return
				}
				rpcInfo = RpcInfo{
					RpcAddress:          rpc.Rpc,
					Reachable:           true,
					Moniker:             rpcStatus.Result.NodeInfo.Moniker,
					ChainId:             rpcStatus.Result.NodeInfo.Network,
					TxIndex:             rpcStatus.Result.NodeInfo.Other.TxIndex == "on",
					EarliestBlockHeight: rpcStatus.Result.SyncInfo.EarliestBlockHeight,
					EarliestBlockTime:   rpcStatus.Result.SyncInfo.EarliestBlockTime,
					LatestBlockHeight:   rpcStatus.Result.SyncInfo.LatestBlockHeight,
					LatestBlockTime:     rpcStatus.Result.SyncInfo.LatestBlockTime,
					CatchingUp:          rpcStatus.Result.SyncInfo.CatchingUp,
					IsArchive:           rpcStatus.Result.SyncInfo.EarliestBlockHeight == "1",
					IsValidator:         rpcStatus.Result.ValidatorInfo.VotingPower != "0",
				}
				break
			}
			mutex.Lock()
			rpcInfos = append(rpcInfos, rpcInfo)
			mutex.Unlock()
		}(rpc)
	}

	wg.Wait()
	//rpcinfos ranked by latest block height
	sort.Slice(rpcInfos, func(i, j int) bool {
		return rpcInfos[i].LatestBlockHeight > rpcInfos[j].LatestBlockHeight
	})
	return rpcInfos, nil
}

// GetRpcsInfoTableByOneRpc Here we use tablewriter to print the result
func GetRpcsInfoTableByOneRpc(rpcaddr string, chainid string, timeout float64, maxRetries int) error {
	rpcInfos, err := GetRpcsInfoByOneRpc(rpcaddr, chainid, timeout, maxRetries)
	if err != nil {
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"RpcAddress", "Reachable", "Moniker", "ChainId", "TxIndex", "EarliestBlockHeight", "EarliestBlockTime", "LatestBlockHeight", "LatestBlockTime", "CatchingUp", "IsArchive", "IsValidator"})
	for _, rpcInfo := range rpcInfos {
		table.Append([]string{rpcInfo.RpcAddress, fmt.Sprint(rpcInfo.Reachable), rpcInfo.Moniker, rpcInfo.ChainId, fmt.Sprint(rpcInfo.TxIndex), rpcInfo.EarliestBlockHeight, fmt.Sprint(rpcInfo.EarliestBlockTime), rpcInfo.LatestBlockHeight, fmt.Sprint(rpcInfo.LatestBlockTime), fmt.Sprint(rpcInfo.CatchingUp), fmt.Sprint(rpcInfo.IsArchive), fmt.Sprint(rpcInfo.IsValidator)})
	}
	fmt.Println("Found:", len(rpcInfos), "available", chainid, "rpcs")
	table.Render()
	return nil
}

package cmds

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"runtime"

	integer "github.com/cybriq/proc/pkg/opts/Integer"
	"github.com/cybriq/proc/pkg/opts/config"
	"github.com/cybriq/proc/pkg/opts/duration"
	"github.com/cybriq/proc/pkg/opts/float"
	"github.com/cybriq/proc/pkg/opts/list"
	"github.com/cybriq/proc/pkg/opts/meta"
	"github.com/cybriq/proc/pkg/opts/text"
	"github.com/cybriq/proc/pkg/opts/toggle"
)

const lorem = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim 	veniam, quis 
nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. 
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu 
fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in 
culpa qui officia deserunt mollit anim id est laborum.`

// GetExampleCommands returns available subcommands in hypothetical Parallelcoin
// Pod example for testing (derived from btcd and btcwallet plus
// parallelcoin kopach miner)
func GetExampleCommands() (c *Command) {
	c = &Command{
		Name:          "pod123",
		Description:   "All in one everything for parallelcoin",
		Documentation: lorem,
		Default:       Tags("gui"),
		Configs: config.Opts{
			"AutoPorts": toggle.New(meta.Data{
				Label:         "Automatic Ports",
				Tags:          Tags("node", "wallet"),
				Description:   "RPC and controller ports are randomized, use with controller for automatic peer discovery",
				Documentation: lorem,
				Default:       "false",
			}),
			"AutoListen": toggle.New(meta.Data{
				Aliases:       Tags("AL"),
				Tags:          Tags("node", "wallet"),
				Label:         "Automatic Listeners",
				Description:   "automatically update inbound addresses dynamically according to discovered network interfaces",
				Documentation: lorem,
				Default:       "false",
			}),
			"CAFile": text.New(meta.Data{
				Aliases:       Tags("CA"),
				Tags:          Tags("node", "wallet", "tls"),
				Label:         "Certificate Authority File",
				Description:   "certificate authority file for TLS certificate validation",
				Documentation: lorem,
			}),
			"CPUProfile": text.New(meta.Data{
				Aliases:       Tags("CPR"),
				Tags:          Tags("node", "wallet", "kopach", "worker"),
				Label:         "CPU Profile",
				Description:   "write cpu profile to this file",
				Documentation: lorem,
			}),
			"DisableRPC": toggle.New(meta.Data{
				Aliases:       Tags("NRPC"),
				Tags:          Tags("node", "wallet"),
				Label:         "Disable RPC",
				Description:   "disable rpc servers, as well as kopach controller",
				Documentation: lorem,
				Default:       "false",
			}),
			"Discovery": toggle.New(meta.Data{
				Aliases:       Tags("DI"),
				Tags:          Tags("node"),
				Label:         "Disovery",
				Description:   "enable LAN peer discovery in GUI",
				Documentation: lorem,
				Default:       "false",
			}),
			"Hilite": list.New(meta.Data{
				Aliases: Tags("HL"),
				Tags: Tags(
					"node",
					"wallet",
					"ctl",
					"kopach",
					"worker",
				),
				Label:         "Hilite",
				Description:   "list of packages that will print with attention getters",
				Documentation: lorem,
			}),
			"Locale": text.New(meta.Data{
				Aliases: Tags("LC"),
				Tags: Tags(
					"node",
					"wallet",
					"ctl",
					"kopach",
					"worker",
				),
				Label:         "Language",
				Description:   "user interface language i18 localization",
				Documentation: lorem,
				Options:       Tags("en"),
				Default:       "en",
			}),
			"LimitPass": text.New(meta.Data{
				Aliases:       Tags("LP"),
				Tags:          Tags("node", "wallet"),
				Label:         "Limit Password",
				Description:   "limited user password",
				Documentation: lorem,
				Default:       genPassword(),
			}),
			"LimitUser": text.New(meta.Data{
				Aliases:       Tags("LU"),
				Tags:          Tags("node", "wallet"),
				Label:         "Limit Username",
				Description:   "limited user name",
				Documentation: lorem,
				Default:       "limituser",
			}),
			"LogFilter": list.New(meta.Data{
				Aliases: Tags("LF"),
				Tags: Tags(
					"node",
					"wallet",
					"ctl",
					"kopach",
					"worker",
				),
				Label:         "Log Filter",
				Description:   "list of packages that will not print logs",
				Documentation: lorem,
			}),
			"OneTimeTLSKey": toggle.New(meta.Data{
				Aliases:       Tags("OTK"),
				Tags:          Tags("node", "wallet"),
				Label:         "One Time TLS Key",
				Description:   "generate a new TLS certificate pair at startup, but only write the certificate to disk",
				Documentation: lorem,
				Default:       "false",
			}),
			"Password": text.New(meta.Data{
				Aliases:       Tags("PW"),
				Tags:          Tags("node", "wallet"),
				Label:         "Password",
				Description:   "password for client RPC connections",
				Documentation: lorem,
				Default:       genPassword(),
			}),
			"PipeLog": toggle.New(meta.Data{
				Aliases: Tags("PL"),
				Label:   "Pipe Logger",
				Tags: Tags(
					"node",
					"wallet",
					"ctl",
					"kopach",
					"worker",
				),
				Description:   "enable pipe based logger IPC",
				Documentation: lorem,
				Default:       "false",
			}),
			"Profile": text.New(meta.Data{
				Aliases: Tags("HPR"),
				Tags: Tags("node", "wallet", "ctl", "kopach",
					"worker"),
				Label:       "Profile",
				Description: "http profiling on given port (1024-40000)",
				// Type:        "",
				Documentation: lorem,
			}),
			"RPCCert": text.New(meta.Data{
				Aliases:       Tags("RC"),
				Tags:          Tags("node", "wallet"),
				Label:         "RPC Cert",
				Description:   "location of RPC TLS certificate",
				Documentation: lorem,
				Default:       "~/.pod/rpc.cert",
			}),
			"RPCKey": text.New(meta.Data{
				Aliases:       Tags("RK"),
				Tags:          Tags("node", "wallet"),
				Label:         "RPC Key",
				Description:   "location of rpc TLS key",
				Documentation: lorem,
				Default:       "~/.pod/rpc.key",
			}),
			"RunAsService": toggle.New(meta.Data{
				Aliases:       Tags("RS"),
				Label:         "Run As Service",
				Description:   "shuts down on lock timeout",
				Documentation: lorem,
				Default:       "false",
			}),
			"Save": toggle.New(meta.Data{
				Aliases:       Tags("SV"),
				Label:         "Save Configuration",
				Description:   "save opts given on commandline",
				Documentation: lorem,
				Default:       "false",
			}),
			"ServerTLS": toggle.New(meta.Data{
				Aliases:       Tags("ST"),
				Tags:          Tags("node", "wallet"),
				Label:         "Server TLS",
				Description:   "enable TLS for the wallet connection to node RPC server",
				Documentation: lorem,
				Default:       "true",
			}),
			"ClientTLS": toggle.New(meta.Data{
				Aliases:       Tags("CT"),
				Tags:          Tags("node", "wallet"),
				Label:         "TLS",
				Description:   "enable TLS for RPC client connections",
				Documentation: lorem,
				Default:       "true",
			},
			),
			"TLSSkipVerify": toggle.New(meta.Data{
				Aliases:       Tags("TSV"),
				Tags:          Tags("node", "wallet"),
				Label:         "TLS Skip Verify",
				Description:   "skip TLS certificate verification (ignore CA errors)",
				Documentation: lorem,
				Default:       "false",
			}),
			"Username": text.New(meta.Data{
				Aliases:       Tags("UN"),
				Tags:          Tags("node", "wallet"),
				Label:         "Username",
				Description:   "username for client RPC connections",
				Documentation: lorem,
				Default:       "username",
			}),
			"UseWallet": toggle.New(meta.Data{
				Aliases:       Tags("WC"),
				Tags:          Tags("ctl"),
				Label:         "Connect to Wallet",
				Description:   "set ctl to connect to wallet instead of chain server",
				Documentation: lorem,
				Default:       "false",
			}),
			"WalletOff": toggle.New(meta.Data{
				Aliases:       Tags("WO"),
				Tags:          Tags("wallet"),
				Label:         "Wallet Off",
				Description:   "turn off the wallet backend",
				Documentation: lorem,
				Default:       "false",
			}),
		},
		Commands: Commands{
			{
				Name:          "gui",
				Description:   "ParallelCoin GUI Wallet/Miner/Explorer",
				Documentation: lorem,
				Configs: config.Opts{
					"DarkTheme": toggle.New(meta.Data{
						Aliases:       Tags("DT"),
						Tags:          Tags("gui"),
						Label:         "Dark Theme",
						Description:   "sets dark theme for GUI",
						Documentation: lorem,
						Default:       "false",
					}),
				},
			},
			{
				Name:          "version",
				Description:   "print version and exit",
				Documentation: lorem,
			},
			{
				Name:          "ctl",
				Description:   "command line wallet and chain RPC client",
				Documentation: lorem,
			},
			{
				Name:          "node",
				Description:   "ParallelCoin blockchain node",
				Documentation: lorem,
				Entrypoint: func(c *Command, args []string) error {
					log.I.Ln("running node")
					return nil
				},
				Commands: []*Command{
					{
						Name:          "dropaddrindex",
						Description:   "drop the address database index",
						Documentation: lorem,
					},
					{
						Name:          "droptxindex",
						Description:   "drop the transaction database index",
						Documentation: lorem,
					},
					{
						Name:          "dropcfindex",
						Description:   "drop the cfilter database index",
						Documentation: lorem,
					},
					{
						Name:          "dropindexes",
						Description:   "drop all of the indexes",
						Documentation: lorem,
					},
					{
						Name:          "resetchain",
						Description:   "deletes the current blockchain cache to force redownload",
						Documentation: lorem,
					},
				},
				Configs: config.Opts{
					"AddCheckpoints": list.New(meta.Data{
						Aliases:       Tags("AC"),
						Tags:          Tags("node"),
						Label:         "Add Checkpoints",
						Description:   "add custom checkpoints",
						Documentation: lorem,
					}),
					"AddPeers": list.New(meta.Data{
						Aliases:       Tags("AP"),
						Tags:          Tags("node"),
						Label:         "Add Peers",
						Description:   "manually adds addresses to try to connect to",
						Documentation: lorem,
					}),
					"AddrIndex": toggle.New(meta.Data{
						Aliases:       Tags("AI"),
						Tags:          Tags("node"),
						Label:         "Address Index",
						Description:   "maintain a full address-based transaction index which makes the searchrawtransactions RPC available",
						Documentation: lorem,
						Default:       "false",
					}),
					"BanDuration": duration.New(meta.Data{
						Aliases:       Tags("BD"),
						Tags:          Tags("node", "policy"),
						Label:         "Ban Duration",
						Description:   "how long a ban of a misbehaving peer lasts",
						Documentation: lorem,
						Default:       "24h0m0s",
					}),
					"BanThreshold": integer.New(meta.Data{
						Aliases:       Tags("BT"),
						Tags:          Tags("node", "policy"),
						Label:         "Ban Threshold",
						Description:   "ban score that triggers a ban (default 100)",
						Documentation: lorem,
						Default:       "100",
					}),
					"BlockMaxSize": integer.New(meta.Data{
						Aliases:       Tags("BMXS"),
						Tags:          Tags("node", "policy"),
						Label:         "Block Max Size",
						Description:   "maximum block size in bytes to be used when creating a block",
						Documentation: lorem,
						Default:       "999000",
					}),
					"BlockMaxWeight": integer.New(meta.Data{
						Aliases:       Tags("BMXW"),
						Tags:          Tags("node", "policy"),
						Label:         "Block Max Weight",
						Description:   "maximum block weight to be used when creating a block",
						Documentation: lorem,
						Default:       "3996000",
					}),
					"BlockMinSize": integer.New(meta.Data{
						Aliases:       Tags("BMS"),
						Tags:          Tags("node", "policy"),
						Label:         "Block Min Size",
						Description:   "minimum block size in bytes to be used when creating a block",
						Documentation: lorem,
						Default:       "1000",
					}),
					"BlockMinWeight": integer.New(meta.Data{
						Aliases:       Tags("BMW"),
						Tags:          Tags("node"),
						Label:         "Block Min Weight",
						Description:   "minimum block weight to be used when creating a block",
						Documentation: lorem,
						Default:       "4000",
					}),
					"BlockPrioritySize": integer.New(meta.Data{
						Aliases:       Tags("BPS"),
						Tags:          Tags("node"),
						Label:         "Block Priority Size",
						Description:   "size in bytes for high-priority/low-fee transactions when creating a block",
						Documentation: lorem,
						Default:       "50000",
					}),
					"BlocksOnly": toggle.New(meta.Data{
						Aliases:       Tags("BO"),
						Tags:          Tags("node"),
						Label:         "Blocks Only",
						Description:   "do not accept transactions from remote peers",
						Documentation: lorem,
						Default:       "false",
					}),
					"ConnectPeers": list.New(meta.Data{
						Aliases:       Tags("CPS"),
						Tags:          Tags("node"),
						Label:         "Connect Peers",
						Description:   "connect ONLY to these addresses (disables inbound connections)",
						Documentation: lorem,
					}),
					"Controller": toggle.New(meta.Data{
						Aliases:       Tags("CN"),
						Tags:          Tags("node"),
						Label:         "Enable Controller",
						Description:   "delivers mining jobs over multicast",
						Documentation: lorem,
						Default:       "false",
					}),
					"DbType": text.New(meta.Data{
						Aliases: Tags("DB"),
						Tags:    Tags("node"),
						Label:   "Database Type",
						Description: "type of database storage engine to use for node (" +
							"only one right now, ffldb)",
						Documentation: lorem,
						Options:       Tags("ffldb"),
						Default:       "ffldb",
					}),
					"DisableBanning": toggle.New(meta.Data{
						Aliases:       Tags("NB"),
						Tags:          Tags("node"),
						Label:         "Disable Banning",
						Description:   "disables banning of misbehaving peers",
						Documentation: lorem,
						Default:       "false",
					}),
					"DisableCheckpoints": toggle.New(meta.Data{
						Aliases:       Tags("NCP"),
						Tags:          Tags("node"),
						Label:         "Disable Checkpoints",
						Description:   "disables all checkpoints",
						Documentation: lorem,
						Default:       "false",
					}),
					"DisableDNSSeed": toggle.New(meta.Data{
						Aliases:       Tags("NDS"),
						Tags:          Tags("node"),
						Label:         "Disable DNS Seed",
						Description:   "disable seeding of addresses to peers",
						Documentation: lorem,
						Default:       "false",
					}),
					"DisableListen": toggle.New(meta.Data{
						Aliases:       Tags("NL"),
						Tags:          Tags("node", "wallet"),
						Label:         "Disable Listen",
						Description:   "disables inbound connections for the peer to peer network",
						Documentation: lorem,
						Default:       "false",
					}),
					"ExternalIPs": list.New(meta.Data{
						Aliases:       Tags("EI"),
						Tags:          Tags("node"),
						Label:         "External IP Addresses",
						Description:   "extra addresses to tell peers they can connect to",
						Documentation: lorem,
					}),
					"FreeTxRelayLimit": float.New(meta.Data{
						Aliases:       Tags("LR"),
						Tags:          Tags("node"),
						Label:         "Free Tx Relay Limit",
						Description:   "limit relay of transactions with no transaction fee to the given amount in thousands of bytes per minute",
						Documentation: lorem,
						Default:       "15.0",
					}),
					"LAN": toggle.New(meta.Data{
						Tags:          Tags("node"),
						Label:         "LAN Testnet Mode",
						Description:   "run without any connection to nodes on the internet (does not apply on mainnet)",
						Documentation: lorem,
						Default:       "false",
					}),
					"MaxOrphanTxs": integer.New(meta.Data{
						Aliases:       Tags("MO"),
						Tags:          Tags("node"),
						Label:         "Max Orphan Txs",
						Description:   "max number of orphan transactions to keep in memory",
						Documentation: lorem,
						Default:       "100",
					}),
					"MaxPeers": integer.New(meta.Data{
						Aliases:       Tags("MP"),
						Tags:          Tags("node"),
						Label:         "Max Peers",
						Description:   "maximum number of peers to hold connections with",
						Documentation: lorem,
						Default:       "25",
					}),
					"MinRelayTxFee": float.New(meta.Data{
						Aliases:       Tags("MRTF"),
						Tags:          Tags("node"),
						Label:         "Min Relay Transaction Fee",
						Description:   "the minimum transaction fee in DUO/kB to be considered a non-zero fee",
						Documentation: lorem,
						Default:       "0.00001000",
					}),
					"Network": text.New(meta.Data{
						Aliases:     Tags("NW"),
						Tags:        Tags("node", "wallet"),
						Label:       "Network",
						Description: "connect to this network:",
						Options: []string{
							"mainnet",
							"testnet",
							"regtestnet",
							"simnet",
						},
						Documentation: lorem,
						Default:       "mainnet",
					}),
					"NoCFilters": toggle.New(meta.Data{
						Aliases:       Tags("NCF"),
						Tags:          Tags("node"),
						Label:         "No CFilters",
						Description:   "disable committed filtering (CF) support",
						Documentation: lorem,
						Default:       "false",
					}),
					"NodeOff": toggle.New(meta.Data{
						Aliases:       Tags("NO"),
						Tags:          Tags("node"),
						Label:         "Node Off",
						Description:   "turn off the node backend",
						Documentation: lorem,
						Default:       "false",
					}),
					"NoPeerBloomFilters": toggle.New(meta.Data{
						Aliases:       Tags("NPBF"),
						Tags:          Tags("node"),
						Label:         "No Peer Bloom Filters",
						Description:   "disable bloom filtering support",
						Documentation: lorem,
						Default:       "false",
					}),
					"NoRelayPriority": toggle.New(meta.Data{
						Aliases:       Tags("NRPR"),
						Tags:          Tags("node"),
						Label:         "No Relay Priority",
						Description:   "do not require free or low-fee transactions to have high priority for relaying",
						Documentation: lorem,
						Default:       "false",
					}),
					"OnionEnabled": toggle.New(meta.Data{
						Aliases:       Tags("OE"),
						Tags:          Tags("node"),
						Label:         "Onion Enabled",
						Description:   "enable tor proxy",
						Documentation: lorem,
						Default:       "false",
					}),
					"OnionProxyAddress": text.New(meta.Data{
						Aliases:       Tags("OPA"),
						Tags:          Tags("node"),
						Label:         "Onion Proxy Address",
						Description:   "address of tor proxy you want to connect to",
						Documentation: lorem,
						Default:       "127.0.0.1:1108",
					}),
					"OnionProxyPass": text.New(meta.Data{
						Aliases:       Tags("OPW"),
						Tags:          Tags("node"),
						Label:         "Onion Proxy Password",
						Description:   "password for tor proxy",
						Documentation: lorem,
						Default:       genPassword(),
					}),
					"OnionProxyUser": text.New(meta.Data{
						Aliases:       Tags("OU"),
						Tags:          Tags("node"),
						Label:         "Onion Proxy Username",
						Description:   "tor proxy username",
						Documentation: lorem,
						Default:       "onionproxyuser",
					}),
					"P2PConnect": list.New(meta.Data{
						Aliases:       Tags("P2P"),
						Tags:          Tags("node"),
						Label:         "P2P Connect",
						Description:   "list of addresses reachable from connected networks",
						Documentation: lorem,
						Default:       "127.0.0.1:11048",
					}),
					"P2PListeners": list.New(meta.Data{
						Aliases:       Tags("LA"),
						Tags:          Tags("node"),
						Label:         "P2PListeners",
						Description:   "list of addresses to bind the node listener to",
						Documentation: lorem,
						Default:       "127.0.0.1:11048,127.0.0.11:11048",
					}),
					"ProxyAddress": text.New(meta.Data{
						Aliases:       Tags("PA"),
						Tags:          Tags("node"),
						Label:         "Proxy",
						Description:   "address of proxy to connect to for outbound connections",
						Documentation: lorem,
						Default:       "127.0.0.1:8989",
					}),
					"ProxyPass": text.New(meta.Data{
						Aliases:       Tags("PPW"),
						Tags:          Tags("node"),
						Label:         "Proxy Pass",
						Description:   "proxy password, if required",
						Documentation: lorem,
						Default:       genPassword(),
					}),
					"ProxyUser": text.New(meta.Data{
						Aliases:       Tags("PU"),
						Tags:          Tags("node"),
						Label:         "ProxyUser",
						Description:   "proxy username, if required",
						Documentation: lorem,
						Default:       "proxyuser",
					}),
					"RejectNonStd": toggle.New(meta.Data{
						Aliases:       Tags("REJ"),
						Tags:          Tags("node"),
						Label:         "Reject Non Std",
						Description:   "reject non-standard transactions regardless of the default settings for the active network",
						Documentation: lorem,
						Default:       "false",
					}),
					"RelayNonStd": toggle.New(meta.Data{
						Aliases:       Tags("RNS"),
						Tags:          Tags("node"),
						Label:         "Relay Nonstandard Transactions",
						Description:   "relay non-standard transactions regardless of the default settings for the active network",
						Documentation: lorem,
						Default:       "false",
					}),
					"RPCConnect": text.New(meta.Data{
						Aliases: Tags("RA"),
						Tags:    Tags("node"),
						Label:   "RPC Connect",
						Description: "address of full node RPC for wallet" +
							" to connect to",
						Documentation: lorem,
						Default:       "127.0.0.1:11048",
					}),
					"RPCListeners": list.New(meta.Data{
						Aliases:       Tags("RL"),
						Tags:          Tags("node"),
						Label:         "Node RPC Listeners",
						Description:   "addresses to listen for RPC connections",
						Documentation: lorem,
						Default:       "127.0.0.1:11048",
					}),
					"RPCMaxClients": integer.New(meta.Data{
						Aliases:       Tags("RMXC"),
						Tags:          Tags("node"),
						Label:         "Maximum Node RPC Clients",
						Description:   "maximum number of clients for regular RPC",
						Documentation: lorem,
						Default:       "10",
					}),
					"RPCMaxConcurrentReqs": integer.New(meta.Data{
						Aliases:       Tags("RMCR"),
						Tags:          Tags("node"),
						Label:         "Maximum Node RPC Concurrent Reqs",
						Description:   "maximum number of requests to process concurrently",
						Documentation: lorem,
						Default:       fmt.Sprint(runtime.NumCPU()),
					}),
					"RPCMaxWebsockets": integer.New(meta.Data{
						Aliases:       Tags("RMWS"),
						Tags:          Tags("node"),
						Label:         "Maximum Node RPC Websockets",
						Description:   "maximum number of websocket clients to allow",
						Documentation: lorem,
						Default:       "25",
					}),
					"RPCQuirks": toggle.New(meta.Data{
						Aliases:       Tags("RQ"),
						Tags:          Tags("node"),
						Label:         "Emulate Bitcoin Core RPC Quirks",
						Description:   "enable bugs that replicate bitcoin core RPC's JSON",
						Documentation: lorem,
						Default:       "false",
					}),
					"SigCacheMaxSize": integer.New(meta.Data{
						Aliases:       Tags("SCM"),
						Tags:          Tags("node"),
						Label:         "Signature Cache Max Size",
						Description:   "the maximum number of entries in the signature verification cache",
						Documentation: lorem,
						Default:       "100000",
					}),
					"Solo": toggle.New(meta.Data{
						Label:         "Solo Generate",
						Tags:          Tags("node"),
						Description:   "mine even if not connected to a network",
						Documentation: lorem,
						Default:       "false",
					}),
					"TorIsolation": toggle.New(meta.Data{
						Aliases:       Tags("TI"),
						Tags:          Tags("node"),
						Label:         "Tor Isolation",
						Description:   "makes a separate proxy connection for each connection",
						Documentation: lorem,
						Default:       "false",
					}),
					"TrickleInterval": duration.New(meta.Data{
						Aliases:       Tags("TKI"),
						Tags:          Tags("node"),
						Label:         "Trickle Interval",
						Description:   "minimum time between attempts to send new inventory to a connected peer",
						Documentation: lorem,
						Default:       "1s",
					}),
					"TxIndex": toggle.New(meta.Data{
						Aliases:       Tags("TXI"),
						Tags:          Tags("node"),
						Label:         "Tx Index",
						Description:   "maintain a full hash-based transaction index which makes all transactions available via the getrawtransaction RPC",
						Documentation: lorem,
						Default:       "true",
					}),
					"UPNP": toggle.New(meta.Data{
						Aliases:       Tags("UP"),
						Tags:          Tags("node"),
						Label:         "UPNP",
						Description:   "enable UPNP for NAT traversal",
						Documentation: lorem,
						Default:       "false",
					}),
					"UserAgentComments": list.New(meta.Data{
						Aliases:       Tags("UA"),
						Tags:          Tags("node"),
						Label:         "User Agent Comments",
						Description:   "comment to add to the user agent -- See BIP 14 for more information",
						Documentation: lorem,
					}),
					"UUID": integer.New(meta.Data{
						Label:         "UUID",
						Description:   "instance unique id (32bit random value) (json mangles big 64 bit integers due to float64 numbers)",
						Documentation: lorem,
					}),
					"Whitelists": list.New(meta.Data{
						Aliases:       Tags("WL"),
						Tags:          Tags("node"),
						Label:         "Whitelists",
						Description:   "peers that you don't want to ever ban",
						Documentation: lorem,
					}),
				},
			},
			{
				Name:          "wallet",
				Description:   "run the wallet server (requires a chain node to function)",
				Documentation: lorem,
				Commands: []*Command{
					{
						Name:        "drophistory",
						Description: "reset the wallet transaction history",
						Configs:     config.Opts{},
					},
				},
				Configs: config.Opts{
					"File": text.New(meta.Data{
						Aliases:       Tags("WF"),
						Tags:          Tags("wallet"),
						Label:         "Wallet File",
						Description:   "wallet database file",
						Documentation: lorem,
						Default:       "~/.pod/mainnet/wallet.db",
					}),
					"Pass": text.New(meta.Data{
						Aliases: Tags("WPW"),
						Label:   "Wallet Pass",
						Tags:    Tags("wallet"),
						Description: "password encrypting public data in wallet - only hash is stored" +
							" so give on command line or in environment POD_WALLETPASS",
						Documentation: lorem,
						Default:       genPassword(),
					}),
					"RPCListeners": list.New(meta.Data{
						Aliases:       Tags("WRL"),
						Tags:          Tags("wallet"),
						Label:         "Wallet RPC Listeners",
						Description:   "addresses for wallet RPC server to listen on",
						Documentation: lorem,
					}),
					"RPCMaxClients": integer.New(
						meta.Data{
							Aliases:       Tags("WRMC"),
							Tags:          Tags("wallet"),
							Label:         "Legacy RPC Max Clients",
							Description:   "maximum number of RPC clients allowed for wallet RPC",
							Documentation: lorem,
						}),
					"RPCMaxWebsockets": integer.New(meta.Data{
						Aliases:       Tags("WRMWS"),
						Tags:          Tags("wallet"),
						Label:         "Legacy RPC Max Websockets",
						Description:   "maximum number of websocket clients allowed for wallet RPC",
						Documentation: lorem,
						Default:       "25",
					}),
					"Server": text.New(meta.Data{
						Aliases:       Tags("WS"),
						Tags:          Tags("wallet"),
						Label:         "Wallet Server",
						Description:   "node address to connect wallet server to",
						Documentation: lorem,
					},
					),
				},
			},
			{
				Name:          "kopach",
				Description:   "standalone multicast miner for easy mining farm deployment",
				Documentation: lorem,
				Configs: config.Opts{
					"Generate": toggle.New(meta.Data{
						Aliases:       Tags("GB"),
						Tags:          Tags("node", "kopach"),
						Label:         "Generate Blocks",
						Description:   "turn on Kopach CPU miner",
						Documentation: lorem,
						Default:       "false",
					}),
					"GenThreads": integer.New(meta.Data{
						Aliases:       Tags("GT"),
						Tags:          Tags("kopach"),
						Label:         "Generate Threads",
						Description:   "number of threads to mine with",
						Documentation: lorem,
						Default:       fmt.Sprint(runtime.NumCPU() / 2),
					}),
					"MulticastPass": text.New(meta.Data{
						Aliases:       Tags("PM"),
						Tags:          Tags("node", "kopach"),
						Label:         "Multicast Pass",
						Description:   "password that encrypts the connection to the mining controller",
						Documentation: lorem,
						Default:       genPassword(),
					}),
				},
			},
			{
				Name:        "worker",
				Description: "single thread worker process, normally started by kopach",
			},
		},
	}
	GetConfigBase(c.Configs, c.Name, false)
	return
}

const (
	// RecommendedSeedLen is the recommended length in bytes for a seed to a master node.
	RecommendedSeedLen = 32 // 256 bits
	// HardenedKeyStart is the index at which a hardened key starts. Each extended key has 2^31 normal child keys and
	// 2^31 hardned child keys. Thus the range for normal child keys is [0, 2^31 - 1] and the range for hardened child
	// keys is [2^31, 2^32 - 1].
	HardenedKeyStart = 0x80000000 // 2^31
	// MinSeedBytes is the minimum number of bytes allowed for a seed to a master node.
	MinSeedBytes = 16 // 128 bits
	// MaxSeedBytes is the maximum number of bytes allowed for a seed to a master node.
	MaxSeedBytes = 64 // 512 bits
	// serializedKeyLen is the length of a serialized public or private extended key. It consists of 4 bytes version, 1
	// byte depth, 4 bytes fingerprint, 4 bytes child number, 32 bytes chain code, and 33 bytes public/private key data.
	serializedKeyLen = 4 + 1 + 4 + 4 + 32 + 33 // 78 bytes
	// maxUint8 is the max positive integer which can be serialized in a uint8
	maxUint8 = 1<<8 - 1
)

var (
	ErrInvalidSeedLen = fmt.Errorf(
		"seed length must be between %d and %d "+
			"bits", MinSeedBytes*8, MaxSeedBytes*8,
	)
)

// GenerateSeed returns a cryptographically secure random seed that can be used as the input for the NewMaster function
// to generate a new master node. The length is in bytes and it must be between 16 and 64 (128 to 512 bits). The
// recommended length is 32 (256 bits) as defined by the RecommendedSeedLen constant.
func GenerateSeed(length uint8) ([]byte, error) {
	// Per [BIP32], the seed must be in range [MinSeedBytes, MaxSeedBytes].
	if length < MinSeedBytes || length > MaxSeedBytes {
		return nil, ErrInvalidSeedLen
	}
	buf := make([]byte, length)
	_, e := rand.Read(buf)
	if log.E.Chk(e) {
		return nil, e
	}
	return buf, nil
}

func genPassword() string {
	s, e := GenerateSeed(20)
	if e != nil {
		panic("can't do nothing without entropy! " + e.Error())
	}
	out := make([]byte, 32)
	base32.StdEncoding.Encode(out, s)
	return string(out)
}

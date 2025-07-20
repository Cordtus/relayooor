package config

// ChainConfig represents the configuration for a blockchain
type ChainConfig struct {
	ChainID       string            `json:"chain_id"`
	ChainName     string            `json:"chain_name"`
	AddressPrefix string            `json:"address_prefix"`
	RPCEndpoint   string            `json:"rpc_endpoint,omitempty"`
	RESTEndpoint  string            `json:"rest_endpoint,omitempty"`
	WSEndpoint    string            `json:"ws_endpoint,omitempty"`
	GRPCEndpoint  string            `json:"grpc_endpoint,omitempty"`
	Explorer      string            `json:"explorer"`
	Logo          string            `json:"logo,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// ChannelConfig represents an IBC channel configuration
type ChannelConfig struct {
	SourceChain      string `json:"source_chain"`
	SourceChannel    string `json:"source_channel"`
	DestChain        string `json:"dest_chain"`
	DestChannel      string `json:"dest_channel"`
	SourcePort       string `json:"source_port"`
	DestPort         string `json:"dest_port"`
	Status           string `json:"status"`
}

// ChainRegistry provides chain and channel configurations
type ChainRegistry struct {
	Chains   map[string]ChainConfig  `json:"chains"`
	Channels []ChannelConfig         `json:"channels"`
}

// DefaultChainRegistry returns the default chain configurations
// In production, this should be loaded from a database or config file
func DefaultChainRegistry() *ChainRegistry {
	return &ChainRegistry{
		Chains: map[string]ChainConfig{
			"cosmoshub-4": {
				ChainID:       "cosmoshub-4",
				ChainName:     "Cosmos Hub",
				AddressPrefix: "cosmos",
				RPCEndpoint:   "https://rpc.cosmos.network:443",
				RESTEndpoint:  "https://cosmos-rest.publicnode.com",
				WSEndpoint:    "wss://rpc.cosmos.network/websocket",
				GRPCEndpoint:  "grpc.cosmos.network:443",
				Explorer:      "https://www.mintscan.io/cosmos/txs",
				Logo:          "/images/chains/cosmos.svg",
			},
			"osmosis-1": {
				ChainID:       "osmosis-1",
				ChainName:     "Osmosis",
				AddressPrefix: "osmo",
				RPCEndpoint:   "https://rpc.osmosis.zone:443",
				RESTEndpoint:  "https://osmosis-rest.publicnode.com",
				WSEndpoint:    "wss://rpc.osmosis.zone/websocket",
				GRPCEndpoint:  "grpc.osmosis.zone:9090",
				Explorer:      "https://www.mintscan.io/osmosis/txs",
				Logo:          "/images/chains/osmosis.svg",
			},
			"neutron-1": {
				ChainID:       "neutron-1",
				ChainName:     "Neutron",
				AddressPrefix: "neutron",
				RPCEndpoint:   "https://rpc-kralum.neutron-1.neutron.org:443",
				RESTEndpoint:  "https://neutron-rest.publicnode.com",
				WSEndpoint:    "wss://rpc-kralum.neutron-1.neutron.org/websocket",
				GRPCEndpoint:  "grpc-kralum.neutron-1.neutron.org:80",
				Explorer:      "https://www.mintscan.io/neutron/txs",
				Logo:          "/images/chains/neutron.svg",
			},
			"noble-1": {
				ChainID:       "noble-1",
				ChainName:     "Noble",
				AddressPrefix: "noble",
				RPCEndpoint:   "https://noble-rpc.polkachu.com:443",
				RESTEndpoint:  "https://noble-rest.publicnode.com",
				WSEndpoint:    "wss://noble-rpc.polkachu.com/websocket",
				GRPCEndpoint:  "noble-grpc.polkachu.com:11690",
				Explorer:      "https://www.mintscan.io/noble/txs",
				Logo:          "/images/chains/noble.svg",
			},
			"akashnet-2": {
				ChainID:       "akashnet-2",
				ChainName:     "Akash",
				AddressPrefix: "akash",
				RPCEndpoint:   "https://akash-rpc.polkachu.com:443",
				RESTEndpoint:  "https://akash-rest.publicnode.com",
				WSEndpoint:    "wss://akash-rpc.polkachu.com/websocket",
				GRPCEndpoint:  "akash-grpc.polkachu.com:14490",
				Explorer:      "https://www.mintscan.io/akash/txs",
				Logo:          "/images/chains/akash.svg",
			},
			"stargaze-1": {
				ChainID:       "stargaze-1",
				ChainName:     "Stargaze",
				AddressPrefix: "stars",
				RPCEndpoint:   "https://rpc.stargaze-apis.com:443",
				RESTEndpoint:  "https://stargaze-rest.publicnode.com",
				WSEndpoint:    "wss://rpc.stargaze-apis.com/websocket",
				GRPCEndpoint:  "grpc.stargaze-apis.com:443",
				Explorer:      "https://www.mintscan.io/stargaze/txs",
				Logo:          "/images/chains/stargaze.svg",
			},
			"juno-1": {
				ChainID:       "juno-1",
				ChainName:     "Juno",
				AddressPrefix: "juno",
				RPCEndpoint:   "https://juno-rpc.polkachu.com:443",
				RESTEndpoint:  "https://juno-rest.publicnode.com",
				WSEndpoint:    "wss://juno-rpc.polkachu.com/websocket",
				GRPCEndpoint:  "juno-grpc.polkachu.com:12690",
				Explorer:      "https://www.mintscan.io/juno/txs",
				Logo:          "/images/chains/juno.svg",
			},
			"stride-1": {
				ChainID:       "stride-1",
				ChainName:     "Stride",
				AddressPrefix: "stride",
				RPCEndpoint:   "https://stride-rpc.polkachu.com:443",
				RESTEndpoint:  "https://stride-rest.publicnode.com",
				WSEndpoint:    "wss://stride-rpc.polkachu.com/websocket",
				GRPCEndpoint:  "stride-grpc.polkachu.com:12290",
				Explorer:      "https://www.mintscan.io/stride/txs",
				Logo:          "/images/chains/stride.svg",
			},
			"axelar-dojo-1": {
				ChainID:       "axelar-dojo-1",
				ChainName:     "Axelar",
				AddressPrefix: "axelar",
				RPCEndpoint:   "https://axelar-rpc.quickapi.com:443",
				RESTEndpoint:  "https://axelar-rest.publicnode.com",
				WSEndpoint:    "wss://axelar-rpc.quickapi.com/websocket",
				GRPCEndpoint:  "axelar-grpc.quickapi.com:9090",
				Explorer:      "https://www.mintscan.io/axelar/txs",
				Logo:          "/images/chains/axelar.svg",
			},
			"dydx-mainnet-1": {
				ChainID:       "dydx-mainnet-1",
				ChainName:     "dYdX",
				AddressPrefix: "dydx",
				RPCEndpoint:   "https://dydx-rpc.publicnode.com:443",
				RESTEndpoint:  "https://dydx-rest.publicnode.com",
				WSEndpoint:    "wss://dydx-rpc.publicnode.com/websocket",
				GRPCEndpoint:  "dydx-grpc.publicnode.com:443",
				Explorer:      "https://www.mintscan.io/dydx/txs",
				Logo:          "/images/chains/dydx.svg",
			},
		},
		Channels: []ChannelConfig{
			// Cosmos Hub channels
			{SourceChain: "cosmoshub-4", SourceChannel: "channel-141", DestChain: "osmosis-1", DestChannel: "channel-0", SourcePort: "transfer", DestPort: "transfer", Status: "active"},
			{SourceChain: "cosmoshub-4", SourceChannel: "channel-536", DestChain: "noble-1", DestChannel: "channel-4", SourcePort: "transfer", DestPort: "transfer", Status: "active"},
			{SourceChain: "cosmoshub-4", SourceChannel: "channel-569", DestChain: "neutron-1", DestChannel: "channel-1", SourcePort: "transfer", DestPort: "transfer", Status: "active"},
			// Osmosis channels
			{SourceChain: "osmosis-1", SourceChannel: "channel-0", DestChain: "cosmoshub-4", DestChannel: "channel-141", SourcePort: "transfer", DestPort: "transfer", Status: "active"},
			{SourceChain: "osmosis-1", SourceChannel: "channel-750", DestChain: "noble-1", DestChannel: "channel-1", SourcePort: "transfer", DestPort: "transfer", Status: "active"},
			{SourceChain: "osmosis-1", SourceChannel: "channel-874", DestChain: "neutron-1", DestChannel: "channel-10", SourcePort: "transfer", DestPort: "transfer", Status: "active"},
			// Noble channels
			{SourceChain: "noble-1", SourceChannel: "channel-1", DestChain: "osmosis-1", DestChannel: "channel-750", SourcePort: "transfer", DestPort: "transfer", Status: "active"},
			{SourceChain: "noble-1", SourceChannel: "channel-4", DestChain: "cosmoshub-4", DestChannel: "channel-536", SourcePort: "transfer", DestPort: "transfer", Status: "active"},
			{SourceChain: "noble-1", SourceChannel: "channel-18", DestChain: "neutron-1", DestChannel: "channel-30", SourcePort: "transfer", DestPort: "transfer", Status: "active"},
			// Neutron channels
			{SourceChain: "neutron-1", SourceChannel: "channel-1", DestChain: "cosmoshub-4", DestChannel: "channel-569", SourcePort: "transfer", DestPort: "transfer", Status: "active"},
			{SourceChain: "neutron-1", SourceChannel: "channel-10", DestChain: "osmosis-1", DestChannel: "channel-874", SourcePort: "transfer", DestPort: "transfer", Status: "active"},
			{SourceChain: "neutron-1", SourceChannel: "channel-30", DestChain: "noble-1", DestChannel: "channel-18", SourcePort: "transfer", DestPort: "transfer", Status: "active"},
		},
	}
}

// GetChainByID returns a chain configuration by chain ID
func (r *ChainRegistry) GetChainByID(chainID string) (ChainConfig, bool) {
	chain, exists := r.Chains[chainID]
	return chain, exists
}

// GetChainByPrefix returns a chain configuration by address prefix
func (r *ChainRegistry) GetChainByPrefix(prefix string) (ChainConfig, bool) {
	for _, chain := range r.Chains {
		if chain.AddressPrefix == prefix {
			return chain, true
		}
	}
	return ChainConfig{}, false
}

// GetChannelPairs returns all channel pairs for a given chain
func (r *ChainRegistry) GetChannelPairs(chainID string) []ChannelConfig {
	var channels []ChannelConfig
	for _, ch := range r.Channels {
		if ch.SourceChain == chainID || ch.DestChain == chainID {
			channels = append(channels, ch)
		}
	}
	return channels
}
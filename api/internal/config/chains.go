package config

// ChainConfig represents the configuration for a blockchain
type ChainConfig struct {
	ChainID       string            `json:"chain_id"`
	ChainName     string            `json:"chain_name"`
	AddressPrefix string            `json:"address_prefix"`
	RPCEndpoint   string            `json:"rpc_endpoint,omitempty"`
	APIEndpoint   string            `json:"api_endpoint,omitempty"`
	WSEndpoint    string            `json:"ws_endpoint,omitempty"`
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
				Explorer:      "https://www.mintscan.io/cosmos/txs",
				Logo:          "/images/chains/cosmos.svg",
			},
			"osmosis-1": {
				ChainID:       "osmosis-1",
				ChainName:     "Osmosis",
				AddressPrefix: "osmo",
				Explorer:      "https://www.mintscan.io/osmosis/txs",
				Logo:          "/images/chains/osmosis.svg",
			},
			"neutron-1": {
				ChainID:       "neutron-1",
				ChainName:     "Neutron",
				AddressPrefix: "neutron",
				Explorer:      "https://www.mintscan.io/neutron/txs",
				Logo:          "/images/chains/neutron.svg",
			},
			"noble-1": {
				ChainID:       "noble-1",
				ChainName:     "Noble",
				AddressPrefix: "noble",
				Explorer:      "https://www.mintscan.io/noble/txs",
				Logo:          "/images/chains/noble.svg",
			},
			"akash-1": {
				ChainID:       "akash-1",
				ChainName:     "Akash",
				AddressPrefix: "akash",
				Explorer:      "https://www.mintscan.io/akash/txs",
				Logo:          "/images/chains/akash.svg",
			},
			"stargaze-1": {
				ChainID:       "stargaze-1",
				ChainName:     "Stargaze",
				AddressPrefix: "stars",
				Explorer:      "https://www.mintscan.io/stargaze/txs",
				Logo:          "/images/chains/stargaze.svg",
			},
			"juno-1": {
				ChainID:       "juno-1",
				ChainName:     "Juno",
				AddressPrefix: "juno",
				Explorer:      "https://www.mintscan.io/juno/txs",
				Logo:          "/images/chains/juno.svg",
			},
			"stride-1": {
				ChainID:       "stride-1",
				ChainName:     "Stride",
				AddressPrefix: "stride",
				Explorer:      "https://www.mintscan.io/stride/txs",
				Logo:          "/images/chains/stride.svg",
			},
			"axelar-1": {
				ChainID:       "axelar-1",
				ChainName:     "Axelar",
				AddressPrefix: "axelar",
				Explorer:      "https://www.mintscan.io/axelar/txs",
				Logo:          "/images/chains/axelar.svg",
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
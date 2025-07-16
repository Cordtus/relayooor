package handlers

import (
	"context"
)

// HermesClient interface for interacting with Hermes REST API
type HermesClient interface {
	GetVersion(ctx context.Context) (*VersionResponse, error)
	ClearPackets(ctx context.Context, req *ClearPacketsRequest) (*ClearPacketsResponse, error)
}

type VersionResponse struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

type ClearPacketsRequest struct {
	Chain     string   `json:"chain"`
	Channel   string   `json:"channel"`
	Port      string   `json:"port"`
	Sequences []uint64 `json:"sequences"`
}

type ClearPacketsResponse struct {
	Success  bool     `json:"success"`
	TxHashes []string `json:"tx_hashes"`
	Error    string   `json:"error,omitempty"`
}
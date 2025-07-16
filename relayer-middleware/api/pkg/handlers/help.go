package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HelpHandler provides help and educational content
type HelpHandler struct {
	terms map[string]TermDefinition
}

// TermDefinition represents a technical term explanation
type TermDefinition struct {
	Term       string   `json:"term"`
	Definition string   `json:"definition"`
	Examples   []string `json:"examples,omitempty"`
	Related    []string `json:"related,omitempty"`
}

// NewHelpHandler creates a new help handler
func NewHelpHandler() *HelpHandler {
	return &HelpHandler{
		terms: initializeTermDefinitions(),
	}
}

// GetTermDefinition handles GET /api/v1/help/terms/:term
func (h *HelpHandler) GetTermDefinition(c *gin.Context) {
	term := c.Param("term")
	
	if definition, exists := h.terms[term]; exists {
		c.JSON(http.StatusOK, definition)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": gin.H{
			"code":    "TERM_NOT_FOUND",
			"message": "Term definition not found",
		},
	})
}

// GetAllTerms handles GET /api/v1/help/terms
func (h *HelpHandler) GetAllTerms(c *gin.Context) {
	// Return list of terms without full definitions
	terms := make([]string, 0, len(h.terms))
	for term := range h.terms {
		terms = append(terms, term)
	}

	c.JSON(http.StatusOK, gin.H{
		"terms": terms,
		"count": len(terms),
	})
}

// GetGlossary handles GET /api/v1/help/glossary
func (h *HelpHandler) GetGlossary(c *gin.Context) {
	category := c.Query("category")
	
	glossary := make(map[string][]TermDefinition)
	
	for _, term := range h.terms {
		cat := h.categorize(term.Term)
		if category == "" || cat == category {
			glossary[cat] = append(glossary[cat], term)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"glossary":   glossary,
		"categories": h.getCategories(),
	})
}

// RegisterRoutes registers help handler routes
func (h *HelpHandler) RegisterRoutes(router *gin.RouterGroup) {
	help := router.Group("/help")
	{
		help.GET("/terms", h.GetAllTerms)
		help.GET("/terms/:term", h.GetTermDefinition)
		help.GET("/glossary", h.GetGlossary)
	}
}

func initializeTermDefinitions() map[string]TermDefinition {
	return map[string]TermDefinition{
		"channel": {
			Term:       "channel",
			Definition: "An IBC channel is a connection between two blockchains that allows them to exchange tokens and data",
			Examples:   []string{"channel-0", "channel-141"},
			Related:    []string{"ibc", "relayer", "connection"},
		},
		"sequence": {
			Term:       "sequence",
			Definition: "A unique number identifying each transfer in order. Every packet sent through a channel gets the next sequence number",
			Examples:   []string{"Sequence: 12345", "Sequence: 12346"},
			Related:    []string{"packet", "channel"},
		},
		"packet": {
			Term:       "packet",
			Definition: "A bundle of data being transferred between blockchains. Contains your tokens and transfer information",
			Examples:   []string{"Transfer packet", "Acknowledgment packet"},
			Related:    []string{"sequence", "channel", "timeout"},
		},
		"timeout": {
			Term:       "timeout",
			Definition: "The time limit for a transfer to complete before it expires. Protects your funds if something goes wrong",
			Examples:   []string{"15 minute timeout", "1 hour timeout"},
			Related:    []string{"packet", "stuck"},
		},
		"relayer": {
			Term:       "relayer",
			Definition: "A service that helps move transfers between blockchains by relaying packets back and forth",
			Examples:   []string{"Hermes relayer", "Go relayer"},
			Related:    []string{"channel", "packet"},
		},
		"stuck": {
			Term:       "stuck",
			Definition: "A transfer that hasn't completed within the expected time (usually 15+ minutes). Your funds are safe but need help to complete",
			Examples:   []string{"Stuck for 2 hours", "3 stuck packets"},
			Related:    []string{"timeout", "relayer"},
		},
		"memo": {
			Term:       "memo",
			Definition: "A message attached to your payment that tells our system which transfers to clear. Must be exact!",
			Examples:   []string{"CLR-a1b2c3d4-...", "Payment reference"},
			Related:    []string{"payment"},
		},
		"gas": {
			Term:       "gas",
			Definition: "Network fees paid to process transactions on the blockchain. Like postage for your transfer",
			Examples:   []string{"Gas fee: 0.01 OSMO", "Estimated gas: 200,000"},
			Related:    []string{"fee", "transaction"},
		},
		"ibc": {
			Term:       "ibc",
			Definition: "Inter-Blockchain Communication - The protocol that allows different blockchains to talk to each other and transfer tokens",
			Examples:   []string{"IBC transfer", "IBC channel"},
			Related:    []string{"channel", "packet", "relayer"},
		},
		"denom": {
			Term:       "denom",
			Definition: "The type of token or currency. Each blockchain has its own denominations",
			Examples:   []string{"uosmo (Osmosis)", "uatom (Cosmos Hub)"},
			Related:    []string{"token", "chain"},
		},
		"clearing": {
			Term:       "clearing",
			Definition: "The process of completing stuck transfers by submitting the necessary proofs to the blockchain",
			Examples:   []string{"Clear packets", "Clearing service"},
			Related:    []string{"stuck", "packet", "relayer"},
		},
		"acknowledgment": {
			Term:       "acknowledgment",
			Definition: "A confirmation that your transfer was received on the destination chain",
			Examples:   []string{"Waiting for acknowledgment", "Acknowledgment received"},
			Related:    []string{"packet", "channel"},
		},
		"chain": {
			Term:       "chain",
			Definition: "A blockchain network. Each chain is independent but can connect to others through IBC",
			Examples:   []string{"Osmosis chain", "Cosmos Hub chain"},
			Related:    []string{"ibc", "channel"},
		},
		"wallet": {
			Term:       "wallet",
			Definition: "Your account on the blockchain that holds your tokens. Like a digital bank account",
			Examples:   []string{"Keplr wallet", "cosmos1abc..."},
			Related:    []string{"address", "tokens"},
		},
		"fee": {
			Term:       "fee",
			Definition: "Costs associated with processing transactions. Includes both network fees (gas) and service fees",
			Examples:   []string{"Service fee: 1 OSMO", "Total fee: 1.5 OSMO"},
			Related:    []string{"gas", "payment"},
		},
	}
}

func (h *HelpHandler) categorize(term string) string {
	categories := map[string][]string{
		"IBC Basics": {"ibc", "channel", "packet", "sequence", "acknowledgment"},
		"Problems":   {"stuck", "timeout"},
		"Solutions":  {"clearing", "relayer"},
		"Payments":   {"fee", "gas", "memo", "denom"},
		"General":    {"chain", "wallet"},
	}

	for category, terms := range categories {
		for _, t := range terms {
			if t == term {
				return category
			}
		}
	}

	return "Other"
}

func (h *HelpHandler) getCategories() []string {
	return []string{
		"IBC Basics",
		"Problems",
		"Solutions", 
		"Payments",
		"General",
		"Other",
	}
}
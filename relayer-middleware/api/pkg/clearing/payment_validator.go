package clearing

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"google.golang.org/protobuf/proto"
)

type PaymentValidator struct {
	tolerance         *big.Int // Allow small variance for gas estimation
	serviceAddress    string
}

type Payment struct {
	FromAddress string
	ToAddress   string
	Amount      string
	Denom       string
}

type Transaction struct {
	Hash     string
	Messages []Message
}

type Message struct {
	Type  string
	Value []byte
}

// Custom error types for better handling
type ErrOverpayment struct {
	Required string
	Paid     string
	Denom    string
}

func (e *ErrOverpayment) Error() string {
	return fmt.Sprintf("overpayment detected: required %s %s, paid %s %s",
		e.Required, e.Denom, e.Paid, e.Denom)
}

type ErrUnderpayment struct {
	Required string
	Paid     string
	Denom    string
}

func (e *ErrUnderpayment) Error() string {
	return fmt.Sprintf("insufficient payment: required %s %s, paid %s %s",
		e.Required, e.Denom, e.Paid, e.Denom)
}

var (
	ErrNoPaymentFound = fmt.Errorf("no payment found in transaction")
)

func NewPaymentValidator(serviceAddress string) *PaymentValidator {
	// Allow 1% tolerance for gas estimation variance
	return &PaymentValidator{
		tolerance:      big.NewInt(1), // 1%
		serviceAddress: serviceAddress,
	}
}

func (v *PaymentValidator) ValidatePayment(ctx context.Context, token *ClearingToken, tx *Transaction) error {
	// Extract all payments from transaction (handle multiple)
	payments, err := v.extractAllPayments(tx)
	if err != nil {
		return fmt.Errorf("failed to extract payments: %w", err)
	}

	// Check for multiple payments to same token
	relevantPayments := []Payment{}
	for _, payment := range payments {
		if payment.ToAddress == v.serviceAddress {
			relevantPayments = append(relevantPayments, payment)
		}
	}

	if len(relevantPayments) == 0 {
		return ErrNoPaymentFound
	}

	// Handle multiple payments (aggregate)
	totalPaid := new(big.Int)
	seenDenoms := make(map[string]bool)

	for _, payment := range relevantPayments {
		if payment.Denom != token.AcceptedDenom {
			return fmt.Errorf("invalid payment denomination: expected %s, got %s",
				token.AcceptedDenom, payment.Denom)
		}
		seenDenoms[payment.Denom] = true

		amount := new(big.Int)
		amount.SetString(payment.Amount, 10)
		totalPaid.Add(totalPaid, amount)
	}

	// Validate total amount with tolerance
	required := new(big.Int)
	required.SetString(token.TotalRequired, 10)

	// Calculate tolerance amount
	toleranceAmount := new(big.Int)
	toleranceAmount.Mul(required, v.tolerance)
	toleranceAmount.Div(toleranceAmount, big.NewInt(100))

	// Check if payment is within tolerance
	diff := new(big.Int)
	diff.Sub(totalPaid, required)
	diff.Abs(diff)

	if diff.Cmp(toleranceAmount) > 0 {
		// Check if overpayment (for refund)
		if totalPaid.Cmp(required) > 0 {
			return &ErrOverpayment{
				Required: required.String(),
				Paid:     totalPaid.String(),
				Denom:    token.AcceptedDenom,
			}
		}

		return &ErrUnderpayment{
			Required: required.String(),
			Paid:     totalPaid.String(),
			Denom:    token.AcceptedDenom,
		}
	}

	return nil
}

func (v *PaymentValidator) extractAllPayments(tx *Transaction) ([]Payment, error) {
	payments := []Payment{}

	// Parse transaction to find bank send messages
	for _, msg := range tx.Messages {
		if msg.Type == "/cosmos.bank.v1beta1.MsgSend" {
			// Decode message
			var sendMsg banktypes.MsgSend
			if err := proto.Unmarshal(msg.Value, &sendMsg); err != nil {
				continue
			}

			// Extract each coin as a payment
			for _, coin := range sendMsg.Amount {
				payments = append(payments, Payment{
					FromAddress: sendMsg.FromAddress,
					ToAddress:   sendMsg.ToAddress,
					Amount:      coin.Amount.String(),
					Denom:       coin.Denom,
				})
			}
		}
	}

	return payments, nil
}

func (v *PaymentValidator) extractPaymentAmount(tx *Transaction) (string, string, error) {
	// Parse transaction to find bank send message
	for _, msg := range tx.Messages {
		if msg.Type == "/cosmos.bank.v1beta1.MsgSend" {
			// Decode message
			var sendMsg banktypes.MsgSend
			if err := proto.Unmarshal(msg.Value, &sendMsg); err != nil {
				continue
			}

			// Check if this is our payment
			if sendMsg.ToAddress == v.serviceAddress {
				if len(sendMsg.Amount) != 1 {
					return "", "", fmt.Errorf("payment must contain exactly one coin")
				}

				coin := sendMsg.Amount[0]
				return coin.Amount.String(), coin.Denom, nil
			}
		}
	}

	return "", "", fmt.Errorf("payment not found in transaction")
}

// ValidateMemo checks if the transaction memo matches the expected token
func (v *PaymentValidator) ValidateMemo(tx *Transaction, expectedMemo string) error {
	// TODO: Extract and validate memo from transaction
	// This would require parsing the full transaction structure
	return nil
}

// IsOverpayment checks if an error is an overpayment error
func IsOverpayment(err error) bool {
	_, ok := err.(*ErrOverpayment)
	return ok
}

// IsUnderpayment checks if an error is an underpayment error
func IsUnderpayment(err error) bool {
	_, ok := err.(*ErrUnderpayment)
	return ok
}
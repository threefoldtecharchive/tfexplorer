package phonebook

import (
	"encoding/json"

	schema "github.com/threefoldtech/tfexplorer/schema"
)

type WalletAddress struct {
	Asset   string `bson:"asset" json:"asset"`
	Address string `bson:"address" json:"address"`
}

type User struct {
	ID              schema.ID       `bson:"_id" json:"id"`
	Name            string          `bson:"name" json:"name"`
	Email           string          `bson:"email" json:"email"`
	Pubkey          string          `bson:"pubkey" json:"pubkey"`
	Host            string          `bson:"host" json:"host"`
	Description     string          `bson:"description" json:"description"`
	WalletAddresses []WalletAddress `bson:"wallet_addresses" json:"wallet_addresses"`
	Signature       string          `bson:"-" json:"signature,omitempty"`

	// Trusted Sales channel
	// is a special flag that is only set by TF, if set this user can
	// - sponsors pools
	// - get special discount
	IsTrustedChannel bool `bson:"trusted_sales_channel" json:"trusted_sales_channel"`
}

func NewUser() (User, error) {
	const value = "{}"
	var object User
	if err := json.Unmarshal([]byte(value), &object); err != nil {
		return object, err
	}
	return object, nil
}

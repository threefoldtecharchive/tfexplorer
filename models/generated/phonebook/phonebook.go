package phonebook

import (
	schema "github.com/threefoldtech/tfexplorer/schema"
)

type User struct {
	ID          schema.ID `bson:"_id" json:"id"`
	Name        string    `bson:"name" json:"name"`
	Email       string    `bson:"email" json:"email"`
	Pubkey      string    `bson:"pubkey" json:"pubkey"`
	Host        string    `bson:"host" json:"host"`
	Description string    `bson:"description" json:"description"`

	AutomaticUpgradAgreement bool `bson:"automatic_upgrade_agreement" json:"automatic_upgrade_agreement"`
}

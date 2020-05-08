package threebotconnect

import (
	"context"
	"net/http"

	"crypto/ed25519"

	"github.com/gorilla/mux"
	"github.com/threefoldtech/tfexplorer/pkg/workloads/types"
	"go.mongodb.org/mongo-driver/mongo"
)

// Setup injects and initializes directory package
func Setup(parent *mux.Router, db *mongo.Database) error {
	if err := types.Setup(context.TODO(), db); err != nil {
		return err
	}

	// _, priv, err := box.GenerateKey(crypto_rand.Reader)
	// if err != nil {
	// 	return err
	// }
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}

	var auth = New(priv, "6a27488b.ngrok.io", []byte("IDycw5LUcT+zjdua9QdypQRhfH5XvQkcIuFZaGUf+2s="))
	router := parent.PathPrefix("/auth").Subrouter()

	router.HandleFunc("/login", auth.Login).Methods(http.MethodGet).Name("threebot-login")
	router.HandleFunc("/callback_threebot", auth.Callback).Methods(http.MethodGet).Name("threebot-callback")

	return nil
}

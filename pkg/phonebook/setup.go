package phonebook

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/threefoldtech/tfexplorer/mw"
	"github.com/threefoldtech/tfexplorer/pkg/events"
	phonebook "github.com/threefoldtech/tfexplorer/pkg/phonebook/types"
	"github.com/zaibon/httpsig"
	"go.mongodb.org/mongo-driver/mongo"
)

// Setup injects and initializes directory package
func Setup(parent *mux.Router, db *mongo.Database, e *events.EventProcessingService) error {
	if err := phonebook.Setup(context.TODO(), db); err != nil {
		return err
	}

	userVerifier := httpsig.NewVerifier(mw.NewUserKeyGetter(db))

	var userAPI = UserAPI{
		verifier:               userVerifier,
		eventProcessingService: e,
	}
	users := parent.PathPrefix("/users").Subrouter()

	users.HandleFunc("", mw.AsHandlerFunc(userAPI.create)).Methods(http.MethodPost).Name("user-create")
	users.HandleFunc("", mw.AsHandlerFunc(userAPI.list)).Methods(http.MethodGet).Name(("user-list"))
	users.HandleFunc("/{user_id}", mw.AsHandlerFunc(userAPI.register)).Methods(http.MethodPut).Name("user-register")
	users.HandleFunc("/{user_id}", mw.AsHandlerFunc(userAPI.get)).Methods(http.MethodGet).Name("user-get")
	users.HandleFunc("/{user_id}/validate", mw.AsHandlerFunc(userAPI.validate)).Methods(http.MethodPost).Name("user-validate")

	return nil
}

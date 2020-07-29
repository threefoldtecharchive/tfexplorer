package phonebook

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models"
	"github.com/threefoldtech/tfexplorer/mw"
	"github.com/threefoldtech/tfexplorer/pkg/phonebook/types"
	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/threefoldtech/zos/pkg/crypto"
	"github.com/zaibon/httpsig"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserAPI struct
type UserAPI struct {
	verifier              *httpsig.Verifier
	threebotConnectAPIURL string
}

func (u *UserAPI) isAuthenticated(r *http.Request) bool {
	_, err := u.verifier.Verify(r)
	return err == nil
}

// create user entry point, makes sure name is free for reservation
func (u *UserAPI) create(r *http.Request) (interface{}, mw.Response) {
	var user types.User

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return nil, mw.BadRequest(err)
	}

	// https://github.com/threefoldtech/zos/issues/706
	if err := user.Validate(); err != nil {
		return nil, mw.BadRequest(err)
	}

	// Ensure we do not add user that exist in 3bot connect DB with another public key
	// https://github.com/threefoldtech/home/issues/859
	if err := u.compareUser(user); err != nil {
		if errors.As(err, &errWrongPubKey{}) {
			return nil, mw.Conflict(err)
		}
		return nil, mw.Error(err)
	}

	db := mw.Database(r)
	user, err := types.UserCreate(r.Context(), db, user.Name, user.Email, user.Pubkey)
	if err != nil && errors.Is(err, types.ErrUserExists) {
		return nil, mw.Conflict(err)
	} else if err != nil {
		return nil, mw.Error(err)
	}

	return user, mw.Created()
}

/*
register
As implemented in threebot. It works as a USER update function. To update
any fields, you need to make sure your payload has an extra "sender_signature_hex"
field that is the signature of the payload using the user private key.

This signature is done on a message that is built as defined by the User.Encode() method
*/
func (u *UserAPI) register(r *http.Request) (interface{}, mw.Response) {
	id, err := u.parseID(mux.Vars(r)["user_id"])
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid user id"))
	}

	var payload struct {
		types.User
		Signature string `json:"sender_signature_hex"` // because why not `signature`!
	}

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, mw.BadRequest(err)
	}

	if len(payload.Signature) == 0 {
		return nil, mw.BadRequest(fmt.Errorf("signature is required"))
	}

	signature, err := hex.DecodeString(payload.Signature)
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid signature hex"))
	}
	db := mw.Database(r)

	if err := types.UserUpdate(r.Context(), db, schema.ID(id), signature, payload.User); err != nil {
		if errors.Is(err, types.ErrBadUserUpdate) {
			return nil, mw.BadRequest(err)
		}
		return nil, mw.Error(err)
	}

	return nil, nil
}

func (u *UserAPI) list(r *http.Request) (interface{}, mw.Response) {
	var filter types.UserFilter
	filter = filter.WithName(r.FormValue("name"))
	filter = filter.WithEmail(r.FormValue("email"))

	db := mw.Database(r)

	findOpts := make([]*options.FindOptions, 0, 2)
	pager := models.PageFromRequest(r)
	findOpts = append(findOpts, pager)

	// hide the email of the user for any non authenticated user
	if !u.isAuthenticated(r) {
		findOpts = append(findOpts, options.Find().SetProjection(bson.D{
			{Key: "email", Value: 0},
		}))
	}

	cur, err := filter.Find(r.Context(), db, findOpts...)
	if err != nil {
		return nil, mw.Error(err)
	}

	users := []types.User{}
	if err := cur.All(r.Context(), &users); err != nil {
		return nil, mw.Error(err)
	}

	total, err := filter.Count(r.Context(), db)
	if err != nil {
		return nil, mw.Error(err, http.StatusInternalServerError)
	}

	nrPages := math.Ceil(float64(total) / float64(*pager.Limit))
	pages := fmt.Sprintf("%d", int64(nrPages))

	return users, mw.Ok().WithHeader("Pages", pages)
}

func (u *UserAPI) parseID(id string) (schema.ID, error) {
	v, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "invalid id format")
	}

	return schema.ID(v), nil
}

func (u *UserAPI) get(r *http.Request) (interface{}, mw.Response) {

	userID, err := u.parseID(mux.Vars(r)["user_id"])
	if err != nil {
		return nil, mw.BadRequest(err)
	}
	var filter types.UserFilter
	filter = filter.WithID(userID)

	db := mw.Database(r)
	user, err := filter.Get(r.Context(), db)
	if err != nil {
		return nil, mw.NotFound(err)
	}

	// hide the email of the user for any non authenticated user
	if !u.isAuthenticated(r) {
		user.Email = ""
	}

	return user, nil
}

func (u *UserAPI) validate(r *http.Request) (interface{}, mw.Response) {
	var payload struct {
		Payload   string `json:"payload"`
		Signature string `json:"signature"`
	}

	userID, err := u.parseID(mux.Vars(r)["user_id"])
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, mw.BadRequest(err)
	}

	data, err := hex.DecodeString(payload.Payload)
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "payload must be hex encoded string of original data"))
	}

	signature, err := hex.DecodeString(payload.Signature)
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "signature must be hex encoded string of original data"))
	}

	var filter types.UserFilter
	filter = filter.WithID(userID)

	db := mw.Database(r)
	user, err := filter.Get(r.Context(), db)
	if err != nil {
		return nil, mw.NotFound(err)
	}

	key, err := crypto.KeyFromHex(user.Pubkey)
	if err != nil {
		return nil, mw.Error(err)
	}

	if len(key) != ed25519.PublicKeySize {
		return nil, mw.Error(fmt.Errorf("public key has the wrong size"))
	}

	return struct {
		IsValid bool `json:"is_valid"`
	}{
		IsValid: ed25519.Verify(key, data, signature),
	}, nil
}

type errWrongPubKey struct {
	name   string
	pubkey string
}

func (e errWrongPubKey) Error() string {
	return fmt.Sprintf("user %s already exist in 3bot connect with another public key: %s", e.name, e.pubkey)
}

func (u *UserAPI) compareUser(user types.User) error {
	if u.threebotConnectAPIURL == "" {
		return nil
	}

	record := struct {
		Doublename string `json:"doublename"`
		PublicKey  string `json:"publickey"`
	}{}

	url := fmt.Sprintf("%s/%s", u.threebotConnectAPIURL, user.Name)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// user doesn't exist yet in 3bot connect, nothing to do
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error while checking 3botConnect API for duplicate user: wrong HTTP response status %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return err
	}

	pubKey, err := base64.StdEncoding.DecodeString(record.PublicKey)
	if err != nil {
		return err
	}

	hexPub := hex.EncodeToString(pubKey)
	if hexPub != user.Pubkey {
		return errWrongPubKey{
			name:   record.Doublename,
			pubkey: hexPub,
		}
	}

	return nil
}

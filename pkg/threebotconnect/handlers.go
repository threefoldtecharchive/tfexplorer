package threebotconnect

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"

	"github.com/agl/ed25519/extra25519"

	"github.com/gorilla/sessions"
)

type Auth struct {
	AppID      string
	PrivateKey ed25519.PrivateKey

	store *sessions.CookieStore
}

func New(pk ed25519.PrivateKey, appID string, sessionKey []byte) *Auth {
	return &Auth{
		AppID:      appID,
		PrivateKey: pk,
		store:      sessions.NewCookieStore(sessionKey),
	}
}

func (a *Auth) B64PublicKey() string {
	pub := a.PrivateKey.Public().(ed25519.PublicKey)
	return base64.RawStdEncoding.EncodeToString([]byte(pub))
}

func (a *Auth) verify(msg, sig []byte) bool {
	return ed25519.Verify(a.PrivateKey.Public().(ed25519.PublicKey), msg, sig)
}

func (a *Auth) Login(w http.ResponseWriter, r *http.Request) {
	// Public backend authenticator service
	authurl := "https://login.threefold.me"

	// Application id, this host will be used for callback url
	callback := "/explorer/auth/callback_threebot"

	//  State is a random string
	state := randomString(32)

	//  Encode payload with urlencode then passing data to the GET request
	v := url.Values{}
	v.Set("appid", a.AppID)
	v.Add("publickey", a.B64PublicKey())
	v.Add("state", state)
	v.Add("redirecturl", callback)

	url := fmt.Sprintf("%s?%s", authurl, v.Encode())

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusFound)
}

type userInfo struct {
	Doublename string `json:"doublename"`
	PublicKey  string `json:"publickey"`
}

func (u *userInfo) Curve25519PublicKey() (k [32]byte, err error) {
	b, err := base64.RawStdEncoding.DecodeString(u.PublicKey)
	if err != nil {
		return k, err
	}

	return publicKeyToCurve25519(ed25519.PublicKey(b)), nil
}

func publicKeyToCurve25519(pk ed25519.PublicKey) [32]byte {
	curvePub := [32]byte{}
	edPriv := [ed25519.PublicKeySize]byte{}
	copy(edPriv[:], pk)
	extra25519.PublicKeyToCurve25519(&curvePub, &edPriv)
	return curvePub
}

func (a *Auth) Callback(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("got redirect")
	q := r.URL.Query()
	cerr := q.Get("error")
	if cerr != "" {
		panic(cerr) //FIXME
		// return "Authentication failed: %s" % q.Get("error"), 400
	}

	// Signedhash contains state signed by user"s bot key
	sa := struct {
		DoubleName  string `json:"doubleName"`
		SignAttempt string `json:"signedAttempt"`
	}{}

	log.Info().Msgf("%+v", q)

	if err := json.Unmarshal([]byte(q.Get("signedAttempt")), &sa); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", sa)

	userInfo, err := getUserInfo(sa.DoubleName)
	if err != nil {
		panic(err)
	}

	userPub, err := base64.RawStdEncoding.DecodeString(userInfo.PublicKey)
	if err != nil {
		panic(err)
	}

	signature := []byte(sa.SignAttempt[:64])
	data := []byte(sa.SignAttempt[64:])
	if !ed25519.Verify(ed25519.PublicKey(userPub), data, signature) {
		panic("wrong signature")
	}

	fmt.Printf("%v", string(data))

	// var decryptNonce [24]byte
	// copy(decryptNonce[:], sa.SignAttempt[:24])

	// senderPublicKey, err := userInfo.Curve25519PublicKey()
	// if err != nil {
	// 	panic(err)
	// }

	// var recipientPrivateKey [32]byte
	// copy(recipientPrivateKey[:], a.PrivateKey)

	// payload, ok := box.Open(nil, sa.SignAttempt[24:], &decryptNonce, &senderPublicKey, &recipientPrivateKey)
	// if !ok {
	// 	panic("decryption error")
	// }
	// log.Info().Msg(string(payload))

	// values = json.loads(payload.decode("utf-8"))
	// if values["email"]["verified"] == None:
	// 	return "Email unverified, access denied.", 400

	// print("[+] threebot: user "%s" authenticated" % username)

	// session, _ := a.store.Get(r, email)
	// session.Values["authenticated"] = true
	// session.Values["username"] = username
	// session.Values["email"] = email

	// err = session.Save(r, w)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
}

func getUserInfo(username string) (u userInfo, err error) {
	// Fetching user"s bot information (including public key)
	resp, err := http.Get("https://login.threefold.me/api/users/" + username)
	if err != nil {
		panic(err)
	}

	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return u, err
	}
	return u, nil
}

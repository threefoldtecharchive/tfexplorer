//go:generate $GOPATH/bin/statik -f -src=../../dist -dest=../../
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rusart/muxprom"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	pprof "net/http/pprof"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/threefoldtech/tfexplorer/config"
	"github.com/threefoldtech/tfexplorer/mw"
	"github.com/threefoldtech/tfexplorer/pkg/capacity"
	capacitydb "github.com/threefoldtech/tfexplorer/pkg/capacity/types"
	"github.com/threefoldtech/tfexplorer/pkg/directory"
	"github.com/threefoldtech/tfexplorer/pkg/escrow"
	escrowdb "github.com/threefoldtech/tfexplorer/pkg/escrow/types"
	"github.com/threefoldtech/tfexplorer/pkg/phonebook"
	"github.com/threefoldtech/tfexplorer/pkg/stellar"
	"github.com/threefoldtech/tfexplorer/pkg/workloads"
	_ "github.com/threefoldtech/tfexplorer/statik"
	"github.com/threefoldtech/zos/pkg/app"
	"github.com/threefoldtech/zos/pkg/version"
)

func main() {
	app.Initialize()

	var (
		listen             string
		dbConf             string
		dbName             string
		seed               string
		foundationAddress  string
		threebotConnectURL string
		ver                bool
		flushEscrows       bool
		backupSigners      stellar.Signers
	)

	flag.StringVar(&listen, "listen", ":8080", "listen address, default :8080")
	flag.StringVar(&dbConf, "mongo", "mongodb://localhost:27017", "connection string to mongo database")
	flag.StringVar(&dbName, "name", "explorer", "database name")
	flag.StringVar(&seed, "seed", "", "wallet seed")
	flag.StringVar(&config.Config.Network, "network", "", "tfchain network")
	flag.StringVar(&foundationAddress, "foundation-address", "", "foundation address for the escrow foundation payment cut, if not set and the foundation should receive a cut from a resersvation payment, the wallet seed will receive the payment instead")
	flag.StringVar(&threebotConnectURL, "threebot-connect", "", "URL to the 3bot Connect app API. if specified, new user will be check against it and ensure public key are the same")
	flag.BoolVar(&ver, "v", false, "show version and exit")
	flag.Var(&backupSigners, "backupsigner", "reusable flag which adds a signer to the escrow accounts, we need atleast 5 signers to activate multisig")
	flag.BoolVar(&flushEscrows, "flush-escrows", false, "flush all escrows in the database, including currently active ones, and their associated addressses")

	flag.Parse()

	if ver {
		version.ShowAndExit(false)
	}

	if err := config.Valid(); err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
	}

	dropEscrow := false

	if flushEscrows {
		dropEscrow = userInputYesNo("Are you sure you want to drop all escrows and related addresses?") &&
			userInputYesNo("Are you REALLY sure you want to drop all escrows and related addresses?")
	}

	if flushEscrows && !dropEscrow {
		log.Fatal().Msg("user indicated he does not want to remove existing escrow information - please restart the explorer without the \"flush-escrows\" flag")
	}

	ctx := context.Background()
	client, err := connectDB(ctx, dbConf)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to connect to database")
	}

	s, err := createServer(listen, dbName, client, seed, foundationAddress, dropEscrow, backupSigners, threebotConnectURL)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to create HTTP server")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go s.ListenAndServe()

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Printf("error during server shutdown: %v\n", err)
	}
}

func connectDB(ctx context.Context, connectionURI string) (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(connectionURI))
	if err != nil {
		return nil, err
	}

	if err := client.Connect(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func createServer(listen, dbName string, client *mongo.Client, seed string, foundationAddress string, dropEscrowData bool, backupSigners stellar.Signers, threebotConnectURL string) (*http.Server, error) {
	db, err := mw.NewDatabaseMiddleware(dbName, client)
	if err != nil {
		return nil, err
	}

	var prom *muxprom.MuxProm
	router := mux.NewRouter()
	statikFS, err := fs.New()
	if err != nil {
		return nil, err
	}

	prom = muxprom.New(
		muxprom.Router(router),
		muxprom.Namespace("explorer"),
	)
	prom.Instrument()
	router.Use(db.Middleware)

	router.Path("/metrics").Handler(promhttp.Handler()).Name("metrics")
	router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(statikFS)))

	router.Path("/").HandlerFunc(serveStatic("/index.html", statikFS))
	router.Path("/explorer").HandlerFunc(serveStatic("/docs.html", statikFS))
	router.Path("/api/v1").HandlerFunc(serveStatic("/docs.html", statikFS))

	if dropEscrowData {
		log.Warn().Msg("dropping escrow and address collection")
		if err := db.Database().Collection(escrowdb.AddressCollection).Drop(context.Background()); err != nil {
			log.Fatal().Err(err).Msg("failed to drop address collection")
		}
		if err := db.Database().Collection(escrowdb.EscrowCollection).Drop(context.Background()); err != nil {
			log.Fatal().Err(err).Msg("failed to drop escrow collection")
		}
		log.Info().Msg("escrow and address collection dropped successfully. restart the explorer without \"flush-escrows\" flag")
		os.Exit(0)
	}

	var e escrow.Escrow
	if seed != "" {
		log.Info().Msgf("escrow enabled on %s", config.Config.Network)
		if err := escrowdb.Setup(context.Background(), db.Database()); err != nil {
			log.Fatal().Err(err).Msg("failed to create escrow database indexes")
		}

		wallet, err := stellar.New(seed, config.Config.Network, backupSigners)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create stellar wallet")
		}

		e = escrow.NewStellar(wallet, db.Database(), foundationAddress)

	} else {
		log.Info().Msg("escrow disabled")
		e = escrow.NewFree(db.Database())
	}

	go e.Run(context.Background())

	router.HandleFunc("/debug/pprof/profile", pprof.Profile)

	if err := directory.Setup(router, db.Database()); err != nil {
		log.Fatal().Err(err).Msg("failed to register directory package")
	}

	if err := phonebook.Setup(router, db.Database(), threebotConnectURL); err != nil {
		log.Fatal().Err(err).Msg("failed to register phonebook package")
	}

	if err = capacitydb.Setup(context.Background(), db.Database()); err != nil {
		log.Fatal().Err(err).Msg("failed to create capacity database indexes")
	}

	planner := capacity.NewNaivePlanner(e, db.Database())
	go planner.Run(context.Background())
	if err = workloads.Setup(router, db.Database(), e, planner); err != nil {
		log.Error().Err(err).Msg("failed to register workloads package")
	}

	log.Printf("start on %s\n", listen)
	r := handlers.LoggingHandler(os.Stderr, router)
	r = handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
		handlers.ExposedHeaders([]string{"Pages"}),
	)(r)

	return &http.Server{
		Addr:    listen,
		Handler: r,
	}, nil
}

func userInputYesNo(question string) bool {
	var reply string
	fmt.Printf("%s (yes/no): ", question)
	_, err := fmt.Scan(&reply)
	if err != nil {
		// if we can't read from the cmd line something is really really wrong
		log.Fatal().Err(err).Msg("could not read user reply from command line")
	}

	reply = strings.TrimSpace(reply)
	reply = strings.ToLower(reply)

	return reply == "y" || reply == "yes"
}

func serveStatic(path string, statikFS http.FileSystem) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		r, err := statikFS.Open(path)
		if err != nil {
			mw.Error(err, http.StatusInternalServerError)
			return
		}
		defer r.Close()
		w.WriteHeader(http.StatusOK)
		if _, err := io.Copy(w, r); err != nil {
			log.Error().Err(err).Send()
		}
	}
}

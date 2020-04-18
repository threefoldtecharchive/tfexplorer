module github.com/threefoldtech/tfexplorer

go 1.14

require (
	github.com/dave/jennifer v1.3.0
	github.com/emicklei/dot v0.10.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365
	github.com/jbenet/go-base58 v0.0.0-20150317085156-6237cf65f3a6
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	github.com/rakyll/statik v0.1.7
	github.com/rs/zerolog v1.18.0
	github.com/rusart/muxprom v0.0.0-20200323164249-36ea051efbe6
	github.com/stellar/go v0.0.0-20200325172527-9cabbc6b9388
	github.com/stretchr/testify v1.5.1
	github.com/threefoldtech/zos v0.2.3
	github.com/urfave/cli v1.22.3
	github.com/zaibon/httpsig v0.0.0-20200401133919-ea9cb57b0946
	go.mongodb.org/mongo-driver v1.3.2
	golang.org/x/crypto v0.0.0-20200311171314-f7b00557c8c4
	gotest.tools v2.2.0+incompatible
)

// replace github.com/threefoldtech/zos => ../zos

// replace github.com/threefoldtech/tfgateway => ../tf_gateway

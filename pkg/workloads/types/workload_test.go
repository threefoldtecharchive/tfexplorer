package types

import (
	"fmt"
	"testing"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"go.mongodb.org/mongo-driver/bson"
)

func TestEncoding(t *testing.T) {
	volume := workloads.Volume{}
	w := WorkloaderType{Workloader: &volume}

	b, err := bson.Marshal(w)
	if err != nil {
		fmt.Println(err)
		t.Fatal()
	}

	fmt.Println(string(b))
}

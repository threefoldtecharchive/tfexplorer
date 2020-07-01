package workloads

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	schema "github.com/threefoldtech/tfexplorer/schema"
)

func TestEncode(t *testing.T) {
	k := K8S{}
	k.SetID(schema.ID(666))

	err := json.NewEncoder(os.Stdout).Encode(k)
	if err != nil {
		fmt.Println(err)
	}
}

package sync

import (
	"fmt"
	"hash/fnv"

	"github.com/davecgh/go-spew/spew"
	"k8s.io/apimachinery/pkg/util/rand"
)

// ComputeHash returns a hash value calculated from pod template and
// a collisionCount to avoid hash collision. The hash will be safe encoded to
// avoid bad words.
//
// Proudly modified and adopted from k8s.io/kubernetes/pkg/util/hash.DeepHashObject and
// k8s.io/kubernetes/pkg/controller.ComputeHash.
func ComputeHash(template interface{}) string {
	hasher := fnv.New32a()

	hasher.Reset()

	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", template)

	return rand.SafeEncodeString(fmt.Sprint(hasher.Sum32()))
}

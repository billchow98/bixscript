// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package benchmark

import (
	"github.com/billchow98/bixscript/cmd/bix/internal/runtime"
	"log"
	"testing"
)

func benchmark(b *testing.B, src string, filename string) {
	for i := 0; i < b.N; i++ {
		v := runtime.New()
		errs := v.Run(src, filename)

		if errs != nil {
			for i, err := range errs {
				switch i {
				case len(errs) - 1:
					log.Fatalln(err)
				default:
					log.Println(err)
				}
			}
		}
	}
}

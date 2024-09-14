// © 2023 Bill Chow. All rights reserved.
// Unauthorized use, modification, or distribution of this code is strictly
// prohibited.

package lib

import "github.com/billchow98/bixscript/cmd/bix/internal/value"

type vmer interface {
	RegisterNativeFunction(name *value.Identifier, function value.Value)
}

package lib

import "github.com/billchow98/bixscript/cmd/bix/internal/value"

type vmer interface {
	RegisterNativeFunction(name *value.Identifier, function value.Value)
}

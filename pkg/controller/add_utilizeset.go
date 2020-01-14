package controller

import (
	"kube-utilize-operator/pkg/controller/utilizeset"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, utilizeset.Add)
}

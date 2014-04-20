package logo

import (
	"errors"
	"fmt"
)

func toError(code int, node Node, message string) error {
	if node != nil {
		line, col := node.position()
		return errors.New(fmt.Sprintf("%000d: (%d, %d), %s", code, line, col, message))
	}
	err := errors.New(fmt.Sprintf("%000d: %s", code, message))
	return err
}

func errorKeywordExpected(node Node, keyword string) error {
	return toError(1, node, "Keyword "+keyword+" expected.")
}

func errorWordExpected(node Node) error {
	return toError(2, node, "Word expected.")
}

func errorListExpected(node Node) error {
	return toError(3, node, "List expected.")
}

func errorProcedureNotFound(node Node, name string) error {
	return toError(4, node, "I don't know how to "+name+".")
}

func errorProcedureExpected(node Node) error {
	return toError(5, node, "Procedure expected.")
}

func errorNotEnoughParameters(node Node) error {
	return toError(6, node, "Not enough inputs.")
}

func errorNumberExpected(node Node) error {
	return toError(7, node, "Number expected.")
}

func errorBooleanExpected(node Node) error {
	return toError(8, node, "Boolean expected.")
}

func errorPositiveIntegerExpected(node Node) error {
	return toError(9, node, "Positive integer expected.")
}

func errorPositiveNumberExpected(node Node) error {
	return toError(10, node, "Positive number expected.")
}

func errorNoCurrentTestValue(node Node) error {
	return toError(11, node, "No current test value.")
}

func errorVariableNotFound(node Node, name string) error {
	return toError(12, node, name+" has no value.")
}

func errorReturnValueUnused(node Node, value Node) error {
	return toError(13, node, "You don't say what to do with "+value.String())
}

func errorBadInput(value Node) error {
	return toError(14, value, "I don't like "+value.String()+" as an input.")
}

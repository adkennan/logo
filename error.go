package main

import (
	"errors"
	"fmt"
)

func toError(code int, node Node, message string) error {
	return errors.New(message)
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
	return toError(5, node, "Procedure expected")
}

func errorNotEnoughParameters(caller *WordNode, node Node) error {
	return toError(6, node, "Not enough inputs to "+caller.value+".")
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

func errorReturnValueUnused(node Node) error {
	return toError(13, node, "You don't say what to do with "+node.String())
}

func errorBadInput(value Node) error {
	return toError(14, value, "I don't like "+value.String()+" as an input.")
}

func errorFileNotOpen(name string) error {
	return toError(15, nil, "File "+name+" is not open.")
}

func errorListOfNItemsExpected(node Node, n int) error {
	return toError(16, node, "Expected list of "+fmt.Sprint(n)+" items.")
}

func errorUnknownColor(node Node, name string) error {
	return toError(17, node, "I don't know the color "+name+".")
}

func errorNumberNotInRange(node Node, hi, low int) error {
	return toError(18, node, fmt.Sprintf("Expected a number between %d and %d.", hi, low))
}

func errorNotDir(path string) error {
	return toError(19, nil, path+" is not a directory.")
}

func errorNotFile(path string) error {
	return toError(20, nil, path+" is not a file.")
}

func errorNoInterpretedFrame(node *WordNode) error {
	return toError(21, node, "Can only use "+node.value+" inside a procedure.")
}

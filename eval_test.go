package main

import (
	"testing"
)

var ws *Workspace = CreateWorkspace()

func assertExpression(t *testing.T, expr, expectedVal string) {

	n, err := ParseString(expr)
	if err != nil {
		t.Error(err)
	}

	cr, _ := evaluateExpression(ws.currentFrame, n)

	if cr.err != nil {
		t.Error(cr.err)
	}

	if cr.returnValue == nil {
		t.Error("No return value.")
	}

	if cr.returnValue.String() != expectedVal {
		t.Errorf("Expected \"%s\" was \"%s\"", expectedVal, cr.returnValue.String())
	}
}

func TestExprAddTwoNumbers(t *testing.T) {

	assertExpression(t, "1 + 1", "2")
}

func TestExprAddThreeNumbers(t *testing.T) {

	assertExpression(t, "1 + 2 + 3", "6")
}

func TestExprAddFourNumbers(t *testing.T) {

	assertExpression(t, "1 + 2 + 3 + 4", "10")
}

func TestExprMulWithAdd(t *testing.T) {

	assertExpression(t, "2 + 3 * 4 + 5", "19")
}

func TestExprWithProcCall(t *testing.T) {

	assertExpression(t, "5 + MOD 10 2 + 1", "6")
}

func TestExprWithUnaryMinus(t *testing.T) {

	assertExpression(t, "10 * -(2*2)", "-40")
}

func TestExprWithNegative(t *testing.T) {

	assertExpression(t, "-4 * 10", "-40")
}

func TestExprWithSub(t *testing.T) {

	assertExpression(t, "5 * 5 - 20", "5")
}

func TestExprWithDiv(t *testing.T) {
	assertExpression(t, "5 * 10 / 5", "10")
}

package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type NodeType int

const (
	Word NodeType = iota
	List
	Group
)

type NodeEnumerator struct {
	nodes []Node
	c     Node
}

func EnumerateWords(node Node) *NodeEnumerator {
	return &NodeEnumerator{make([]Node, 0, 5), node}
}

func (this *NodeEnumerator) nextWord() *WordNode {

	n := this.c
	for {
		for n == nil {
			l := len(this.nodes)
			if l == 0 {
				return nil
			}
			n = this.nodes[l-1].next()
			this.nodes = this.nodes[0 : l-1]
		}
		switch nn := n.(type) {
		case *WordNode:
			this.c = n.next()
			return nn
		case *ListNode:
			this.nodes = append(this.nodes, n)
			n = nn.firstChild
		case *GroupNode:
			this.nodes = append(this.nodes, n)
			n = nn.firstChild
		}
	}

	return nil
}

type Node interface {
	nodeType() NodeType
	next() Node
	addNode(node Node)
	position() (int, int)
	String() string
	clone() Node
	setLiteral()
}

type BaseNode struct {
	next Node
	line int
	col  int
}

type WordNode struct {
	BaseNode
	value          string
	isLiteral      bool
	isFirstOfGroup bool
}

func newWordNode(line, col int, value string, isLiteral bool) *WordNode {
	n := &WordNode{}
	n.BaseNode.line = line
	n.BaseNode.col = col
	n.value = value
	n.isLiteral = isLiteral

	return n
}

func (this *WordNode) nodeType() NodeType { return Word }

func (this *WordNode) next() Node { return this.BaseNode.next }

func (this *WordNode) addNode(node Node) {
	this.BaseNode.next = node
}

func (this *WordNode) position() (int, int) { return this.line, this.col }

func (this *WordNode) String() string {
	return this.value
}

func (this *WordNode) clone() Node {
	n := newWordNode(this.line, this.col, this.value, this.isLiteral)
	n.isFirstOfGroup = this.isFirstOfGroup
	return n
}

func (this *WordNode) setLiteral() { this.isLiteral = true }

type ListNode struct {
	BaseNode
	firstChild Node
}

func newListNode(line, col int, firstChild Node) *ListNode {
	n := &ListNode{}
	n.BaseNode.line = line
	n.BaseNode.col = col
	n.firstChild = firstChild

	return n
}

func (this *ListNode) nodeType() NodeType { return List }

func (this *ListNode) next() Node { return this.BaseNode.next }

func (this *ListNode) addNode(node Node) {
	this.BaseNode.next = node
}

func (this *ListNode) position() (int, int) { return this.line, this.col }

func (this *ListNode) String() string {
	s := "[ "
	n := this.firstChild
	for n != nil {
		s += n.String() + " "
		n = n.next()
	}
	s += "]"

	return s
}

func (this *ListNode) length() int {
	n := 0
	for c := this.firstChild; c != nil; c = c.next() {
		n++
	}
	return n
}

func (this *ListNode) clone() Node {
	var fn Node
	if this.firstChild != nil {
		fn = this.firstChild.clone()
		var nn = fn
		for on := this.firstChild.next(); on != nil; on = on.next() {
			nn.addNode(on.clone())
			nn = nn.next()
		}
	}

	return newListNode(this.line, this.col, fn)
}

func (this *ListNode) setLiteral() {

	for n := this.firstChild; n != nil; n = n.next() {
		n.setLiteral()
	}
}

type GroupNode struct {
	BaseNode
	firstChild Node
}

func newGroupNode(line, col int, firstChild Node) *GroupNode {
	n := &GroupNode{}
	n.BaseNode.line = line
	n.BaseNode.col = col
	n.firstChild = firstChild

	return n
}

func (this *GroupNode) nodeType() NodeType { return Group }

func (this *GroupNode) next() Node { return this.BaseNode.next }

func (this *GroupNode) addNode(node Node) {
	this.BaseNode.next = node
}

func (this *GroupNode) position() (int, int) { return this.line, this.col }

func (this *GroupNode) String() string {
	s := "( "
	n := this.firstChild
	for n != nil {
		s += n.String() + " "
		n = n.next()
	}
	s += ")"

	return s
}

func (this *GroupNode) length() int {
	n := 0
	for c := this.firstChild; c != nil; c = c.next() {
		n++
	}
	return n
}

func (this *GroupNode) clone() Node {
	var fn Node
	if this.firstChild != nil {
		fn = this.firstChild.clone()
		var nn = fn
		for on := this.firstChild.next(); on != nil; on = on.next() {
			nn.addNode(on.clone())
			nn = nn.next()
		}
	}

	return newListNode(this.line, this.col, fn)
}

func (this *GroupNode) setLiteral() {

	for n := this.firstChild; n != nil; n = n.next() {
		n.setLiteral()
	}
}

func printNode(ws *Workspace, n Node, includeBrackets bool) {
	buf := &bytes.Buffer{}

	nodeToText(buf, n, includeBrackets)

	ws.print(buf.String())
}

func printLine(ws *Workspace, firstNode, currentNode Node) {

	buf := &bytes.Buffer{}

	fl, _ := firstNode.position()
	n := firstNode
	l, _ := n.position()
	for n != nil && l == fl {
		buf.WriteString(" ")
		if n == currentNode {
			buf.WriteString(">")
		}
		nodeToText(buf, n, true)
		n = n.next()
		if n != nil {
			l, _ = n.position()
		}
	}

	buf.WriteString("\n")
	ws.print(buf.String())
}

func nodeToText(buf *bytes.Buffer, n Node, includeBrackets bool) {

	switch pn := n.(type) {
	case *WordNode:
		buf.WriteString(pn.value)

	case *ListNode:
		if includeBrackets {
			buf.WriteString("[ ")
		}
		for nn := pn.firstChild; nn != nil; nn = nn.next() {
			nodeToText(buf, nn, true)
			buf.WriteString(" ")
		}
		if includeBrackets {
			buf.WriteString("]")
		}
	}
}

func evalToWord(node Node) (string, error) {

	s, _, err := readWordValue(node)
	return s, err
}

func isWordNodeWithValue(n Node, val string) bool {
	switch wn := n.(type) {
	case *WordNode:
		return strings.ToUpper(wn.value) == val
	}
	return false
}

func readWordValue(node Node) (string, Node, error) {
	wn := node.(*WordNode)

	if wn == nil {
		return "", nil, errorWordExpected(node)
	}

	return wn.value, wn.next(), nil
}

func evalToNumber(node Node) (float64, error) {

	switch pn := node.(type) {
	case *WordNode:
		r, err := strconv.ParseFloat(pn.value, 64)
		if err != nil {
			return 0, errorBadInput(node)
		}
		return r, nil
	}
	return 0, errorNumberExpected(node)
}

func evalNumericParams(nx, ny Node) (float64, float64, error) {

	x, err := evalToNumber(nx)
	if err != nil {
		return 0, 0, err
	}

	y, err := evalToNumber(ny)
	if err != nil {
		return 0, 0, err
	}

	return x, y, nil
}

func evalToBoolean(node Node) (bool, error) {
	switch pn := node.(type) {
	case *WordNode:
		switch strings.ToUpper(pn.value) {
		case keywordTrue:
			return true, nil
		case keywordFalse:
			return false, nil
		}
	}
	return false, errorBooleanExpected(node)
}

func evalList(frame Frame, node *ListNode) (*ListNode, error) {

	var fn Node
	var ln Node
	var rv *CallResult
	n := node.firstChild
	for n != nil {
		rv, n = evaluateNode(frame, n, true)
		if rv.hasError() {
			return nil, rv.err
		}
		if fn == nil {
			fn = rv.returnValue
		} else {
			ln.addNode(rv.returnValue)
		}
		ln = rv.returnValue
	}

	nln := newListNode(-1, -1, fn)

	return nln, nil
}

func nodesEqual(x, y Node, numericCompare bool) bool {
	if x.nodeType() != y.nodeType() {
		return false
	}

	switch x.nodeType() {
	case Word:
		wx := x.(*WordNode)
		wy := y.(*WordNode)

		if numericCompare {
			nx, ex := evalToNumber(wx)
			ny, ey := evalToNumber(wy)

			if ex == nil && ey == nil {
				return nx == ny
			}
		}

		return strings.ToUpper(wx.value) == strings.ToUpper(wy.value)

	case List:
		lx := x.(*ListNode)
		ly := y.(*ListNode)

		if ly.length() != lx.length() {
			return false
		}

		cx := lx.firstChild
		cy := ly.firstChild
		for cx != nil && cy != nil {

			if !nodesEqual(cx, cy, numericCompare) {
				return false
			}

			cx = cx.next()
			cy = cy.next()
		}
	}

	return true
}

func createNumericNode(n float64) Node {
	return newWordNode(-1, -1, fmt.Sprint(n), true)
}

func dumpNodes(frame Frame, n Node, max int) {

	for n != nil && max >= 0 {
		frame.workspace().print(n.String() + " ")
		n = n.next()
		max--
	}
	frame.workspace().print("\n")
}

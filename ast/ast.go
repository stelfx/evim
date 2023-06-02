package ast

import (
	"math"
	"math/rand"
	"reflect"
	"strconv"
	noise "github.com/Go_Projects/simplex_noise"
)

type Node interface {
	Eval(x, y float32) float32
	String() string
	AddRandom(node Node)
	SetParent(parent Node)
	AddLeaf(leaf Node) bool
	NodeCount() int
	GetChildren() []Node
	SetChildren([]Node)
	GetParent() Node
}

type BaseNode struct {
	Parent   Node
	Children []Node
}

type OpPicture struct {
	BaseNode
}

func NewOpPicture() *OpPicture {
	return &OpPicture{BaseNode{nil, make([]Node, 3)}}
}

func (op *OpPicture) Eval(x, y float32) float32 {
	panic("eval called on root of a picture tree")
}

func (op *OpPicture) String() string {
	return "( Picture\n" + op.Children[0].String() + "\n" + op.Children[1].String() + "\n" + op.Children[2].String() + " )"
}

func CopyTree(node Node, parent Node) Node {
	copy := reflect.New(reflect.ValueOf(node).Elem().Type()).Interface().(Node)

	switch n := node.(type) {
	case *OpConstant:
		copy.(*OpConstant).value = n.value
	}

	copy.SetParent(parent)
	copyChildren := make([]Node, len(node.GetChildren()))
	copy.SetChildren(copyChildren)
	for i := range copyChildren {
		copyChildren[i] = CopyTree(node.GetChildren()[i], copy)
	}
	return copy
}

func ReplaceNode(old Node, new Node) {
	oldParent := old.GetParent()
	if oldParent != nil {
		for i, child := range oldParent.GetChildren() {
			if child == old {
				oldParent.GetChildren()[i] = new
			}
		}
	}
	new.SetParent(oldParent)
}

func (node *BaseNode) GetChildren() []Node {
	return node.Children
}

func (node *BaseNode) SetChildren(children []Node) {
	node.Children = children
}

func GetNthNode(node Node, n, count int) (Node, int) {
	if n == count {
		return node, count
	}
	var result Node
	for _, child := range node.GetChildren() {
		count++
		result, count = GetNthNode(child, n, count)
		if result != nil {
			return result, count
		}
	}
	return nil, count
}

func (node *BaseNode) GetParent() Node {
	return node.Parent
}

func Mutate(node Node) Node {
	r := rand.Intn(24)

	var mutatedNode Node

	if r <= 21 {
		mutatedNode = GetRandomBaseNode()
	} else {
		mutatedNode = GetRandomLeaf()
	}

	if node.GetParent() != nil {
		for i, parentChild := range node.GetParent().GetChildren() {
			if parentChild == node {
				node.GetParent().GetChildren()[i] = mutatedNode
			}
		}
	}

	for i, child := range node.GetChildren() {
		if i >= len(mutatedNode.GetChildren()) {
			break
		}
		mutatedNode.GetChildren()[i] = child
		child.SetParent(mutatedNode)
	}

	for i, child := range mutatedNode.GetChildren() {
		if child == nil {
			leaf := GetRandomLeaf()
			leaf.SetParent(mutatedNode)
			mutatedNode.GetChildren()[i] = leaf
		}
	}

	mutatedNode.SetParent(node.GetParent())
	return mutatedNode

}

func (node *BaseNode) NodeCount() int {
	count := 1
	for _, child := range node.Children {
		count += child.NodeCount()
	}
	return count
}

func (node *BaseNode) Eval(x, y float32) float32 {
	panic("tried to call evall on basenode")
}

func (node *BaseNode) String() string {
	panic("tried to call string on basenode")
}

func (node *BaseNode) SetParent(parent Node) {
	node.Parent = parent
}

func (node *BaseNode) AddRandom(nodeToAdd Node) {
	addIndex := rand.Intn(len(node.Children))

	if node.Children[addIndex] == nil {
		nodeToAdd.SetParent(node)
		node.Children[addIndex] = nodeToAdd
	} else {
		node.Children[addIndex].AddRandom(nodeToAdd)
	}
}

func (node *BaseNode) AddLeaf(leaf Node) bool {
	for i, child := range node.Children {
		if child == nil {
			leaf.SetParent(node)
			node.Children[i] = leaf
			return true
		} else if node.Children[i].AddLeaf(leaf) {
			return true
		}
	}
	return false
}

type OpNegate struct {
	BaseNode
}

func NewOpNegate() *OpNegate {
	return &OpNegate{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpNegate) Eval(x, y float32) float32 {
	return -op.Children[0].Eval(x, y)
}

func (op *OpNegate) String() string {
	return "( Negate " + op.Children[0].String() + " )"
}

type OpCeil struct {
	BaseNode
}

func NewOpCeil() *OpCeil {
	return &OpCeil{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpCeil) Eval(x, y float32) float32 {

	return float32(math.Ceil(float64(op.Children[0].Eval(x, y))))
}

func (op *OpCeil) String() string {
	return "( Ceil " + op.Children[0].String() + " )"
}

type OpPlus struct {
	BaseNode
}

func NewOpPlus() *OpPlus {
	return &OpPlus{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpPlus) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) + op.Children[1].Eval(x, y)
}

func (op *OpPlus) String() string {
	return "( + " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpMinus struct {
	BaseNode
}

func NewOpMinus() *OpMinus {
	return &OpMinus{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpMinus) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) - op.Children[1].Eval(x, y)
}

func (op *OpMinus) String() string {
	return "( - " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpMult struct {
	BaseNode
}

func NewOpMult() *OpMult {
	return &OpMult{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpMult) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) * op.Children[1].Eval(x, y)
}

func (op *OpMult) String() string {
	return "( * " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpDiv struct {
	BaseNode
}

func NewOpDiv() *OpDiv {
	return &OpDiv{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpDiv) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) / op.Children[1].Eval(x, y)
}

func (op *OpDiv) String() string {
	return "( / " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpX struct {
	BaseNode
}

func NewOpX() *OpX {
	return &OpX{BaseNode{nil, make([]Node, 0)}}
}

func (op *OpX) Eval(x, y float32) float32 {
	return x
}

func (op *OpX) String() string {
	return "X"
}

type OpY struct {
	BaseNode
}

func NewOpY() *OpY {
	return &OpY{BaseNode{nil, make([]Node, 0)}}
}

func (op *OpY) Eval(x, y float32) float32 {
	return y
}

func (op *OpY) String() string {
	return "Y"
}

type OpClip struct {
	BaseNode
}

func NewOpClip() *OpClip {
	return &OpClip{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpClip) Eval(x, y float32) float32 {
	value := op.Children[0].Eval(x, y)
	max := float32(math.Abs(float64(op.Children[1].Eval(x, y))))
	if value > max {
		return max
	} else if value < -max {
		return -max
	}
	return value
}

func (op *OpClip) String() string {
	return "( Clip " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpConstant struct {
	BaseNode
	value float32
}

func NewOpConstant() *OpConstant {
	return &OpConstant{BaseNode{nil, make([]Node, 0)}, rand.Float32()*2 - 1}
}

func (op *OpConstant) Eval(x, y float32) float32 {
	return op.value
}

func (op *OpConstant) String() string {
	return strconv.FormatFloat(float64(op.value), 'f', 9, 32)
}

type OpSquare struct {
	BaseNode
}

func NewOpSquare() *OpSquare {
	return &OpSquare{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpSquare) Eval(x, y float32) float32 {
	square := op.Children[0].Eval(x, y) * op.Children[1].Eval(x, y)

	return square * square
}

func (op *OpSquare) String() string {
	return "( Square " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpLerp struct {
	BaseNode
}

func NewOpLerp() *OpLerp {
	return &OpLerp{BaseNode{nil, make([]Node, 3)}}
}

func (op *OpLerp) Eval(x, y float32) float32 {
	a := op.Children[0].Eval(x, y)
	b := op.Children[1].Eval(x, y)
	p := op.Children[2].Eval(x, y)

	return a + p*(b-a)
}

func (op *OpLerp) String() string {
	return "( Lerp " + op.Children[0].String() + " " + op.Children[1].String() + " " + op.Children[2].String() + " )"
}

type OpSin struct {
	BaseNode
}

func NewOpSin() *OpSin {
	return &OpSin{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpSin) Eval(x, y float32) float32 {
	return float32(math.Sin(float64(op.Children[0].Eval(x, y))))
}

func (op *OpSin) String() string {
	return "( Sin " + op.Children[0].String() + " )"
}

type OpCos struct {
	BaseNode
}

func NewOpCos() *OpCos {
	return &OpCos{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpCos) Eval(x, y float32) float32 {
	return float32(math.Cos(float64(op.Children[0].Eval(x, y))))
}

func (op *OpCos) String() string {
	return "( Cos " + op.Children[0].String() + " )"
}

type OpFloor struct {
	BaseNode
}

func NewOpFloor() *OpFloor {
	return &OpFloor{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpFloor) Eval(x, y float32) float32 {
	return float32(math.Floor(float64(op.Children[0].Eval(x, y))))
}

func (op *OpFloor) String() string {
	return "( Floor " + op.Children[0].String() + " )"
}

type OpLog struct {
	BaseNode
}

func NewOpLog() *OpLog {
	return &OpLog{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpLog) Eval(x, y float32) float32 {
	return float32(math.Log2(float64(op.Children[0].Eval(x, y))))
}

func (op *OpLog) String() string {
	return "( Log " + op.Children[0].String() + " )"
}

type OpWrap struct {
	BaseNode
}

func NewOpWrap() *OpWrap {
	return &OpWrap{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpWrap) Eval(x, y float32) float32 {
	f := op.Children[0].Eval(x, y)
	temp := (f - 1.0) / 2.0
	return -1.0 + 2.0*(temp-float32(math.Floor(float64(temp))))
}

func (op *OpWrap) String() string {
	return "(Wrap " + op.Children[0].String() + ")"
}

type OpAbs struct {
	BaseNode
}

func NewOpAbs() *OpAbs {
	return &OpAbs{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpAbs) Eval(x, y float32) float32 {
	return float32(math.Abs(float64(op.Children[0].Eval(x, y))))
}

func (op *OpAbs) String() string {
	return "( Abs " + op.Children[0].String() + " )"
}

type OpAtan struct {
	BaseNode
}

func NewOpAtan() *OpAtan {
	return &OpAtan{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpAtan) Eval(x, y float32) float32 {
	return float32(math.Atan(float64(op.Children[0].Eval(x, y))))
}

func (op *OpAtan) String() string {
	return "( Atan " + op.Children[0].String() + " )"
}

type OpNoise struct {
	BaseNode
}

func NewOpNoise() *OpNoise {
	return &OpNoise{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpNoise) Eval(x, y float32) float32 {
	return 80*noise.Snoise2(op.Children[0].Eval(x, y), op.Children[1].Eval(x, y)) - 2.0
}

func (op *OpNoise) String() string {
	return "( Noise " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpFBM struct {
	BaseNode
}

func NewOpFBM() *OpFBM {
	return &OpFBM{BaseNode{nil, make([]Node, 3)}}
}

func (op *OpFBM) Eval(x, y float32) float32 {
	return 2*3.627*noise.Fbm(op.Children[0].Eval(x, y), op.Children[1].Eval(x, y), 5*op.Children[2].Eval(x, y), 0.5, 2, 3) + .492 - 1
}

func (op *OpFBM) String() string {
	return "( FBM " + op.Children[0].String() + " " + op.Children[1].String() + " " + op.Children[2].String() + " )"
}

type OpTurbulence struct {
	BaseNode
}

func NewOpTurbulence() *OpTurbulence {
	return &OpTurbulence{BaseNode{nil, make([]Node, 3)}}
}

func (op *OpTurbulence) Eval(x, y float32) float32 {
	return 2*6.96*noise.Turbulence(op.Children[0].Eval(x, y), op.Children[1].Eval(x, y), 5*op.Children[2].Eval(x, y), 0.5, 2, 3) - 1
}

func (op *OpTurbulence) String() string {
	return "( Turbulence " + op.Children[0].String() + " " + op.Children[1].String() + " " + op.Children[2].String() + " )"
}

type OpGamma struct {
	BaseNode
}

func NewOpGamma() *OpGamma {
	return &OpGamma{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpGamma) Eval(x, y float32) float32 {
	return float32(math.Gamma(float64(op.Children[0].Eval(x, y))))
}

func (op *OpGamma) String() string {
	return "( Gamma " + op.Children[0].String() + " )"
}

type OpHypot struct {
	BaseNode
}

func NewOpHypot() *OpHypot {
	return &OpHypot{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpHypot) Eval(x, y float32) float32 {
	return float32(math.Hypot(float64(op.Children[0].Eval(x, y)), float64(op.Children[1].Eval(x, y))))
}

func (op *OpHypot) String() string {
	return "( Hypot " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

func GetRandomBaseNode() Node {
	r := rand.Intn(21)
	switch r {
	case 0:
		return NewOpClip()
	case 1:
		return NewOpNegate()
	case 2:
		return NewOpMult()
	case 3:
		return NewOpPlus()
	case 4:
		return NewOpCeil()
	case 5:
		return NewOpMinus()
	case 6:
		return NewOpDiv()
	case 7:
		return NewOpSquare()
	case 8:
		return NewOpLerp()
	case 9:
		return NewOpSin()
	case 10:
		return NewOpCos()
	case 11:
		return NewOpFloor()
	case 12:
		return NewOpLog()
	case 13:
		return NewOpWrap()
	case 14:
		return NewOpAbs()
	case 15:
		return NewOpAtan()
	case 16:
		return NewOpNoise()
	case 17:
		return NewOpFBM()
	case 18:
		return NewOpTurbulence()
	case 19:
		return NewOpGamma()
	case 20:
		return NewOpHypot()
	}
	panic("Get Random Double Node Failed!")
}

func GetRandomLeaf() Node {
	r := rand.Intn(3)
	switch r {
	case 0:
		return NewOpConstant()
	case 1:
		return NewOpX()
	case 2:
		return NewOpY()
	}
	panic("Get Leaf Node Failed!")
}

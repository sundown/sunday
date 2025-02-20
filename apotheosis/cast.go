package apotheosis

import (
	"github.com/sundown/solution/prism"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

func (env *Environment) castInt(from prism.Value) value.Value {
	switch from.Type.Kind() {
	case prism.TypeInt:
		return from.Value
	case prism.TypeBool, prism.TypeChar:
		return env.Block.NewSExt(from.Value, types.I64)
	}

	prism.Panic("Unreachable")
	panic("unlabelled error")
}

func (env *Environment) castReal(from prism.Value) value.Value {
	switch from.Type.Kind() {
	case prism.TypeBool, prism.TypeChar, prism.TypeInt:
		return env.Block.NewSIToFP(from.Value, types.Double)
	case prism.TypeReal:
		return from.Value
	}

	prism.Panic("Unreachable")
	panic("unlabelled error")
}

func (env *Environment) castChar(from prism.Value) value.Value {
	switch from.Type.Kind() {
	case prism.TypeChar:
		return from.Value
	case prism.TypeBool:
		return env.Block.NewSExt(from.Value, types.I8)
	}

	prism.Panic("Unreachable")
	panic("unlabelled error")
}

func (env *Environment) castBool(from prism.Value) value.Value {
	switch from.Type.Kind() {
	case prism.TypeBool:
		return from.Value
	}

	prism.Panic("Unreachable")

	panic("unlabelled error")
}

func (env *Environment) newCast(cast prism.Cast) value.Value {
	val := prism.Value{Value: env.newExpression(&cast.Value), Type: cast.Value.Type()}
	var castfn prism.MCallable
	var from prism.Type
	pred := false
	if _, ok := cast.Value.Type().(prism.VectorType); ok {
		from = cast.ToType.(prism.VectorType).Type
		pred = true
	} else {
		from = cast.ToType
	}

	switch from.Kind() {
	case prism.TypeInt:
		castfn = env.castInt
	case prism.TypeReal:
		castfn = env.castReal
	case prism.TypeBool:
		castfn = env.castBool
	case prism.TypeChar:
		castfn = env.castChar
	default:
		prism.Panic("Unreachable")
	}

	if pred {
		return env.vectorCast(castfn, val, cast.ToType.(prism.VectorType).Type)
	}

	return castfn(val)
}

func (env *Environment) vectorCast(caster prism.MCallable, vec prism.Value, to prism.Type) value.Value {
	elmType := vec.Type.(prism.VectorType).Type.Realise()
	irToHeadType := prism.VectorType{Type: to}
	toHeadType := irToHeadType.Realise()
	toElmType := to.Realise()
	leng := env.readVectorLength(vec)

	var head *ir.InstAlloca
	var body *ir.InstBitCast

	cap := env.readVectorCapacity(vec)
	head = env.Block.NewAlloca(toHeadType)

	env.writeVectorLength(prism.Value{Value: head, Type: irToHeadType}, leng)
	env.writeVectorCapacity(prism.Value{Value: head, Type: irToHeadType}, cap)

	// Allocate a body of capacity * element width, and cast to element type
	body = env.Block.NewBitCast(
		env.Block.NewCall(env.getCalloc(),
			i32(to.Width()), // Byte size of elements
			cap),            // How much memory to alloc
		types.NewPointer(toElmType)) // Cast alloc'd memory to type

	// --- Loop body ---
	vecBody := env.Block.NewLoad(
		types.NewPointer(elmType),
		env.Block.NewGetElementPtr(vec.Type.Realise(), vec.Value, i32(0), vectorBodyOffset))

	counter := env.Block.NewAlloca(types.I32)
	env.Block.NewStore(i32(0), counter)

	// get elem, add to accum, increment counter, conditional jump to body
	loopblock := env.newBlock(env.CurrentFunction)
	env.Block.NewBr(loopblock)
	env.Block = loopblock
	// Add to accum
	curCounter := loopblock.NewLoad(types.I32, counter)

	var curElm value.Value = loopblock.NewGetElementPtr(elmType, vecBody, curCounter)

	if _, ok := vec.Type.(prism.VectorType).Type.(prism.AtomicType); ok {
		curElm = loopblock.NewLoad(elmType, curElm)
	}

	loopblock.NewStore(
		caster(prism.Value{
			Value: curElm,
			Type:  vec.Type.(prism.VectorType).Type}),
		loopblock.NewGetElementPtr(toElmType, body, curCounter))

	incr := loopblock.NewAdd(curCounter, i32(1))

	loopblock.NewStore(incr, counter)

	exitblock := env.newBlock(env.CurrentFunction)

	loopblock.NewCondBr(loopblock.NewICmp(enum.IPredSLT, incr, leng), loopblock, exitblock)

	env.Block = exitblock

	env.writeVectorPointer(prism.Val(head, irToHeadType), body)

	return head
}

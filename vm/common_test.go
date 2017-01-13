package vm

import (
	"math/big"
	"fmt"
)

func ExampleCommon(){

	var c BigIntSorter

	c = append( c, *big.NewInt(1) )
	c = append( c, *big.NewInt(2) )
	c = append( c, *big.NewInt(3) )
	fmt.Println( "Len() test:", c.Len() )

	fmt.Println( "Swap() test:", c )
	c.Swap( 1, 3 )
	fmt.Println( "Swap() test:", c )

	fmt.Println( "Less() test:", c.Less( 0, 1 ) )

	sum := SumBigInt( c )
	fmt.Println( "SumBigInt() test:", sum )

	min := MinBigInt(c)
	fmt.Println( "MinBigInt() test:", min )

	max := MaxBigInt(c)
	fmt.Println( "MaxBigInt() test:", max )

	d := []int64{300,100,200}

	dmin := MinInt64( d )
	fmt.Println( "MinInt64() test:", dmin )

	dmax := MaxInt64( d )
	fmt.Println( "MaxInt64() test:", dmax )

	b1 := []byte{ 1, 2, 3 }
	b2 := []byte{ 1, 2, 3 }
	//b2 := []byte{ 5, 6, 7, 8, 9 }
	b3 := Concat( b1, b2 )
	fmt.Println( "Concat() test:", b3 )

	fmt.Println( "BigIntsOp() test:", c )
	bi := BigIntsOp( c, OP_1ADD )
	fmt.Println( "BigIntsOp() 1ADD test:", bi )
	bi = BigIntsOp( bi, OP_1SUB )
	fmt.Println( "BigIntsOp() 1SUB test:", bi )
	bi = BigIntsOp( bi, OP_2MUL )
	fmt.Println( "BigIntsOp() 2MUL test:", bi )
	bi = BigIntsOp( bi, OP_2DIV )
	fmt.Println( "BigIntsOp() 2DIV test:", bi )
	bi = BigIntsOp( bi, OP_NEGATE )
	fmt.Println( "BigIntsOp() NEGATE test:", bi )
	bi = BigIntsOp( bi, OP_ABS )
	fmt.Println( "BigIntsOp() ABS test:", bi )

	bi = BigIntOp( c, OP_1ADD )
	fmt.Println( "BigIntOp() 1ADD test:", bi )
	bi = BigIntOp( bi, OP_1SUB )
	fmt.Println( "BigIntOp() 1SUB test:", bi )
	bi = BigIntOp( bi, OP_2MUL )
	fmt.Println( "BigIntOp() 2MUL test:", bi )
	bi = BigIntOp( bi, OP_2DIV )
	fmt.Println( "BigIntOp() 2DIV test:", bi )
	bi = BigIntOp( bi, OP_NEGATE )
	fmt.Println( "BigIntOp() NEGATE test:", bi )
	bi = BigIntOp( bi, OP_ABS )
	fmt.Println( "BigIntOp() ABS test:", bi )

	fmt.Println( "AsBool() test:", AsBool(b3) )

	b4 := []byte{1,2,3,4,5,6,7,8}
	i6,_ := AsInt64( b4 )
	fmt.Println( "AsInt64() test:", i6 )

	s1 := [][]byte{ {1,2,3}, {10,20,30}, {100,200,201} }
	s2 := [][]byte{ {4,5,6}, {40,50,60}, {202,203,204} }
	s3 := ByteArrZip( s1, s2, OP_CONCAT )
	fmt.Println( "ByteArrZip() test:", s3 )

	var it1 []big.Int
	var it2 []big.Int
	it1  = append( it1, *big.NewInt(1) )
	it1  = append( it1, *big.NewInt(2) )
	it1  = append( it1, *big.NewInt(3) )
	it2  = append( it2, *big.NewInt(5) )
	it2  = append( it2, *big.NewInt(6) )
	it2  = append( it2, *big.NewInt(7) )

	fmt.Println( "BigIntZip() test:", it1 )
	fmt.Println( "BigIntZip() test:", it2 )
	it3 := BigIntZip( it1, it2, OP_AND )
	fmt.Println( "BigIntZip() OP_AND test:", it3 )
	it3 = BigIntZip( it1, it2, OP_OR )
	fmt.Println( "BigIntZip() OP_OR test:", it3 )
	it3 = BigIntZip( it1, it2, OP_XOR )
	fmt.Println( "BigIntZip() OP_XOR test:", it3 )
	it3 = BigIntZip( it1, it2, OP_ADD )
	fmt.Println( "BigIntZip() OP_ADD test:", it3 )
	it3 = BigIntZip( it1, it2, OP_SUB )
	fmt.Println( "BigIntZip() OP_SUB test:", it3 )

	fmt.Println( "BigIntsComp() test:", BigIntsComp( it1, OP_0NOTEQUAL ) )


	bl1 := []bool{true,true,false,false,false}
	bl2 := []bool{false,true,false,true,true}
	fmt.Println( "BoolsZip() OP_BOOLAND test:", BoolsZip( bl1, bl2, OP_BOOLAND ) )
	fmt.Println( "BoolsZip() OP_BOOLOR test:", BoolsZip( bl1, bl2, OP_BOOLOR ) )

	fmt.Println( "BoolArrayOp() test:", BoolArrayOp( bl1, OP_NOT ) )

	fmt.Println( "IsEqualBytes() test:", IsEqualBytes( b1, b2 ) )
	fmt.Println( "IsEqual() test:", IsEqual( b1, b2 ) )

	// output: ok
}
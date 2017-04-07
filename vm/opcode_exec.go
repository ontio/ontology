package vm

type OpExec struct {
	Opcode OpCode
	Name   string
	Exec   func(*ExecutionEngine) (VMState, error)
}

var (
	OpExecList = [256]OpExec{
		// control flow
		PUSH0:       {PUSH0, "0", opPushData},
		PUSHBYTES1:  {PUSHBYTES1, "PUSHBYTES1", opPushData},
		PUSHBYTES75: {PUSHBYTES75, "PUSHBYTES75", opPushData},
		PUSHDATA1:   {PUSHDATA1, "PUSHDATA1", opPushData},
		PUSHDATA2:   {PUSHDATA2, "PUSHDATA2", opPushData},
		PUSHDATA4:   {PUSHDATA4, "PUSHDATA4", opPushData},
		PUSHM1:      {PUSHM1, "PUSHM1", opPushData},
		PUSH1:       {PUSH1, "1", opPushData},
		PUSH2:       {PUSH2, "2", opPushData},
		PUSH3:       {PUSH3, "3", opPushData},
		PUSH4:       {PUSH4, "4", opPushData},
		PUSH5:       {PUSH5, "5", opPushData},
		PUSH6:       {PUSH6, "6", opPushData},
		PUSH7:       {PUSH7, "7", opPushData},
		PUSH8:       {PUSH8, "8", opPushData},
		PUSH9:       {PUSH9, "9", opPushData},
		PUSH10:      {PUSH10, "10", opPushData},
		PUSH11:      {PUSH11, "11", opPushData},
		PUSH12:      {PUSH12, "12", opPushData},
		PUSH13:      {PUSH13, "13", opPushData},
		PUSH14:      {PUSH14, "14", opPushData},
		PUSH15:      {PUSH15, "15", opPushData},
		PUSH16:      {PUSH16, "16", opPushData},

		//Control
		NOP:      {NOP, "NOP", opNop},
		JMP:      {JMP, "JMP", opJmp},
		JMPIF:    {JMPIF, "JMPIF", opJmp},
		JMPIFNOT: {JMPIFNOT, "JMPIFNOT", opJmp},
		CALL:     {CALL, "CALL", opCall},
		RET:      {RET, "RET", opRet},
		APPCALL:  {APPCALL, "APPCALL", opAppCall},
		SYSCALL:  {SYSCALL, "SYSCALL", opSysCall},

		//Stack ops
		TOALTSTACK:   {TOALTSTACK, "TOALTSTACK", opToAltStack},
		FROMALTSTACK: {FROMALTSTACK, "FROMALTSTACK", opFromAltStack},
		XDROP:        {XDROP, "XDROP", opXDrop},
		XSWAP:        {XSWAP, "XSWAPP", opXSwap},
		XTUCK:        {XTUCK, "XTUCK", opXTuck},
		DEPTH:        {DEPTH, "DEPTH", opDepth},
		DROP:         {DROP, "DROP", opDrop},
		DUP:          {DUP, "DUP", opDup},
		NIP:          {NIP, "NIP", opNip},
		OVER:         {OVER, "OVER", opOver},
		PICK:         {PICK, "PICK", opPick},
		ROLL:         {ROLL, "ROLL", opRoll},
		ROT:          {ROT, "ROT", opRot},
		SWAP:         {SWAP, "SWAP", opSwap},
		TUCK:         {TUCK, "TUCK", opTuck},

		//Splice
		CAT:    {CAT, "CAT", opCat},
		SUBSTR: {SUBSTR, "SUBSTR", opSubStr},
		LEFT:   {LEFT, "LEFT", opLeft},
		RIGHT:  {RIGHT, "RIGHT", opRight},
		SIZE:   {SIZE, "SIZE", opSize},

		//Bitwiase logic
		INVERT: {INVERT, "INVERT", opInvert},
		AND:    {AND, "AND", opBigIntZip},
		OR:     {OR, "OR", opBigIntZip},
		XOR:    {XOR, "XOR", opBigIntZip},
		EQUAL:  {EQUAL, "EQUAL", opEqual},

		//Arithmetic
		INC:         {INC, "INC", opBigInt},
		DEC:         {DEC, "DEC", opBigInt},
		SAL:         {SAL, "SAL", opBigInt},
		SAR:         {SAR, "SAR", opBigInt},
		NEGATE:      {NEGATE, "NEGATE", opBigInt},
		ABS:         {ABS, "ABS", opBigInt},
		NOT:         {NOT, "NOT", opNot},
		NZ:          {NZ, "NZ", opNz},
		ADD:         {ADD, "ADD", opBigIntZip},
		SUB:         {SUB, "SUB", opBigIntZip},
		MUL:         {MUL, "MUL", opBigIntZip},
		DIV:         {DIV, "DIV", opBigIntZip},
		MOD:         {MOD, "MOD", opBigIntZip},
		SHL:         {SHL, "SHL", opBigIntZip},
		SHR:         {SHR, "SHR", opBigIntZip},
		BOOLAND:     {BOOLAND, "BOOLAND", opBoolZip},
		BOOLOR:      {BOOLOR, "BOOLOR", opBoolZip},
		NUMEQUAL:    {NUMEQUAL, "NUMEQUAL", opBigIntComp},
		NUMNOTEQUAL: {NUMNOTEQUAL, "NUMNOTEQUAL", opBigIntComp},
		LT:          {LT, "LT", opBigIntComp},
		GT:          {GT, "GT", opBigIntComp},
		LTE:         {LTE, "LTE", opBigIntComp},
		GTE:         {GTE, "GTE", opBigIntComp},
		MIN:         {MIN, "MIN", opBigIntZip},
		MAX:         {MAX, "MAX", opBigIntZip},
		WITHIN:      {WITHIN, "WITHIN", opWithIn},

		//Crypto
		SHA1:          {SHA1, "SHA1", opHash},
		SHA256:        {SHA256, "SHA256", opHash},
		HASH160:       {HASH160, "HASH160", opHash},
		HASH256:       {HASH256, "HASH256", opHash},
		CHECKSIG:      {CHECKSIG, "CHECKSIG", opCheckSig},
		CHECKMULTISIG: {CHECKMULTISIG, "CHECKMULTISIG", opCheckMultiSig},

		//Array
		ARRAYSIZE: {ARRAYSIZE, "ARRAYSIZE", opArraySize},
		PACK:      {PACK, "PACK", opPack},
		UNPACK:    {UNPACK, "UNPACK", opUnpack},
		PICKITEM:  {PICKITEM, "PICKITEM", opPickItem},
	}
)

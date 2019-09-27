OntCversion = '2.0.0'
from ontology.interop.System.ExecutionEngine import GetExecutingScriptHash
from ontology.interop.Ontology.Wasm import InvokeWasm


def Main(operation, args):
    if operation == "bytearray":
        address = args
        return testbytearray(address)
    elif operation == "boolean":
        istrue = args
        return testboolean(istrue)
    elif operation == "intype":
        num = args
        return testintype(num)
    elif operation == "H256":
        h256 = args
        return testh256(h256)
    elif operation == "neolist":
        return testlist(*args)
    elif operation == "testcase":
        return testcase()
    elif operation == "add":
        return add(args[0], args[1])
    elif operation == "callwasm":
        return testcallwasm(args[1])
    else:
        assert(False)


def testcallwasm(textContext):
    magicversion = b'\x00'
    typebytearray = b'\x00'
    typestring = b'\x01'
    typeaddress = b'\x02'
    typebool = b'\x03'
    typeint = b'\x04'
    typeh256 = b'\x05'
    typelist = b'\x10'

    # uint32 lsize = 2
    lsize = b'\x03\x00\x00\x00'
    magicversion = concat(magicversion, typelist)
    magicversion = concat(magicversion, lsize)

    #int128 a = 2, int 128 b = 3. base len is 15bytes
    base = b'\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00'
    magicversion = concat(magicversion, typestring)
    magicversion = concat(magicversion, '\x03\x00\x00\x00')
    magicversion = concat(magicversion, "add")

    a = concat('\x02', base)
    b = concat('\x03', base)
    magicversion = concat(magicversion, typeint)
    magicversion = concat(magicversion, a)
    magicversion = concat(magicversion, typeint)
    magicversion = concat(magicversion, b)
    
    address = textContext[0]["test_add.wasm"]
    res = InvokeWasm(address, magicversion)
    c = concat('\x05', base)
    assert(res == c)
    return c+0


def testcase():
    return '''
    [
        [{"needcontext":false,"env":{"witness":[]}, "method":"add", "param":"[int:1,int:2]", "expected":"int:3"},
        {"needcontext":true,"env":{"witness":[]}, "method":"callwasm", "param":"[int:1]", "expected":"int:5"}
        ]
    ]'''


def testbytearray(addr):
    assert(GetExecutingScriptHash() == addr)
    assert(len(addr) == 20)
    return addr


def add(a, b):
    return a + b


def testboolean(istrue):
    assert(istrue == True or istrue == False)
    return istrue == True


def testintype(num):
    res = num + 0x101
    return res


def testh256(h256):
    return h256


def testlist(*arg):
    assert(len(arg) == 3)
    arg[0] += 1
    arg[2] -= 1
    return arg

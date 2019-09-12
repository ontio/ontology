OntCversion='2.0.0'
from ontology.interop.System.ExecutionEngine import GetExecutingScriptHash

def Main(operation, args):
    print(operation)
    print(args)
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
    else:
        assert(False)


def testbytearray(addr):
    assert(GetExecutingScriptHash() == addr)
    assert(len(addr) == 20)
    return addr

def testboolean(istrue):
    assert(istrue == True or istrue == False)
    return istrue == True

def testintype(num):
    res = num + 0x101
    print(res)
    return res

def testh256(h256):
    return h256

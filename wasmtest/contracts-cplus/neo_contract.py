OntCversion = '2.0.0'
from ontology.interop.System.ExecutionEngine import GetExecutingScriptHash


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
    return res


def testh256(h256):
    return h256


def testlist(*arg):
    assert(len(arg) == 3)
    arg[0] += 1
    arg[2] -= 1
    return arg

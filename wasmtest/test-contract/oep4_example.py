OntCversion = '2.0.0'
"""
APC Token
"""
from ontology.interop.System.Storage import GetContext, Get, Put, Delete
from ontology.interop.System.Runtime import CheckWitness
from ontology.interop.System.Action import RegisterAction
from ontology.builtins import concat
from ontology.interop.Ontology.Runtime import Base58ToAddress

TransferEvent = RegisterAction("transfer", "from", "to", "amount")
ApprovalEvent = RegisterAction("approval", "owner", "spender", "amount")

ctx = GetContext()

NAME = 'TOKEN_NAME'
SYMBOL = 'SYMBOL'
DECIMALS = 6
FACTOR = 1000000
TotalSupply = 1000000000
Admin = Base58ToAddress("ARGK44mXXZfU6vcdSfFKMzjaabWxyog1qb")
ZERO_ADDRESS = bytearray(b'\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00')

BALANCE_PREFIX = bytearray(b'\x01')
APPROVE_PREFIX = bytearray(b'\x02')
TOTAL_SUPPLY_KEY = bytearray(b'\x03')


def Main(operation, args):
    """
    :param operation:
    :param args:
    :return:
    """
    if operation == 'init':
        return init()
    if operation == 'name':
        return name()
    if operation == 'symbol':
        return symbol()
    if operation == 'decimals':
        return decimals()
    if operation == 'totalSupply':
        return totalSupply()
    if operation == 'balanceOf':
        assert (len(args) == 1)
        acct = args[0]
        return balanceOf(acct)
    if operation == 'transfer':
        assert (len(args) == 3)
        from_acct = args[0]
        to_acct = args[1]
        amount = args[2]
        return transfer(from_acct, to_acct, amount)
    if operation == 'transferMulti':
        return transferMulti(args)
    if operation == 'transferFrom':
        assert (len(args) == 4)
        spender = args[0]
        from_acct = args[1]
        to_acct = args[2]
        amount = args[3]
        return transferFrom(spender, from_acct, to_acct, amount)
    if operation == 'approve':
        assert (len(args) == 3)
        owner = args[0]
        spender = args[1]
        amount = args[2]
        return approve(owner, spender, amount)
    if operation == 'allowance':
        assert (len(args) == 2)
        owner = args[0]
        spender = args[1]
        return allowance(owner, spender)
    return False


def init():
    supply = Get(GetContext(), TOTAL_SUPPLY_KEY)
    assert(supply == 0)
    total = TotalSupply * FACTOR
    Put(GetContext(), TOTAL_SUPPLY_KEY, total)
    Put(GetContext(), concat(BALANCE_PREFIX, Admin), total)
    TransferEvent(ZERO_ADDRESS, Admin, total)
    return True


def name():
    """
    :return: name of the token
    """
    return NAME


def symbol():
    """
    :return: symbol of the token
    """
    return SYMBOL


def decimals():
    """
    :return: the decimals of the token
    """
    return DECIMALS + 0


def totalSupply():
    """
    :return: the total supply of the token
    """
    return Get(ctx, TOTAL_SUPPLY_KEY) + 0


def balanceOf(account):
    """
    :param account:
    :return: the token balance of account
    """
    assert (len(account) == 20)
    return Get(ctx, concat(BALANCE_PREFIX, account)) + 0


def transfer(from_acct, to_acct, amount):
    """
    Transfer amount of tokens from from_acct to to_acct
    :param from_acct: the account from which the amount of tokens will be transferred
    :param to_acct: the account to which the amount of tokens will be transferred
    :param amount: the amount of the tokens to be transferred, >= 0
    :return: True means success, False or raising exception means failure.
    """
    assert (len(to_acct) == 20)
    assert (len(from_acct) == 20)
    assert (CheckWitness(from_acct))
    assert (amount > 0)

    fromKey = concat(BALANCE_PREFIX, from_acct)
    fromBalance = Get(ctx, fromKey)

    assert (fromBalance >= amount)

    if amount == fromBalance:
        Delete(ctx, fromKey)
    else:
        Put(ctx, fromKey, fromBalance - amount)

    toKey = concat(BALANCE_PREFIX, to_acct)
    toBalance = Get(ctx, toKey)
    Put(ctx, toKey, toBalance + amount)

    TransferEvent(from_acct, to_acct, amount)

    return True


def transferMulti(args):
    """
    :param args: the parameter is an array, containing element like [from, to, amount]
    :return: True means success, False or raising exception means failure.
    """
    for p in args:
        assert (len(p) == 3)
        assert (transfer(p[0], p[1], p[2]))

        # return False is wrong since the previous transaction will be successful

    return True


def approve(owner, spender, amount):
    """
    owner allow spender to spend amount of token from owner account
    Note here, the amount should be less than the balance of owner right now.
    :param owner:
    :param spender:
    :param amount: amount>=0
    :return: True means success, False or raising exception means failure.
    """
    assert (len(owner) == 20)
    assert (len(spender) == 20)
    assert (CheckWitness(owner))
    assert (amount >= 0)

    key = concat(concat(APPROVE_PREFIX, owner), spender)
    Put(ctx, key, amount)

    ApprovalEvent(owner, spender, amount)

    return True


def transferFrom(spender, from_acct, to_acct, amount):
    """
    spender spends amount of tokens on the behalf of from_acct, spender makes a transaction of amount of tokens
    from from_acct to to_acct
    :param spender:
    :param from_acct:
    :param to_acct:
    :param amount:
    :return:
    """
    assert (len(to_acct) == 20)
    assert (len(from_acct) == 20)
    assert (len(spender) == 20)
    assert (amount >= 0)
    assert (CheckWitness(spender))

    fromKey = concat(BALANCE_PREFIX, from_acct)
    fromBalance = Get(ctx, fromKey)

    assert (fromBalance >= amount)

    approveKey = concat(concat(APPROVE_PREFIX, from_acct), spender)
    approvedAmount = Get(ctx, approveKey)
    toKey = concat(BALANCE_PREFIX, to_acct)

    assert (approvedAmount >= amount)

    if amount == approvedAmount:
        Delete(ctx, approveKey)
        Put(ctx, fromKey, fromBalance - amount)
    else:
        Put(ctx, approveKey, approvedAmount - amount)
        Put(ctx, fromKey, fromBalance - amount)

    toBalance = Get(ctx, toKey)
    Put(ctx, toKey, toBalance + amount)

    TransferEvent(from_acct, to_acct, amount)

    return True


def allowance(owner, spender):
    """
    check how many token the spender is allowed to spend from owner account
    :param owner: token owner
    :param spender:  token spender
    :return: the allowed amount of tokens
    """
    key = concat(concat(APPROVE_PREFIX, owner), spender)
    return Get(ctx, key) + 0
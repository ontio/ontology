"""
DAD token pledge
"""
from boa.interop.Ontology.Contract import Migrate
from boa.interop.System.Storage import GetContext, Get, Put, Delete
from boa.interop.System.Runtime import CheckWitness, GetTime, Notify, Serialize, Deserialize
from boa.interop.System.ExecutionEngine import GetExecutingScriptHash, GetCallingScriptHash, GetEntryScriptHash
from boa.interop.Ontology.Native import Invoke
from boa.interop.Ontology.Runtime import GetCurrentBlockHash
from boa.builtins import ToScriptHash, concat, state


"""
https://github.com/tonyclarking/python-template/blob/master/libs/Utils.py
"""
def Revert():
    """
    Revert the transaction. The opcodes of this function is `09f7f6f5f4f3f2f1f000f0`,
    but it will be changed to `ffffffffffffffffffffff` since opcode THROW doesn't
    work, so, revert by calling unused opcode.
    """
    raise Exception(0xF1F1F2F2F3F3F4F4)


"""
https://github.com/tonyclarking/python-template/blob/master/libs/SafeCheck.py
"""
def Require(condition):
    """
	If condition is not satisfied, return false
	:param condition: required condition
	:return: True or false
	"""
    if not condition:
        Revert()
    return True

def RequireScriptHash(key):
    """
    Checks the bytearray parameter is script hash or not. Script Hash
    length should be equal to 20.
    :param key: bytearray parameter to check script hash format.
    :return: True if script hash or revert the transaction.
    """
    Require(len(key) == 20)
    return True

def RequireWitness(witness):
    """
	Checks the transaction sender is equal to the witness. If not
	satisfying, revert the transaction.
	:param witness: required transaction sender
	:return: True if transaction sender or revert the transaction.
	"""
    Require(CheckWitness(witness))
    return True
"""
SafeMath 
"""

def Add(a, b):
	"""
	Adds two pledges, throws on overflow.
	"""
	c = a + b
	Require(c >= a)
	return c

def Sub(a, b):
	"""
	Substracts two pledges, throws on overflow (i.e. if subtrahend is greater than minuend).
    :param a: operand a
    :param b: operand b
    :return: a - b if a - b > 0 or revert the transaction.
	"""
	Require(a>=b)
	return a-b

def ASub(a, b):
    if a > b:
        return a - b
    if a < b:
        return b - a
    else:
        return 0

def Mul(a, b):
	"""
	Multiplies two pledges, throws on overflow.
    :param a: operand a
    :param b: operand b
    :return: a - b if a - b > 0 or revert the transaction.
	"""
	if a == 0:
		return 0
	c = a * b
	Require(c / a == b)
	return c

def Div(a, b):
	"""
	Integer division of two pledges, truncating the quotient.
	"""
	Require(b > 0)
	c = a / b
	return c

def Pwr(a, b):
    """
    a to the power of b
    :param a the base
    :param b the power value
    :return a^b
    """
    c = 0
    if a == 0:
        c = 0
    elif b == 0:
        c = 1
    else:
        i = 0
        c = 1
        while i < b:
            c = Mul(c, a)
            i = i + 1
    return c

def Sqrt(a):
    """
    Return sqrt of a
    :param a:
    :return: sqrt(a)
    """
    c = Div(Add(a, 1), 2)
    b = a
    while(c < b):
        b = c
        c = Div(Add(Div(a, c), c), 2)
    return c


######################### Global pledge info ########################
ROUND_PREFIX = "G01"
# GAS_VAULT_KEY -- store the fee for calculating the cost
GAS_VAULT_KEY = "G02"

TOTAL_DAD_KEY = "G03"
TOTAL_ONG_KEY = "G04"
CURRET_ROUND_NUM_KEY = "G05"

# PROFIT_PER_DAD_KEY -- store the profit per DAD (when it is bought)
PROFIT_PER_DAD_KEY = "G06"

# TOTAL_DIVIDEND_OF_PREFIX + account -- store the total accumulated dividend of account
# when user withdraws, the total dividend will go to ZERO
TOTAL_DIVIDEND_OF_PREFIX = "G07"
REFERRAL_BALANCE_OF_PREFIX = "G08"
AWARD_BALANCE_OF_PREFFIX = "G09"
# DAD_BALANCE_PREFIX + account -- store the current blank DAD amount of account
DAD_BALANCE_PREFIX = "G10"

REFERRAL_PREFIX = "G11"

# PROFIT_PER_DAD_FROM_PREFIX + account -- store the filled DAD amount in round i
PROFIT_PER_DAD_FROM_PREFIX = "G12"


################## Round i User info ##################
# ROUND_PREFIX + CURRET_ROUND_NUM_KEY + FILLED_DAD_BALANCE_PREFIX + account -- store the filled DAD amount in round i
FILLED_DAD_BALANCE_PREFIX = "U01"

###################### Round i Public info ###########################
# ROUND_PREFIX + CURRET_ROUND_NUM_KEY + AWARD_VAULT_KEY -- store the total award for the winner in roung i
AWARD_VAULT_KEY = "R1"

# ROUND_PREFIX + CURRET_ROUND_NUM_KEY + ROUND_STATUS_KEY -- store the status of round i Pledge
ROUND_STATUS_KEY = "R4"

# ROUND_PREFIX + CURRET_ROUND_NUM_KEY  + WINNER_KEY -- store the win info
# key = ROUND_PREFIX + CURRET_ROUND_NUM_KEY  + WINNER_KEY

WINNER_KEY = "R5"

# ROUND_PREFIX + CURRET_ROUND_NUM_KEY + FILLED_DAD_AMOUNT -- store the asset for the next round in round i+1
FILLED_DAD_AMOUNT = "R6"

############################### other info ###################################
INIIT_KEY = "Initialized"
COMMISSION_KEY = "AdminComission"
STATUS_ON = "RUNNING"
STATUS_OFF = "END"

MagnitudeForProfitPerDAD = 100000000000000000000

InitialPrice = 1000000000
DADHolderPercentage = 50
ReferralAwardPercentage = 1
AwardPercentage = 45

PureAwardExcludeCommissionFee = 98

# the script hash of this contract
ContractAddress = GetExecutingScriptHash()
ONGAddress = bytearray(b'\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02')
######################## Skyinglyh account
# Admin = ToScriptHash('AQf4Mzu1YJrhz9f3aRkkwSm9n3qhXGSh4p')
Admin = ToScriptHash('AYqCVffRcbPkf1BVCYPJqqoiFTFmvwYKhG')


# Beijing time 
StartTime = 1542870540
RoundDurationMinutes = 3

def Main(operation, args):
    ######################## for Admin to invoke Begin ###############
    if operation == "init":
        return init()
    if operation == "startNewRound":
        return startNewRound()
    if operation == "assignDAD":
        if len(args) != 2:
            return False
        account = args[0]
        DADAmount = args[1]
        return assignDAD(account, DADAmount)
    if operation == "multiAssignDAD":
        return multiAssignDAD(args)
    if operation == "withdrawGas":
        return withdrawGas()
    if operation == "withdrawCommission":
        return withdrawCommission()
    if operation == "endCurrentRound":
        return endCurrentRound()
    if operation == "getTest":
        return getTest()
    ######################## for Admin to invoke End ###############
    ######################## for User to invoke Begin ###############
    if operation == "buyDAD":
        if len(args) != 2:
            return False
        account = args[0]
        DADAmount = args[1]
        return buyDAD(account, DADAmount)
    if operation == "reinvest":
        if len(args) != 2:
            return False
        account = args[0]
        DADAmount = args[1]
        return reinvest(account, DADAmount)
    if operation == "fillDAD":
        if len(args) != 2:
            return False
        account = args[0]
        return fillDAD(account, pledgeList)
    if operation == "withdraw":
        if len(args) != 1:
            return False
        account = args[0]
        return withdraw(account)
    if operation == "updateDividendBalance":
        if len(args) != 1:
            return False
        account = args[0]
        return updateDividendBalance(account)
    ######################## for User to invoke End ###############
    ######################### General Info to pre-execute Begin ##############
    if operation == "getTotalONGAmount":
        return getTotalONGAmount()
    if operation == "getTotalDAD":
        return getTotalDAD()
    if operation == "getGasVault":
        return getGasVault()
    if operation == "getCurrentRound":
        return getCurrentRound()
    if operation == "getCurrentRoundEndTime":
        return getCurrentRoundEndTime()
    if operation == "getCommissionAmount":
        return getCommissionAmount()
    if operation == "getDADBalance":
        if len(args) != 1:
            return False
        account = args[0]
        return getDADBalance(account)
    if operation == "getReferralBalance":
        if len(args) != 1:
            return False
        account = args[0]
        return getReferralBalance(account)
    if operation == "getDividendBalance":
        if len(args) != 1:
            return False
        account = args[0]
        return getDividendBalance(account)




####################### Methods that only Admin can invoke Start #######################
def init():
    RequireWitness(Admin)
    inited = Get(GetContext(), INIIT_KEY)
    if inited:
        Notify(["idiot admin, you have initialized the contract"])
        return False
    else:
        Put(GetContext(), INIIT_KEY, 1)
        Notify(["Initialized contract successfully"])
        # startNewRound()
    return True


def startNewRound():
    """
    Only admin can start new round
    :return:
    """

    RequireWitness(Admin)
    currentRound = getCurrentRound()
    nextRoundNum = Add(currentRound, 1)

    Put(GetContext(), concatKey(concatKey(ROUND_PREFIX, nextRoundNum), ROUND_STATUS_KEY), STATUS_ON)
    Put(GetContext(), CURRET_ROUND_NUM_KEY, nextRoundNum)
    Notify(["startRound", nextRoundNum, GetTime()])
    return True

def assignDAD(account, DADAmount):
    RequireWitness(Admin)

    updateDividendBalance(account)

    # update account's profit per DAD from value
    Put(GetContext(), concatKey(PROFIT_PER_DAD_FROM_PREFIX, account), getProfitPerDAD())

    # update account DAD balance
    balanceKey = concatKey(DAD_BALANCE_PREFIX, account)
    Put(GetContext(), balanceKey, Add(DADAmount, getDADBalance(account)))

    # update total DAD amount
    Put(GetContext(), TOTAL_DAD_KEY, Add(DADAmount, getTotalDAD()))

    Notify(["assignDAD", account, DADAmount, GetTime()])

    return True


def multiAssignDAD(args):
    RequireWitness(Admin)
    for p in args:
        Require(assignDAD(p[0], p[1]))
    return True

def withdrawGas():
    """
    Only admin can withdraw
    :return:
    """
    RequireWitness(Admin)

    Require(transferONGFromContact(Admin, getGasVault()))

    # update total ong amount
    Put(GetContext(), TOTAL_ONG_KEY, Sub(getTotalONGAmount(), getGasVault()))
    Delete(GetContext(), GAS_VAULT_KEY)
    return True

def withdrawCommission():
    RequireWitness(Admin)

    Require(transferONGFromContact(Admin, getCommissionAmount()))

    # update total ong amount
    Put(GetContext(), TOTAL_ONG_KEY, Sub(getTotalONGAmount(), getCommissionAmount()))
    Delete(GetContext(), COMMISSION_KEY)

    return True


def getTest():
    return 999999

####################### Methods that only Admin can invoke End #######################


######################## Methods for Users Start ######################################
def buyDAD(account, DADAmount):
    RequireWitness(account)

    currentRound = getCurrentRound()

    Require(getPledgeStatus(currentRound) == STATUS_ON)

    ongAmount = DADToONG(DADAmount)

    Require(transferONG(account, ContractAddress, ongAmount))

    # DADHolderPercentage = 50
    dividend1 = Div(Mul(ongAmount, DADHolderPercentage), 100)
    # update referral balance <---> Get(GetContext(), concatKey(REFERRAL_PREFIX, account))
    referral = getReferral(account)
    referralAmount = 0

    if referral:
        # ReferralAwardPercentage = 1
        referralAmount = Div(Mul(ongAmount, ReferralAwardPercentage), 100)
        Put(GetContext(), concatKey(REFERRAL_BALANCE_OF_PREFIX, referral), Add(referralAmount, getReferralBalance(referral)))
    dividend = Sub(dividend1, referralAmount)

    # update award vault, AwardPercentage = 45
    awardVaultToBeAdd = Div(Mul(ongAmount, AwardPercentage), 100)
    awardVaultKey = concatKey(concatKey(ROUND_PREFIX, currentRound), AWARD_VAULT_KEY)
    Put(GetContext(), awardVaultKey, Add(awardVaultToBeAdd, getAwardVault(currentRound)))

    # update gas vault
    gasVaultToBeAdd = Sub(Sub(ongAmount, dividend1), awardVaultToBeAdd)
    Put(GetContext(), GAS_VAULT_KEY, Add(gasVaultToBeAdd, getGasVault()))

    oldProfitPerDAD = Get(GetContext(), PROFIT_PER_DAD_KEY)
    oldTotalDADAmount = getTotalDAD()

    if oldTotalDADAmount != 0:
        # profitPerDADToBeAdd = Div(dividend, totalDADAmount)
        profitPerDADToBeAdd = Div(Mul(dividend, MagnitudeForProfitPerDAD), oldTotalDADAmount)
        # update profitPerDAD\
        Put(GetContext(), PROFIT_PER_DAD_KEY, Add(profitPerDADToBeAdd, oldProfitPerDAD))
    else:
        # if current total DAD is ZERO, the dividend will be assigned as the commission part
        Put(GetContext(), COMMISSION_KEY, Add(dividend, getCommissionAmount()))

    updateDividendBalance(account)

    # update DAD balance of account
    Put(GetContext(), concatKey(DAD_BALANCE_PREFIX, account), Add(DADAmount, getDADBalance(account)))

    # update total DAD amount
    Put(GetContext(), TOTAL_DAD_KEY, Add(DADAmount, getTotalDAD()))

    # update total ONG
    Put(GetContext(), TOTAL_ONG_KEY, Add(getTotalONGAmount(), ongAmount))
    Notify(["buyDAD", account, ongAmount, DADAmount, GetTime()])

    return True


def reinvest(account, DADAmount):
    RequireWitness(account)

    currentRound = getCurrentRound()

    Require(getPledgeStatus(currentRound) == STATUS_ON)

    ongAmount = DADToONG(DADAmount)

    # updateDividendBalance(account)
    dividendBalance = getDividendBalance(account)
    awardBalance = getAwardBalance(account)
    referralBalance = getReferralBalance(account)
    assetToBeReinvest = Add(Add(dividendBalance, awardBalance), referralBalance)

    Require(assetToBeReinvest >= ongAmount)

    dividend1 = Div(Mul(ongAmount, DADHolderPercentage), 100)
    # update referral balance
    referral = getReferral(account)
    referralAmount = 0
    if referral:
        referralAmount = Div(Mul(ongAmount, ReferralAwardPercentage), 100)
        Put(GetContext(), concatKey(REFERRAL_BALANCE_OF_PREFIX, referral), Add(referralAmount, getReferralBalance(referral)))
    dividend = Sub(dividend1, referralAmount)

    # update award vault
    awardVaultToBeAdd = Div(Mul(ongAmount, AwardPercentage), 100)
    awardVaultKey = concatKey(concatKey(ROUND_PREFIX, currentRound), AWARD_VAULT_KEY)
    Put(GetContext(), awardVaultKey, Add(awardVaultToBeAdd, getAwardVault(currentRound)))

    # update gas vault
    gasVaultToBeAdd = Sub(Sub(ongAmount, dividend1), awardVaultToBeAdd)
    Put(GetContext(), GAS_VAULT_KEY, Add(gasVaultToBeAdd, getGasVault()))

    # update profitPerDAD
    oldProfitPerDAD = Get(GetContext(), PROFIT_PER_DAD_KEY)
    oldTotalDADAmount = getTotalDAD()

    if oldTotalDADAmount != 0:
        profitPerDADToBeAdd = Div(Mul(dividend, MagnitudeForProfitPerDAD), oldTotalDADAmount)
        # update profitPerDAD
        Put(GetContext(), PROFIT_PER_DAD_KEY, Add(profitPerDADToBeAdd, oldProfitPerDAD))
    else:
        # if current total DAD is ZERO, the dividend will be assigned as the commission part
        Put(GetContext(), COMMISSION_KEY, Add(dividend, getCommissionAmount()))

    updateDividendBalance(account)

    # update DAD balance of account
    Put(GetContext(), concatKey(DAD_BALANCE_PREFIX, account), Add(DADAmount, getDADBalance(account)))

    # update total DAD amount
    Put(GetContext(), TOTAL_DAD_KEY, Add(DADAmount, getTotalDAD()))

    # update the account balances of dividend, award, referral
    ongAmountNeedToBeDeduct = ongAmount
    if ongAmountNeedToBeDeduct >= dividendBalance:
        ongAmountNeedToBeDeduct = Sub(ongAmountNeedToBeDeduct, dividendBalance)
        Delete(GetContext(), concatKey(TOTAL_DIVIDEND_OF_PREFIX, account))
    else:
        Put(GetContext(), concatKey(TOTAL_DIVIDEND_OF_PREFIX, account), Sub(dividendBalance, ongAmountNeedToBeDeduct))
        ongAmountNeedToBeDeduct = 0
    if ongAmountNeedToBeDeduct != 0:
        if ongAmountNeedToBeDeduct >= referralBalance:
            ongAmountNeedToBeDeduct = Sub(ongAmountNeedToBeDeduct, referralBalance)
            Delete(GetContext(), concatKey(REFERRAL_BALANCE_OF_PREFIX, account))
        else:
            Put(GetContext(), concatKey(REFERRAL_BALANCE_OF_PREFIX, account), Sub(referralBalance, ongAmountNeedToBeDeduct))
            ongAmountNeedToBeDeduct = 0
    if ongAmountNeedToBeDeduct != 0:
        if ongAmountNeedToBeDeduct > awardBalance:
            raise Exception("Reinvest failed!")
        else:
            Put(GetContext(), concatKey(AWARD_BALANCE_OF_PREFFIX, account), Sub(awardBalance, ongAmountNeedToBeDeduct))

    # PurchaseEvent(account, ongAmount, DADAmount)
    Notify(["reBuyDAD", account, ongAmount, DADAmount, GetTime()])

    return True


def fillDAD(account, pledgeList):
    """
    :param account:
    :param pledgeList: can be a list of pledges
    :return:
    """
    RequireWitness(account)

    currentRound = getCurrentRound()

    Require(getPledgeStatus(currentRound) == STATUS_ON)

    # to prevent hack from other contract
    callerHash = GetCallingScriptHash()
    entryHash = GetEntryScriptHash()
    Require(callerHash == entryHash)

    pledgeLen = len(pledgeList)

    Require(pledgeLen >= 1)

    currentDADBalance = getDADBalance(account)

    # make sure his balance is greater or equal to pledgeList length
    Require(currentDADBalance >= pledgeLen)

    pledgeListKey = concatKey(concatKey(ROUND_PREFIX, currentRound), FILLED_PLEDGE_LIST_KEY)
    pledgeListInfo = Get(GetContext(), pledgeListKey)
    pledgeList = []
    if pledgeListInfo:
        pledgeList = Deserialize(pledgeListInfo)

    for pledge in pledgeList:

        # Require is need to raise exception
        Require(pledge < 100)
        Require(pledge >= 0)

        pledgePlayersListKey = concatKey(concatKey(ROUND_PREFIX, currentRound), concatKey(FILLED_pledge_KEY, pledge))
        pledgePlayersListInfo = Get(GetContext(), pledgePlayersListKey)

        pledgePlayersList = []
        if pledgePlayersListInfo:
            pledgePlayersList = Deserialize(pledgePlayersListInfo)

            # make sure account has NOT filled the pledge before in this round
            for player in pledgePlayersList:
                Require(player != account)
        else:
            pledgeList.append(pledge)

        # add account to the players list that filled the pledge in this round
        pledgePlayersList.append(account)

        # Store the pledgePlayers List
        pledgePlayersListInfo = Serialize(pledgePlayersList)
        Put(GetContext(), pledgePlayersListKey, pledgePlayersListInfo)
    # Store the pledgeList
    pledgeListInfo = Serialize(pledgeList)
    Put(GetContext(), pledgeListKey, pledgeListInfo)

    # update dividend
    updateDividendBalance(account)

    # update the DAD balance of account  -- destroy the filled DADs
    Put(GetContext(), concatKey(DAD_BALANCE_PREFIX, account), Sub(currentDADBalance, pledgeLen))

    # update total DAD amount
    Put(GetContext(), TOTAL_DAD_KEY, Sub(getTotalDAD(), pledgeLen))

    # update the filled DAD balance of account in current round
    key1 = concatKey(ROUND_PREFIX, currentRound)
    key2 = concatKey(FILLED_DAD_BALANCE_PREFIX, account)
    key = concatKey(key1, key2)
    Put(GetContext(), key, Add(pledgeLen, getFilledDADBalance(account, currentRound)))

    # update the filled DAD amount in current round
    key = concatKey(concatKey(ROUND_PREFIX, currentRound), FILLED_DAD_AMOUNT)
    Put(GetContext(), key, Add(pledgeLen, getFilledDADAmount(currentRound)))

    Notify(["fillDAD", account, pledgeList, GetTime(), currentRound])

    return True


def withdraw(account):
    """
    account will withdraw his dividend and award to his own account
    :param account:
    :return:
    """
    RequireWitness(account)

    updateDividendBalance(account)
    dividendBalance = getDividendBalance(account)
    awardBalance = getAwardBalance(account)
    referralBalance = getReferralBalance(account)
    assetToBeWithdrawn = Add(Add(dividendBalance, awardBalance), referralBalance)

    Require(assetToBeWithdrawn > 0)
    Require(transferONGFromContact(account, assetToBeWithdrawn))

    Delete(GetContext(), concatKey(TOTAL_DIVIDEND_OF_PREFIX, account))
    Delete(GetContext(), concatKey(AWARD_BALANCE_OF_PREFFIX, account))
    Delete(GetContext(), concatKey(REFERRAL_BALANCE_OF_PREFIX, account))

    # update total ong amount
    Put(GetContext(), TOTAL_ONG_KEY, Sub(getTotalONGAmount(), assetToBeWithdrawn))

    Notify(["withdraw", ContractAddress, account, assetToBeWithdrawn, GetTime()])

    return True

def updateDividendBalance(account):
    """
    reset PROFIT_PER_DAD_FROM_PREFIX of account and update account's dividend till now
    :param account:
    :return:
    """
    # RequireWitness(account)
    key = concatKey(PROFIT_PER_DAD_FROM_PREFIX, account)
    profitPerDADFrom = Get(GetContext(), key)
    profitPerDADNow = getProfitPerDAD()
    profitPerDAD = Sub(profitPerDADNow, profitPerDADFrom)

    if profitPerDAD != 0:
        Put(GetContext(), concatKey(TOTAL_DIVIDEND_OF_PREFIX, account), getDividendBalance(account))
        Put(GetContext(), concatKey(PROFIT_PER_DAD_FROM_PREFIX, account), profitPerDADNow)

    return True
######################## Methods for Users End ######################################


################## Global Info Start #######################
def getTotalONGAmount():
    return Get(GetContext(), TOTAL_ONG_KEY)
def getTotalDAD():
    return Get(GetContext(), TOTAL_DAD_KEY)

def getGasVault():
    return Get(GetContext(), GAS_VAULT_KEY)

def getCurrentRound():
    return Get(GetContext(), CURRET_ROUND_NUM_KEY)

def getCurrentRoundEndTime():
    currentRound = getCurrentRound()
    currentRoundEndTime = Add(StartTime, Mul(currentRound, Mul(RoundDurationMinutes, 60)))
    return currentRoundEndTime
def getCommissionAmount():
    return Get(GetContext(), COMMISSION_KEY)
################## Global Info End #######################


####################### User Info Start #####################
def getDADBalance(account):
    return Get(GetContext(), concatKey(DAD_BALANCE_PREFIX, account))

def getDividendBalance(account):
    key = concatKey(PROFIT_PER_DAD_FROM_PREFIX, account)
    profitPerDADFrom = Get(GetContext(), key)
    profitPerDADNow = Get(GetContext(), PROFIT_PER_DAD_KEY)
    profitPerDAD = profitPerDADNow - profitPerDADFrom
    profit = 0
    if profitPerDAD != 0:
        profit = Div(Mul(profitPerDAD, getDADBalance(account)), MagnitudeForProfitPerDAD)
    return Add(Get(GetContext(), concatKey(TOTAL_DIVIDEND_OF_PREFIX, account)), profit)

def getDividendsBalance(account):
    return [getReferralBalance(account), getDividendBalance(account), getAwardBalance(account)]


def transferONG(fromAcct, toAcct, amount):
    """
    transfer ONG
    :param fromacct:
    :param toacct:
    :param amount:
    :return:
    """
    RequireWitness(fromAcct)
    param = state(fromAcct, toAcct, amount)
    res = Invoke(0, ONGAddress, 'transfer', [param])
    if res and res == b'\x01':
        return True
    else:
        return False

def concatKey(str1,str2):
    """
    connect str1 and str2 together as a key
    :param str1: string1
    :param str2:  string2
    :return: string1_string2
    """
    return concat(concat(str1, '_'), str2)

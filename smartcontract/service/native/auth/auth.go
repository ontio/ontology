/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package auth

import (
	"bytes"
	"fmt"
	"time"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

var (
	future = time.Date(2100, 1, 1, 12, 0, 0, 0, time.UTC)
)

func Init() {
	native.Contracts[utils.AuthContractAddress] = RegisterAuthContract
}

/*
 * contract admin management
 */
func initContractAdmin(native *native.NativeService, contractAddr common.Address, ontID []byte) (bool, error) {
	admin, err := getContractAdmin(native, contractAddr)
	if err != nil {
		return false, err
	}
	if admin != nil {
		//admin is already set, just return
		log.Debugf("admin of contract %s is already set", contractAddr.ToHexString())
		return false, nil
	}
	err = putContractAdmin(native, contractAddr, ontID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func InitContractAdmin(native *native.NativeService) ([]byte, error) {
	param := new(InitContractAdminParam)
	source := common.NewZeroCopySource(native.Input)
	if err := param.Deserialization(source); err != nil {
		return nil, fmt.Errorf("[initContractAdmin] deserialize param failed: %v", err)
	}
	cxt := native.ContextRef.CallingContext()
	if cxt == nil {
		return nil, fmt.Errorf("[initContractAdmin] no calling context")
	}
	invokeAddr := cxt.ContractAddress

	if !account.VerifyID(string(param.AdminOntID)) {
		return nil, fmt.Errorf("[initContractAdmin] invalid param: adminOntID is %x", param.AdminOntID)
	}
	ret, err := initContractAdmin(native, invokeAddr, param.AdminOntID)
	if err != nil {
		return nil, fmt.Errorf("[initContractAdmin] init failed: %v", err)
	}
	if !ret {
		return utils.BYTE_FALSE, nil
	}

	msg := []interface{}{"initContractAdmin", invokeAddr.ToHexString(), string(param.AdminOntID)}
	pushEvent(native, msg)
	return utils.BYTE_TRUE, nil
}

func transfer(native *native.NativeService, contractAddr common.Address, newAdminOntID []byte, keyNo uint64) (bool, error) {
	admin, err := getContractAdmin(native, contractAddr)
	if err != nil {
		return false, fmt.Errorf("getContractAdmin failed: %v", err)
	}
	if admin == nil {
		log.Debugf("admin of contract %s is not set", contractAddr.ToHexString())
		return false, nil
	}

	ret, err := verifySig(native, admin, keyNo)
	if err != nil {
		return false, fmt.Errorf("verifySig failed: %v", err)
	}
	if !ret {
		log.Debugf("verify Admin's signature failed: admin=%s, keyNo=%d", string(admin), keyNo)
		return false, nil
	}

	adminKey := concatContractAdminKey(native, contractAddr)
	utils.PutBytes(native, adminKey, newAdminOntID)
	return true, nil
}

func Transfer(native *native.NativeService) ([]byte, error) {
	//deserialize param
	param := new(TransferParam)
	err := param.Deserialization(common.NewZeroCopySource(native.Input))
	if err != nil {
		return nil, fmt.Errorf("[transfer] deserialize param failed: %v", err)
	}

	if !account.VerifyID(string(param.NewAdminOntID)) {
		return nil, fmt.Errorf("[transfer] invalid param: newAdminOntID is %x", param.NewAdminOntID)
	}
	//prepare event msg
	contract := param.ContractAddr.ToHexString()
	failState := []interface{}{"transfer", contract, false}
	sucState := []interface{}{"transfer", contract, true}

	//call transfer func
	ret, err := transfer(native, param.ContractAddr, param.NewAdminOntID, param.KeyNo)
	if err != nil {
		return nil, fmt.Errorf("[transfer] transfer failed: %v", err)
	}
	if ret {
		pushEvent(native, sucState)
		return utils.BYTE_TRUE, nil
	} else {
		pushEvent(native, failState)
		return utils.BYTE_FALSE, nil
	}
}

func AssignFuncsToRole(native *native.NativeService) ([]byte, error) {
	//deserialize input param
	param := new(FuncsToRoleParam)
	source := common.NewZeroCopySource(native.Input)
	if err := param.Deserialization(source); err != nil {
		return nil, fmt.Errorf("[assignFuncsToRole] deserialize param failed: %v", err)
	}

	//prepare event msg
	contract := param.ContractAddr.ToHexString()
	failState := []interface{}{"assignFuncsToRole", contract, false}
	sucState := []interface{}{"assignFuncsToRole", contract, true}

	if param.Role == nil {
		return nil, fmt.Errorf("[assignFuncsToRole] invalid param: role is nil")
	}

	//check the caller's permission
	admin, err := getContractAdmin(native, param.ContractAddr)
	if err != nil {
		return nil, fmt.Errorf("[assignFuncsToRole] getContractAdmin failed: %v", err)
	}
	if admin == nil { //admin has not been set
		return nil, fmt.Errorf("[assignFuncsToRole] admin of contract %s has not been set",
			param.ContractAddr.ToHexString())
	}
	if bytes.Compare(admin, param.AdminOntID) != 0 {
		log.Debugf("[assignFuncsToRole] invalid param: adminOntID doesn't match %s != %s",
			string(param.AdminOntID), string(admin))
		pushEvent(native, failState)
		return utils.BYTE_FALSE, nil
	}
	ret, err := verifySig(native, param.AdminOntID, param.KeyNo)
	if err != nil {
		return nil, fmt.Errorf("[assignFuncsToRole] verify admin's signature failed: %v", err)
	}
	if !ret {
		log.Debugf("[assignFuncsToRole] verifySig return false: adminOntID=%s, keyNo=%d",
			string(admin), param.KeyNo)
		pushEvent(native, failState)
		return utils.BYTE_FALSE, nil
	}

	funcs, err := getRoleFunc(native, param.ContractAddr, param.Role)
	if err != nil {
		return nil, fmt.Errorf("[assignFuncsToRole] getRoleFunc failed: %v", err)
	}
	if funcs == nil {
		funcs = new(roleFuncs)
	}

	funcs.AppendFuncs(param.FuncNames)

	err = putRoleFunc(native, param.ContractAddr, param.Role, funcs)
	if err != nil {
		return nil, fmt.Errorf("[assignFuncsToRole] putRoleFunc failed: %v", err)
	}

	pushEvent(native, sucState)
	return utils.BYTE_TRUE, nil
}

func assignToRole(native *native.NativeService, param *OntIDsToRoleParam) (bool, error) {
	//check admin's permission
	admin, err := getContractAdmin(native, param.ContractAddr)
	if err != nil {
		return false, fmt.Errorf("getContractAdmin failed: %v", err)
	}
	if admin == nil {
		return false, fmt.Errorf("admin of contract %s is not set", param.ContractAddr.ToHexString())
	}
	if bytes.Compare(admin, param.AdminOntID) != 0 {
		log.Debugf("param's adminOntID doesn't match: %s != %s", string(param.AdminOntID),
			string(admin))
		return false, nil
	}
	valid, err := verifySig(native, param.AdminOntID, param.KeyNo)
	if err != nil {
		return false, fmt.Errorf("verify admin's signature failed: %v", err)
	}
	if !valid {
		log.Debugf("[assignOntIDsToRole] verifySig return false: adminOntID=%s, keyNo=%d",
			string(admin), param.KeyNo)
		return false, nil
	}

	//init a permanent auth token
	token := new(AuthToken)
	token.expireTime = uint32(future.Unix())
	token.level = 2
	token.role = param.Role

	for _, p := range param.Persons {
		if p == nil {
			continue
		}
		tokens, err := getOntIDToken(native, param.ContractAddr, p)
		if err != nil {
			return false, fmt.Errorf("getOntIDToken failed: %v", err)
		}
		if tokens == nil {
			tokens = new(roleTokens)
			tokens.tokens = make([]*AuthToken, 1)
			tokens.tokens[0] = token
		} else {
			ret, err := hasRole(native, param.ContractAddr, p, param.Role)
			if err != nil {
				return false, fmt.Errorf("check if %s has role %s failed: %v", string(p),
					string(param.Role), err)
			}
			if !ret {
				tokens.tokens = append(tokens.tokens, token)
			} else {
				continue
			}
		}
		err = putOntIDToken(native, param.ContractAddr, p, tokens)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func AssignOntIDsToRole(native *native.NativeService) ([]byte, error) {
	//deserialize param
	param := new(OntIDsToRoleParam)
	source := common.NewZeroCopySource(native.Input)
	if err := param.Deserialization(source); err != nil {
		return nil, fmt.Errorf("[assignOntIDsToRole] deserialize param failed: %v", err)
	}

	if param.Role == nil {
		return nil, fmt.Errorf("[assignOntIDsToRole] invalid param: role is nil")
	}
	for i, ontID := range param.Persons {
		if !account.VerifyID(string(ontID)) {
			return nil, fmt.Errorf("[assignOntIDsToRole] invalid param: param.Persons[%d]=%s",
				i, string(ontID))
		}
	}

	ret, err := assignToRole(native, param)
	if err != nil {
		return nil, fmt.Errorf("[assignOntIDsToRole] failed: %v", err)
	}

	contract := param.ContractAddr.ToHexString()
	failState := []interface{}{"assignOntIDsToRole", contract, false}
	sucState := []interface{}{"assignOntIDsToRole", contract, true}
	if ret {
		pushEvent(native, sucState)
		return utils.BYTE_TRUE, nil
	} else {
		pushEvent(native, failState)
		return utils.BYTE_FALSE, nil
	}
}

func getAuthToken(native *native.NativeService, contractAddr common.Address, ontID, role []byte) (*AuthToken, error) {
	tokens, err := getOntIDToken(native, contractAddr, ontID)
	if err != nil {
		return nil, fmt.Errorf("get token failed, caused by %v", err)
	}
	if tokens != nil {
		for _, token := range tokens.tokens {
			if bytes.Compare(token.role, role) == 0 { //permanent token
				return token, nil
			}
		}
	}
	status, err := getDelegateStatus(native, contractAddr, ontID)
	if err != nil {
		return nil, fmt.Errorf("get delegate status failed, caused by %v", err)
	}
	if status != nil {
		for _, s := range status.status {
			if bytes.Compare(s.role, role) == 0 && native.Time < s.expireTime { //temporary token
				token := new(AuthToken)
				token.role = s.role
				token.level = s.level
				token.expireTime = s.expireTime
				return token, nil
			}
		}
	}
	return nil, nil
}

func hasRole(native *native.NativeService, contractAddr common.Address, ontID, role []byte) (bool, error) {
	token, err := getAuthToken(native, contractAddr, ontID, role)
	if err != nil {
		return false, err
	}
	if token == nil {
		return false, nil
	}
	return true, nil
}

func getLevel(native *native.NativeService, contractAddr common.Address, ontID, role []byte) (uint8, error) {
	token, err := getAuthToken(native, contractAddr, ontID, role)
	if err != nil {
		return 0, err
	}
	if token == nil {
		return 0, nil
	}
	return token.level, nil
}

/*
 * if 'from' has the authority and 'to' has not been authorized 'role',
 * then make changes to storage as follows:
 */
func delegate(native *native.NativeService, contractAddr common.Address, from []byte, to []byte,
	role []byte, period uint32, level uint8, keyNo uint64) (bool, error) {
	var fromHasRole, toHasRole bool
	var fromLevel uint8
	var fromExpireTime uint32

	//check input param
	expireTime := native.Time
	if period+expireTime < period {
		//invalid period param, causing overflow
		return false, fmt.Errorf("[delegate] invalid param: overflow, period=%d", period)
	}
	expireTime = expireTime + period

	//check from's permission
	ret, err := verifySig(native, from, keyNo)
	if err != nil {
		return false, fmt.Errorf("verify %s's signature failed: %v", string(from), err)
	}
	if !ret {
		log.Debugf("verifySig return false: from=%s, keyNo=%d", string(from), keyNo)
		return false, nil
	}

	if !account.VerifyID(string(to)) {
		return false, fmt.Errorf("can not pass OntID validity test: to=%s", string(to))
	}

	//get from's auth token
	fromToken, err := getAuthToken(native, contractAddr, from, role)
	if err != nil {
		return false, fmt.Errorf("getAuthToken of %s failed: %v", string(from), err)
	}
	if fromToken == nil {
		fromHasRole = false
		fromLevel = 0
	} else {
		fromHasRole = true
		fromLevel = fromToken.level
		fromExpireTime = fromToken.expireTime
	}

	//get to's auth token
	toToken, err := getAuthToken(native, contractAddr, to, role)
	if err != nil {
		return false, fmt.Errorf("getAuthToken of %s failed: %v", string(to), err)
	}
	if toToken == nil {
		toHasRole = false
	} else {
		toHasRole = true
	}
	if !fromHasRole || toHasRole {
		log.Debugf("%s doesn't have role %s or %s already has role %s", string(from), string(role),
			string(to), string(role))
		return false, nil
	}

	//check if 'from' has the permission to delegate
	if fromLevel == 2 {
		if level < fromLevel && level > 0 && expireTime < fromExpireTime {
			status, err := getDelegateStatus(native, contractAddr, to)
			if err != nil {
				return false, fmt.Errorf("getDelegateStatus failed: %v", err)
			}
			if status == nil {
				status = new(Status)
			}
			j := -1
			for i, s := range status.status {
				if bytes.Compare(s.role, role) == 0 {
					j = i
					break
				}
			}
			if j < 0 {
				newStatus := &DelegateStatus{
					root: from,
				}
				newStatus.expireTime = expireTime
				newStatus.role = role
				newStatus.level = uint8(level)
				status.status = append(status.status, newStatus)
			} else {
				status.status[j].level = uint8(level)
				status.status[j].expireTime = expireTime
				status.status[j].root = from
			}
			err = putDelegateStatus(native, contractAddr, to, status)
			if err != nil {
				return false, fmt.Errorf("putDelegateStatus failed: %v", err)
			}
			return true, nil
		}
	}
	//TODO: for fromLevel > 2 case
	return false, nil
}

func Delegate(native *native.NativeService) ([]byte, error) {
	//deserialize param
	param := &DelegateParam{}
	source := common.NewZeroCopySource(native.Input)
	err := param.Deserialization(source)
	if err != nil {
		return nil, fmt.Errorf("[delegate] deserialize param failed: %v", err)
	}
	if param.Period > 1<<32 || param.Level > 1<<8 {
		return nil, fmt.Errorf("[delegate] period or level is too large")
	}

	//prepare event msg
	contract := param.ContractAddr.ToHexString()
	failState := []interface{}{"delegate", contract, param.From, param.To, false}
	sucState := []interface{}{"delegate", contract, param.From, param.To, true}

	//call the delegate func
	ret, err := delegate(native, param.ContractAddr, param.From, param.To, param.Role,
		uint32(param.Period), uint8(param.Level), param.KeyNo)
	if err != nil {
		return nil, fmt.Errorf("[delegate] failed: %v", err)
	}
	if ret {
		pushEvent(native, sucState)
		return utils.BYTE_TRUE, nil
	} else {
		pushEvent(native, failState)
		return utils.BYTE_FALSE, nil
	}
}

func withdraw(native *native.NativeService, contractAddr common.Address, initiator []byte, delegate []byte,
	role []byte, keyNo uint64) (bool, error) {
	//check from's permission
	ret, err := verifySig(native, initiator, keyNo)
	if err != nil {
		return false, fmt.Errorf("verifySig failed: %v", err)
	}
	if !ret {
		log.Debugf("verifySig return false: initiator=%s, keyNo=%d", string(initiator), keyNo)
		return false, nil
	}

	//code below only works in the case that initiator's level is 2
	//TODO: remove the above limitation
	initToken, err := getAuthToken(native, contractAddr, initiator, role)
	if err != nil {
		return false, fmt.Errorf("getAuthToken failed: %v", err)
	}
	if initToken == nil {
		//initiator does not have the right to withdraw
		log.Debugf("[withdraw] initiator %s does not have the right to withdraw", string(initiator))
		return false, nil
	}
	status, err := getDelegateStatus(native, contractAddr, delegate)
	if err != nil {
		return false, fmt.Errorf("getDelegateStatus failed: %v", err)
	}
	if status == nil {
		return false, nil
	}
	for i, s := range status.status {
		if bytes.Compare(s.role, role) == 0 &&
			bytes.Compare(s.root, initiator) == 0 {
			newStatus := new(Status)
			newStatus.status = append(status.status[:i], status.status[i+1:]...)
			err = putDelegateStatus(native, contractAddr, delegate, newStatus)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func Withdraw(native *native.NativeService) ([]byte, error) {
	//deserialize param
	param := &WithdrawParam{}
	source := common.NewZeroCopySource(native.Input)
	err := param.Deserialization(source)
	if err != nil {
		return nil, fmt.Errorf("[withdraw] deserialize param failed: %v", err)
	}

	//prepare event msg
	contract := param.ContractAddr.ToHexString()
	failState := []interface{}{"withdraw", contract, param.Initiator, param.Delegate, false}
	sucState := []interface{}{"withdraw", contract, param.Initiator, param.Delegate, true}

	//call the withdraw func
	ret, err := withdraw(native, param.ContractAddr, param.Initiator, param.Delegate, param.Role, param.KeyNo)
	if err != nil {
		return nil, fmt.Errorf("[withdraw] withdraw failed: %v", err)
	}
	if ret {
		pushEvent(native, sucState)
		return utils.BYTE_TRUE, nil
	} else {
		pushEvent(native, failState)
		return utils.BYTE_FALSE, nil
	}
}

func verifyToken(native *native.NativeService, contractAddr common.Address, caller []byte, fn string, keyNo uint64) (bool, error) {
	//check caller's identity
	ret, err := verifySig(native, caller, keyNo)
	if err != nil {
		return false, fmt.Errorf("verifySig failed: %v", err)
	}
	if !ret {
		log.Debugf("verifySig return false: caller=%s, keyNo=%d", string(caller), keyNo)
		return false, nil
	}

	//check if caller has the permanent auth token
	tokens, err := getOntIDToken(native, contractAddr, caller)
	if err != nil {
		return false, fmt.Errorf("getOntIDToken failed: %v", err)
	}
	if tokens != nil {
		for _, token := range tokens.tokens {
			funcs, err := getRoleFunc(native, contractAddr, token.role)
			if err != nil {
				return false, fmt.Errorf("getRoleFunc failed: %v", err)
			}
			if funcs == nil || token.expireTime < native.Time {
				continue
			}
			if funcs.ContainsFunc(fn) {
				return true, nil
			}
		}
	}

	status, err := getDelegateStatus(native, contractAddr, caller)
	if err != nil {
		return false, fmt.Errorf("getDelegateStatus failed: %v", err)
	}
	if status != nil {
		for _, s := range status.status {
			funcs, err := getRoleFunc(native, contractAddr, s.role)
			if err != nil {
				return false, fmt.Errorf("getRoleFunc failed: %v", err)
			}
			if funcs == nil || s.expireTime < native.Time {
				continue
			}
			if funcs.ContainsFunc(fn) {
				return true, nil
			}
		}
	}
	return false, nil
}

func VerifyToken(native *native.NativeService) ([]byte, error) {
	//deserialize param
	param := &VerifyTokenParam{}
	source := common.NewZeroCopySource(native.Input)
	err := param.Deserialization(source)
	if err != nil {
		return nil, fmt.Errorf("[verifyToken] deserialize param failed: %v", err)
	}

	contract := param.ContractAddr.ToHexString()
	failState := []interface{}{"verifyToken", contract, param.Caller, param.Fn, false}
	sucState := []interface{}{"verifyToken", contract, param.Caller, param.Fn, true}

	ret, err := verifyToken(native, param.ContractAddr, param.Caller, param.Fn, param.KeyNo)
	if err != nil {
		return nil, fmt.Errorf("[verifyToken] verifyToken failed: %v", err)
	}
	if ret {
		pushEvent(native, sucState)
		return utils.BYTE_TRUE, nil
	}
	pushEvent(native, failState)
	return utils.BYTE_FALSE, nil
}

func verifySig(native *native.NativeService, ontID []byte, keyNo uint64) (bool, error) {
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes(ontID)
	utils.EncodeVarUint(sink, keyNo)
	args := sink.Bytes()
	ret, err := native.NativeCall(utils.OntIDContractAddress, "verifySignature", args)
	if err != nil {
		return false, err
	}
	if bytes.Compare(ret, utils.BYTE_TRUE) == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func RegisterAuthContract(native *native.NativeService) {
	native.Register("initContractAdmin", InitContractAdmin)
	native.Register("assignFuncsToRole", AssignFuncsToRole)
	native.Register("delegate", Delegate)
	native.Register("withdraw", Withdraw)
	native.Register("assignOntIDsToRole", AssignOntIDsToRole)
	native.Register("verifyToken", VerifyToken)
	native.Register("transfer", Transfer)
}

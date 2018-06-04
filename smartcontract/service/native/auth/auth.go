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
	"strings"
	"time"

	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/event"
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
func initContractAdmin(native *native.NativeService, contractAddr, ontID []byte) (bool, error) {
	admin, err := getContractAdmin(native, contractAddr)
	if err != nil {
		return false, err
	}
	if admin != nil {
		//admin is already set, just return
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
	rd := bytes.NewReader(native.Input)
	if err := param.Deserialize(rd); err != nil {
		return nil, err
	}
	cxt := native.ContextRef.CallingContext()
	if cxt == nil {
		return nil, errors.NewErr("no calling context")
	}
	invokeAddr := cxt.ContractAddress

	ret, err := initContractAdmin(native, invokeAddr[:], param.AdminOntID)
	if err != nil {
		return nil, err
	}
	if !ret {
		return utils.BYTE_FALSE, nil
	}
	pushEvent(native, []interface{}{"initContractAdmin", param.AdminOntID})
	return utils.BYTE_TRUE, nil
}

func transfer(native *native.NativeService, contractAddr, newAdminOntID []byte, keyNo uint64) (bool, error) {
	admin, err := getContractAdmin(native, contractAddr)
	if err != nil {
		return false, err
	}
	if admin == nil {
		return false, nil
	}

	ret, err := verifySig(native, admin, keyNo)
	if err != nil {
		return false, err
	}
	if !ret {
		return false, nil
	}

	adminKey, err := concatContractAdminKey(native, contractAddr)
	if err != nil {
		return false, err
	}
	utils.PutBytes(native, adminKey, newAdminOntID)
	return true, nil
}

func Transfer(native *native.NativeService) ([]byte, error) {
	param := new(TransferParam)
	rd := bytes.NewReader(native.Input)
	err := param.Deserialize(rd)
	if err != nil {
		return nil, err
	}
	ret, err := transfer(native, param.ContractAddr, param.NewAdminOntID, param.KeyNo)
	if ret {
		event := new(event.NotifyEventInfo)
		event.ContractAddress = native.ContextRef.CurrentContext().ContractAddress
		event.States = []interface{}{"transfer", param.NewAdminOntID}
		native.Notifications = append(native.Notifications, event)
		return utils.BYTE_TRUE, nil
	} else {
		return utils.BYTE_FALSE, nil
	}
}

func AssignFuncsToRole(native *native.NativeService) ([]byte, error) {
	//deserialize input param
	param := new(FuncsToRoleParam)
	rd := bytes.NewReader(native.Input)
	if err := param.Deserialize(rd); err != nil {
		return nil, fmt.Errorf("deserialize param failed, caused by %v", err)
	}
	if param.Role == nil {
		invokeEvent(native, "assignFuncsToRole", false)
		return utils.BYTE_FALSE, nil
	}

	//check the caller's permission
	admin, err := getContractAdmin(native, param.ContractAddr)
	if err != nil {
		return nil, fmt.Errorf("get contract admin failed, caused by %v", err)
	}
	if admin == nil { //admin has not been set
		invokeEvent(native, "assignFuncsToRole", false)
		return utils.BYTE_FALSE, nil
	}
	if bytes.Compare(admin, param.AdminOntID) != 0 {
		invokeEvent(native, "assignFuncsToRole", false)
		return utils.BYTE_FALSE, nil
	}
	ret, err := verifySig(native, param.AdminOntID, param.KeyNo)
	if err != nil {
		return nil, fmt.Errorf("verify admin's signature failed, caused by %v", err)
	}
	if !ret {
		invokeEvent(native, "assignFuncsToRole", false)
		return utils.BYTE_FALSE, nil
	}

	funcs, err := getRoleFunc(native, param.ContractAddr, param.Role)
	if funcs != nil {
		funcNames := append(funcs.funcNames, param.FuncNames...)
		funcs.funcNames = stringSliceUniq(funcNames)
	} else {
		funcs = new(roleFuncs)
		funcs.funcNames = stringSliceUniq(param.FuncNames)
	}
	err = putRoleFunc(native, param.ContractAddr, param.Role, funcs)
	if err != nil {
		invokeEvent(native, "assignFuncsToRole", false)
		return utils.BYTE_FALSE, err
	}
	invokeEvent(native, "assignFuncsToRole", true)
	return utils.BYTE_TRUE, nil
}

func assignToRole(native *native.NativeService) ([]byte, error) {
	param := new(OntIDsToRoleParam)
	rd := bytes.NewReader(native.Input)
	if err := param.Deserialize(rd); err != nil {
		return nil, fmt.Errorf("deserialize failed, caused by %v", err)
	}
	if param.Role == nil {
		return nil, errors.NewErr("role is null")
	}

	//check admin's permission
	admin, err := getContractAdmin(native, param.ContractAddr)
	if err != nil {
		return nil, fmt.Errorf("get contract admin failed, caused by %v", err)
	}
	if admin == nil {
		invokeEvent(native, "assignOntIDsToRole", false)
		return utils.BYTE_FALSE, nil
	}
	if bytes.Compare(admin, param.AdminOntID) != 0 {
		invokeEvent(native, "assignOntIDsToRole", false)
		return utils.BYTE_FALSE, nil
	}
	valid, err := verifySig(native, param.AdminOntID, param.KeyNo)
	if err != nil {
		return nil, fmt.Errorf("verify admin's signature failed, caused by %v", err)
	}
	if !valid {
		invokeEvent(native, "assignOntIDsToRole", false)
		return utils.BYTE_FALSE, nil
	}

	//init an auth token
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
			return nil, err
		}
		if tokens == nil {
			tokens = new(roleTokens)
			tokens.tokens = make([]*AuthToken, 1)
			tokens.tokens[0] = token
		} else {
			ret, err := hasRole(native, param.ContractAddr, p, param.Role)
			if err != nil {
				return nil, err
			}
			if !ret {
				tokens.tokens = append(tokens.tokens, token)
			} else {
				continue
			}
		}
		err = putOntIDToken(native, param.ContractAddr, p, tokens)
		if err != nil {
			return nil, err
		}
	}
	invokeEvent(native, "assignOntIDsToRole", true)
	return utils.BYTE_TRUE, nil
}

func AssignOntIDsToRole(native *native.NativeService) ([]byte, error) {
	ret, err := assignToRole(native)
	return ret, err
}

func getAuthToken(native *native.NativeService, contractAddr, ontID, role []byte) (*AuthToken, error) {
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

func hasRole(native *native.NativeService, contractAddr, ontID, role []byte) (bool, error) {
	token, err := getAuthToken(native, contractAddr, ontID, role)
	if err != nil {
		return false, err
	}
	if token == nil {
		return false, nil
	}
	return true, nil
}

func getLevel(native *native.NativeService, contractAddr, ontID, role []byte) (uint8, error) {
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
func delegate(native *native.NativeService, contractAddr []byte, from []byte, to []byte,
	role []byte, period uint32, level uint8, keyNo uint64) ([]byte, error) {
	var fromHasRole, toHasRole bool
	var fromLevel uint8
	var fromExpireTime uint32
	//check from's permission
	ret, err := verifySig(native, from, keyNo)
	if err != nil {
		return nil, err
	}
	if !ret {
		invokeEvent(native, "delegate", false)
		return utils.BYTE_FALSE, nil
	}
	expireTime := uint32(time.Now().Unix())
	if period+expireTime < period {
		invokeEvent(native, "delegate", false)
		return utils.BYTE_FALSE, nil //invalid param
	}
	expireTime = expireTime + period

	fromToken, err := getAuthToken(native, contractAddr, from, role)
	if err != nil {
		return nil, err
	}
	if fromToken == nil {
		fromHasRole = false
		fromLevel = 0
	} else {
		fromHasRole = true
		fromLevel = fromToken.level
		fromExpireTime = fromToken.expireTime
	}
	toToken, err := getAuthToken(native, contractAddr, to, role)
	if err != nil {
		return nil, err
	}
	if toToken == nil {
		toHasRole = false
	} else {
		toHasRole = true
	}
	if !fromHasRole || toHasRole {
		invokeEvent(native, "delegate", false)
		return utils.BYTE_FALSE, nil
	}

	//check if 'from' has the permission to delegate
	if fromLevel == 2 {
		if level < fromLevel && level > 0 && expireTime < fromExpireTime {
			status, err := getDelegateStatus(native, contractAddr, to)
			if err != nil {
				return nil, err
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
				return nil, err
			}
			invokeEvent(native, "delegate", true)
			return utils.BYTE_TRUE, nil
		}
	}
	//TODO: for fromLevel > 2 case
	invokeEvent(native, "delegate", false)
	return utils.BYTE_FALSE, nil
}

func Delegate(native *native.NativeService) ([]byte, error) {
	param := &DelegateParam{}
	rd := bytes.NewReader(native.Input)
	err := param.Deserialize(rd)
	if err != nil {
		return nil, err
	}
	if param.Period > 1<<32 || param.Level > 1<<8 {
		return nil, fmt.Errorf("period or level is too large")
	}
	return delegate(native, param.ContractAddr, param.From, param.To, param.Role,
		uint32(param.Period), uint8(param.Level), param.KeyNo)
}

func withdraw(native *native.NativeService, contractAddr []byte, initiator []byte, delegate []byte,
	role []byte, keyNo uint64) (bool, error) {
	//check from's permission
	ret, err := verifySig(native, initiator, keyNo)
	if err != nil {
		return false, err
	}
	if !ret {
		return false, err
	}
	//code below only works in the case that initiator's level is 2
	//TODO
	initToken, err := getAuthToken(native, contractAddr, initiator, role)
	if err != nil {
		return false, err
	}
	if initToken == nil {
		return false, nil
	}
	status, err := getDelegateStatus(native, contractAddr, delegate)
	if err != nil {
		return false, err
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
	param := &WithdrawParam{}
	rd := bytes.NewReader(native.Input)
	param.Deserialize(rd)
	ret, err := withdraw(native, param.ContractAddr, param.Initiator, param.Delegate, param.Role, param.KeyNo)
	if err == nil {
		invokeEvent(native, "withdraw", ret)
		if ret {
			return utils.BYTE_TRUE, nil
		} else {
			return utils.BYTE_FALSE, nil
		}
	}
	return utils.BYTE_FALSE, err
}

/*
 *  VerifyToken(contractAddr []byte, caller []byte, fn []byte) (bool, error)
 *  @caller the ONT ID of the caller
 *  @fn the name of the func to call
 *  @tokenSig the signature on the message
 */
func verifyToken(native *native.NativeService, contractAddr []byte, caller []byte, fn string, keyNo uint64) (bool, error) {
	//check caller's identity
	ret, err := verifySig(native, caller, keyNo)
	if err != nil {
		return false, err
	}
	if !ret {
		return false, nil
	}
	tokens, err := getOntIDToken(native, contractAddr, caller)
	if err != nil {
		return false, err
	}
	if tokens != nil {
		for _, token := range tokens.tokens {
			funcs, err := getRoleFunc(native, contractAddr, token.role)
			if err != nil {
				return false, err
			}
			for _, f := range funcs.funcNames {
				if strings.Compare(fn, f) == 0 {
					return true, nil
				}
			}
		}
	}

	status, err := getDelegateStatus(native, contractAddr, caller)
	if err != nil {
		return false, nil
	}
	if status != nil {
		for _, s := range status.status {
			funcs, err := getRoleFunc(native, contractAddr, s.role)
			if err != nil {
				return false, nil
			}
			for _, f := range funcs.funcNames {
				if strings.Compare(fn, f) == 0 {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func VerifyToken(native *native.NativeService) ([]byte, error) {
	param := &VerifyTokenParam{}
	rd := bytes.NewReader(native.Input)
	err := param.Deserialize(rd)
	if err != nil {
		return nil, err
	}
	ret, err := verifyToken(native, param.ContractAddr, param.Caller, param.Fn, param.KeyNo)
	if err != nil {
		return nil, err
	}
	if !ret {
		invokeEvent(native, "verifyToken", false)
		return utils.BYTE_FALSE, nil
	}
	invokeEvent(native, "verifyToken", true)
	return utils.BYTE_TRUE, nil
}

func verifySig(native *native.NativeService, ontID []byte, keyNo uint64) (bool, error) {
	bf := new(bytes.Buffer)
	if err := serialization.WriteVarBytes(bf, ontID); err != nil {
		return false, err
	}
	if err := serialization.WriteVarUint(bf, keyNo); err != nil {
		return false, err
	}
	args := bf.Bytes()
	ret, err := native.NativeCall(utils.OntIDContractAddress, "verifySignature", args)
	if err != nil {
		return false, err
	}
	valid, ok := ret.([]byte)
	if !ok {
		return false, errors.NewErr("verifySignature return non-bool value")
	}
	if bytes.Compare(valid, utils.BYTE_TRUE) == 0 {
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

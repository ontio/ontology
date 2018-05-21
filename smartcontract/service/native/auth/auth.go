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

	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

var (
	null = []byte{}
)

func Init() {
<<<<<<< HEAD
	Contracts[utils.AuthContractAddress] = RegisterAuthContract
=======
	native.Contracts[genesis.AuthContractAddress] = RegisterAuthContract
>>>>>>> 8c02ad42... refactor the storage model of auth contract
}

/*
 * contract admin management
 */
func initContractAdmin(native *native.NativeService, contractAddr, ontID []byte) (bool, error) {
	admin, err := GetContractAdmin(native, contractAddr)
	if err != nil {
		return false, err
	}
	if admin != nil {
		//admin is already set, just return
		return false, nil
	}

	adminKey, err := GetContractAdminKey(native, contractAddr)
	if err != nil {
		return false, err
	}
	utils.PutBytes(native, adminKey, ontID)

	/* push notify
	event := new(event.NotifyEventInfo)
	event.ContractAddress = native.ContextRef.CurrentContext().ContractAddress
	event.States = ontID
	native.Notifications = append(native.Notifications, event)
	*/
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
	return utils.BYTE_TRUE, nil
}

func transfer(native *native.NativeService, contractAddr, newAdminOntID []byte, keyNo uint32) (bool, error) {
	admin, err := GetContractAdmin(native, contractAddr)
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

	adminKey, err := GetContractAdminKey(native, contractAddr)
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
		return utils.BYTE_TRUE, nil
	} else {
		return utils.BYTE_FALSE, nil
	}
}

/*
 * role -> []FuncName
 */
func AssignFuncsToRole(native *native.NativeService) ([]byte, error) {
	//deserialize input param
	param := new(FuncsToRoleParam)
	rd := bytes.NewReader(native.Input)
	if err := param.Deserialize(rd); err != nil {
		return nil, fmt.Errorf("deserialize param failed, caused by %v", err)
	}
	if param.Role == nil {
		return utils.BYTE_FALSE, nil
	}

	//check the caller's permission
	admin, err := GetContractAdmin(native, param.ContractAddr)
	if err != nil {
		return nil, fmt.Errorf("get contract admin failed, caused by %v", err)
	}
	if admin == nil {
		return utils.BYTE_FALSE, nil
	}
	if bytes.Compare(admin, param.AdminOntID) != 0 {
		return utils.BYTE_FALSE, nil
	}
	ret, err := verifySig(native, param.AdminOntID, param.KeyNo)
	if err != nil {
		return nil, fmt.Errorf("verify admin's signature failed, caused by %v", err)
	}
	if !ret {
		return utils.BYTE_FALSE, nil
	}

	funcs, err := GetRoleFunc(native, param.ContractAddr, param.Role)
	if funcs != nil {
		funcNames := append(funcs.funcNames, param.FuncNames...)
		funcs.funcNames = stringSliceUniq(funcNames)
	} else {
		funcs = new(roleFuncs)
		funcs.funcNames = param.FuncNames
	}

	err = PutRoleFunc(native, param.ContractAddr, param.Role, funcs)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
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
	admin, err := GetContractAdmin(native, param.ContractAddr)
	if err != nil {
		return nil, fmt.Errorf("get contract admin failed, caused by %v", err)
	}
	if admin == nil {
		return utils.BYTE_FALSE, nil
	}
	if bytes.Compare(admin, param.AdminOntID) != 0 {
		return nil, fmt.Errorf("not invoked by contract admin")
	}
	valid, err := verifySig(native, param.AdminOntID, param.KeyNo)
	if err != nil {
		return nil, fmt.Errorf("verify admin's signature failed, caused by %v", err)
	}
	if !valid {
		return utils.BYTE_FALSE, nil
	}

	//init an auth token
	token := new(AuthToken)
	future := time.Date(2100, 1, 1, 12, 0, 0, 0, time.UTC)
	token.expireTime = uint32(future.Unix())
	token.level = 2
	token.role = param.Role
	for _, p := range param.Persons {
		if p == nil {
			continue
		}

		tokens, err := GetOntIDToken(native, param.ContractAddr, p)
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
		err = PutOntIDToken(native, param.ContractAddr, p, tokens)
		if err != nil {
			return nil, err
		}
	}
	return utils.BYTE_TRUE, nil
}

func AssignOntIDsToRole(native *native.NativeService) ([]byte, error) {
	ret, err := assignToRole(native)
	return ret, err
}

func getAuthToken(native *native.NativeService, contractAddr, ontID, role []byte) (*AuthToken, error) {
	tokens, err := GetOntIDToken(native, contractAddr, ontID)
	if err != nil {
		return nil, fmt.Errorf("get token failed, caused by %v", err)
	}
	for _, token := range tokens.tokens {
		if bytes.Compare(token.role, role) == 0 { //permanent
			return token, nil
		}
	}
	status, err := GetDelegateStatus(native, contractAddr, ontID)
	if err != nil {
		return nil, fmt.Errorf("get delegate status failed, caused by %v", err)
	}
	for _, s := range status.status {
		if bytes.Compare(s.role, role) == 0 && native.Time < s.expireTime { //temporarily
			token := new(AuthToken)
			token.role = s.role
			token.level = s.level
			token.expireTime = s.expireTime
			return token, nil
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

func getLevel(native *native.NativeService, contractAddr, ontID, role []byte) (uint32, error) {
	token, err := getAuthToken(native, contractAddr, ontID, role)
	if err != nil {
		return 0, err
	}
	if token == nil {
		return 0, nil
	}
	return uint32(token.level), nil
}

/*
 * if 'from' has the authority and 'to' has not been authorized 'role',
 * then make changes to storage as follows:
 *   - insert 'to' into the linked list of 'role -> [] persons'
 *   - insert 'to' into the delegate list of 'from'.
 *   - put 'func x to' into the 'func x person' table
 */
func delegate(native *native.NativeService, contractAddr []byte, from []byte, to []byte,
	role []byte, period uint32, level uint, keyNo uint32) ([]byte, error) {
	var fromHasRole, toHasRole bool
	var fromLevel uint
	var fromExpireTime uint32
	//check from's permission
	ret, err := verifySig(native, from, keyNo)
	if err != nil {
		return nil, err
	}
	if !ret {
		return nil, err
	}
	expireTime := uint32(time.Now().Unix())
	if period+expireTime < period {
		return utils.BYTE_FALSE, err
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
		fromLevel = uint(fromToken.level)
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
		return nil, fmt.Errorf("from has no %x or to already has %x", role, role) //ignore
	}

	if err != nil {
		return nil, fmt.Errorf("")
	}

	//check if 'from' has the permission to delegate
	if fromLevel == 2 {
		if level < fromLevel && level > 0 && expireTime < fromExpireTime {
			status, err := GetDelegateStatus(native, contractAddr, to)
			if err != nil {
				return nil, err
			}
			newStatus := &DelegateStatus{
				root: from,
			}
			newStatus.expireTime = expireTime
			newStatus.role = role
			newStatus.level = uint8(level)
			status.status = append(status.status, newStatus)
			err = PutDelegateStatus(native, contractAddr, to, status)
			if err != nil {
				return nil, err
			}
			return utils.BYTE_TRUE, nil
		}
	}
	//TODO:
	/*
		if fromLevel > 2 && level < fromLevel && level > 0 {

		}
	*/
	return utils.BYTE_FALSE, nil
}

func Delegate(native *native.NativeService) ([]byte, error) {
	param := &DelegateParam{}
	rd := bytes.NewReader(native.Input)
	err := param.Deserialize(rd)
	if err != nil {
		return nil, err
	}
	return delegate(native, param.ContractAddr, param.From, param.To, param.Role,
		param.Period, param.Level, param.KeyNo)
}

func withdraw(native *native.NativeService, contractAddr []byte, initiator []byte, delegate []byte,
	role []byte, keyNo uint32) ([]byte, error) {
	//check from's permission
	ret, err := verifySig(native, initiator, keyNo)
	if err != nil {
		return nil, err
	}
	if !ret {
		return nil, err
	}

	initToken, err := getAuthToken(native, contractAddr, initiator, role)
	if err != nil {
		return nil, err
	}
	if initToken == nil {
		return utils.BYTE_FALSE, nil
	}
	status, err := GetDelegateStatus(native, contractAddr, delegate)
	if err != nil {
		return nil, err
	}
	for i, s := range status.status {
		if bytes.Compare(s.role, role) == 0 &&
			bytes.Compare(s.root, initiator) == 0 {
			newStatus := new(Status)
			newStatus.status = append(status.status[:i], status.status[i+1:]...)
			err = PutDelegateStatus(native, contractAddr, delegate, newStatus)
			if err != nil {
				return nil, err
			}
			return utils.BYTE_TRUE, nil
		}
	}
	return utils.BYTE_FALSE, nil
}

func Withdraw(native *native.NativeService) ([]byte, error) {
	param := &WithdrawParam{}
	rd := bytes.NewReader(native.Input)
	param.Deserialize(rd)
	return withdraw(native, param.ContractAddr, param.Initiator, param.Delegate, param.Role, param.KeyNo)
}

/*
 *  VerifyToken(contractAddr []byte, caller []byte, fn []byte) (bool, error)
 *  @caller the ONT ID of the caller
 *  @fn the name of the func to call
 *  @tokenSig the signature on the message
 */

func verifyToken(native *native.NativeService, contractAddr []byte, caller []byte, fn []byte, keyNo uint32) (bool, error) {
	//check caller's identity
	ret, err := verifySig(native, caller, keyNo)
	if err != nil {
		return false, err
	}
	if !ret {
		return false, nil
	}

	tokens, err := GetOntIDToken(native, contractAddr, caller)
	if err != nil {
		return false, nil
	}
	if tokens != nil {
		for _, token := range tokens.tokens {
			funcs, err := GetRoleFunc(native, contractAddr, token.role)
			if err != nil {
				return false, nil
			}
			for _, f := range funcs.funcNames {
				if bytes.Compare(fn, []byte(f)) == 0 {
					return true, nil
				}
			}
		}
	}

	status, err := GetDelegateStatus(native, contractAddr, caller)
	if err != nil {
		return false, nil
	}
	if status != nil {
		for _, s := range status.status {
			funcs, err := GetRoleFunc(native, contractAddr, s.role)
			if err != nil {
				return false, nil
			}
			for _, f := range funcs.funcNames {
				if bytes.Compare(fn, []byte(f)) == 0 {
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
	ret, err := verifyToken(native, param.Caller, param.ContractAddr, param.Fn, param.KeyNo)
	if err != nil {
		return nil, err
	}
	if !ret {
		return utils.BYTE_FALSE, nil
	}
	return utils.BYTE_TRUE, nil
}
func verifySig(native *native.NativeService, ontID []byte, keyNo uint32) (bool, error) {
	//return false, nil
	// enable signature verification whenever the OntID contract is merged into the master branch
	bf := new(bytes.Buffer)
	if err := serialization.WriteVarBytes(bf, ontID); err != nil {
		return false, err
	}
	if err := serialization.WriteUint32(bf, keyNo); err != nil {
		return false, err
	}
	args := bf.Bytes()
	ret, err := native.ContextRef.AppCall(utils.OntIDContractAddress, "verifySignature", []byte{}, args)
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
	//native.Register("delegate", Delegate)
	//native.Register("withdraw", Withdraw)
	native.Register("assignOntIDsToRole", AssignOntIDsToRole)
	//native.Register("verifyToken", VerifyToken)
	native.Register("transfer", Transfer)
	//native.Register("test", AuthTest)
}

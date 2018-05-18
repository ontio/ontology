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
	"github.com/ontio/ontology/smartcontract/event"
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

func GetContractAdmin(native *native.NativeService, contractAddr []byte) ([]byte, error) {
	key, err := GetContractAdminKey(native, contractAddr)
	if err != nil {
		return nil, err
	}
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return item.Value, nil
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

	event := new(event.NotifyEventInfo)
	event.ContractAddress = native.ContextRef.CurrentContext().ContractAddress
	event.States = ontID
	native.Notifications = append(native.Notifications, event)
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
	var funcNames []string

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

	funcs := new(roleFuncs)
	roleF, err := GetRoleFKey(native, param.ContractAddr, param.Role)
	if err != nil {
		return nil, err
	}
	funcsItem, err := utils.GetStorageItem(native, roleF)
	if funcsItem != nil {
		rd = bytes.NewReader(funcsItem.Value)
		err = funcs.Deserialize(rd)
		if err != nil {
			return nil, fmt.Errorf("read funcNames failed, caused by %v", err)
		}
		funcNames = append(funcs.funcNames, param.FuncNames...)
		funcNames = stringSliceUniq(funcNames)
	} else {
		funcNames = param.FuncNames
	}
	funcs.funcNames = funcNames

	//store
	bf := new(bytes.Buffer)
	err = funcs.Serialize(bf)
	if err != nil {
		return nil, fmt.Errorf("serialize roleFuncs failed, caused by %v", err)
	}
	utils.PutBytes(native, roleF, bf.Bytes())
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

	roleP, err := GetRolePKey(native, param.ContractAddr, param.Role)
	if err != nil {
		return nil, fmt.Errorf("get RoleP failed with role=%x", param.Role)
	}
	for _, p := range param.Persons {
		if p == nil {
			continue
		}
		//init an auth token
		token := new(AuthToken)
		future := time.Date(2100, 1, 1, 12, 0, 0, 0, time.UTC)
		token.expireTime = uint32(future.Unix())
		token.level = 2
		token.role = param.Role

		tokens := new(roleTokens)
		item, err := utils.GetStorageItem(native, append(roleP, p...))
		if err != nil {
			return nil, fmt.Errorf("get roleTokens failed, caused by %v", err)
		}
		if item == nil {
			tokens.tokens = make([]*AuthToken, 1)
			tokens.tokens[0] = token
		} else {
			rd := bytes.NewReader(item.Value)
			err = tokens.Deserialize(rd)
			if err != nil {
				return nil, fmt.Errorf("deserialize failed")
			}
			flag := true
			for _, role := range tokens.tokens {
				if bytes.Compare(role.role, param.Role) == 0 {
					flag = false
					break
				}
			}
			if flag {
				tokens.tokens = append(tokens.tokens, token)
			} else {
				continue
			}
		}
		w := new(bytes.Buffer)
		err = tokens.Serialize(w)
		if err != nil {
			return nil, fmt.Errorf("")
		}
		utils.PutBytes(native, append(roleP, p...), w.Bytes())

	}
	return utils.BYTE_TRUE, nil
}

func AssignOntIDsToRole(native *native.NativeService) ([]byte, error) {
	ret, err := assignToRole(native)
	return ret, err
}

/*
 * if 'from' has the authority and 'to' has not been authorized 'role',
 * then make changes to storage as follows:
 *   - insert 'to' into the linked list of 'role -> [] persons'
 *   - insert 'to' into the delegate list of 'from'.
 *   - put 'func x to' into the 'func x person' table
 */

/*
func delegate(native *native.NativeService, contractAddr []byte, from []byte, to []byte,
	role []byte, period uint32, level uint, keyNo uint32) ([]byte, error) {
	//check from's permission
	ret, err := verifySig(native, from, keyNo)
	if err != nil {
		return nil, err
	}
	if !ret {
		return nil, err
	}
	//iterate
	roleP, err := GetRolePKey(native, contractAddr, role)
	if err != nil {
		return nil, err
	}
	node_to, err := utils.LinkedlistGetItem(native, roleP, to)
	if err != nil {
		return nil, err
	}
	node_from, err := utils.LinkedlistGetItem(native, roleP, from)
	if err != nil {
		return nil, err
	}

	if node_to != nil || node_from == nil {
		return nil, fmt.Errorf("from has no %x or to already has %x", role, role) //ignore
	}

	auth := new(AuthToken)
	err = auth.deserialize(node_from.GetPayload())
	if err != nil {
		return nil, err
	}
	//check if 'from' has the permission to delegate
	if auth.level >= 2 && level < auth.level && level > 0 {
		//  put `to` in the delegate list
		//  'role x person' -> [] person is a list of person who Gets the auth token
		deleFrom, err := GetDelegateListKey(native, contractAddr, role, from)
		if err != nil {
			return nil, err
		}
		err = utils.LinkedlistInsert(native, deleFrom, to, null)
		if err != nil {
			return nil, err
		}

		//insert into the `role -> []person` list
		authPrime := new(AuthToken)
		authPrime.expireTime = native.Time + period
		if authPrime.expireTime < native.Time {
			return nil, errors.NewErr("[auth contract] overflow of expire time")
		}
		authPrime.level = level
		serAuthPrime, err := authPrime.serialize()
		if err != nil {
			return nil, err
		}
		persons := &OntIDsToRoleParam{ContractAddr: contractAddr, Role: role, Persons: [][]byte{to}}
		bf := new(bytes.Buffer)
		err = persons.Serialize(bf)
		if err != nil {
			return nil, err
		}
		native.Input = bf.Bytes()
		ret, err := assignToRole(native, serAuthPrime)
		if err != nil {
			return nil, err
		}
		//log.Tracef("assign to role return %x", ret)
		return ret, nil
	}
	return utils.BYTE_TRUE, nil
}
*/
/*
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
	roleF, err := GetRoleFKey(native, contractAddr, role)
	if err != nil {
		return nil, err
	}
	now := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	expireAuth, err := (&AuthToken{expireTime: uint32(now.Unix()), level: 0}).serialize()
	if err != nil {
		return nil, err
	}
	fn, err := utils.LinkedlistGetHead(native, roleF) //fn is a func name
	funcs := [][]byte{}
	for {
		if fn == nil {
			break
		}
		nodef, err := utils.LinkedlistGetItem(native, roleF, fn) //
		if err != nil {
			return nil, err
		}
		funcs = append(funcs, fn)
		fn = nodef.GetNext()
	}
	ret, err := verifySig(native, initiator, keyNo)
	if err != nil {
		return nil, err
	}
	if !ret {
		return utils.BYTE_FALSE, nil
	}
	//check if initiator has the right to withdraw
	fromKey, err := GetDelegateListKey(native, contractAddr, role, initiator)
	if err != nil {
		return nil, err
	}
	node_dele, err := utils.LinkedlistGetItem(native, fromKey, delegate)
	if err != nil {
		return nil, err
	}
	if node_dele == nil {
		return nil, errors.NewErr("[auth contract] initiator does not have the right")
	}
	//clear
	wdList := [][]byte{delegate}
	for {
		if len(wdList) == 0 {
			break
		}
		per := wdList[0]
		wdList = wdList[1:]

		for _, fn := range funcs {
			err = writeAuthToken(native, contractAddr, fn, per, expireAuth)
			if err != nil {
				return nil, err
			}
		}
		roleP, err := GetRolePKey(native, contractAddr, role)
		if err != nil {
			return nil, err
		}
		utils.LinkedlistDelete(native, roleP, per)
		// iterate over 'role x person -> [] person' list
		ser_role_per, err := GetDelegateListKey(native, contractAddr, role, per)
		if err != nil {
			return nil, err
		}
		q, err := utils.LinkedlistGetHead(native, ser_role_per)
		if err != nil {
			return nil, err
		}
		for {
			if q == nil {
				break
			}
			wdList = append(wdList, q)
			item, err := utils.LinkedlistGetItem(native, ser_role_per, q)
			if err != nil {
				return nil, err
			}
			_, err = utils.LinkedlistDelete(native, ser_role_per, q)
			if err != nil {
				return nil, err
			}
			q = item.GetNext()
		}
	}
	return utils.BYTE_TRUE, nil
}

func Withdraw(native *native.NativeService) ([]byte, error) {
	param := &WithdrawParam{}
	rd := bytes.NewReader(native.Input)
	param.Deserialize(rd)
	return withdraw(native, param.ContractAddr, param.Initiator, param.Delegate, param.Role, param.KeyNo)
}
*/
/*
 *  VerifyToken(contractAddr []byte, caller []byte, fn []byte) (bool, error)
 *  @caller the ONT ID of the caller
 *  @fn the name of the func to call
 *  @tokenSig the signature on the message
 */

/*
 func verifyToken(native *native.NativeService, contractAddr []byte, caller []byte, fn []byte, keyNo uint32) (bool, error) {
	//check caller's identity
	ret, err := verifySig(native, caller, keyNo)
	if err != nil {
		return false, err
	}
	if !ret {
		return false, nil
	}

	//check if caller has the auth to call 'fn'
	ser_fn_per, err := GetFuncOntIDKey(native, contractAddr, fn, caller)
	if err != nil {
		return false, err
	}
	//FIXME
	data, err := utils.GetStorageItem(native, ser_fn_per)
	if err != nil {
		return false, err
	}
	auth := new(AuthToken)
	err = auth.deserialize(data.Value)
	if err != nil {
		return false, err
	}
	if auth.level >= 1 && native.Time < auth.expireTime {
		return true, nil
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
*/
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

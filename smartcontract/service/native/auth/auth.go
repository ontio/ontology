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
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	. "github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"time"
)

var (
	null = []byte{}
)

func Init() {
	native.Contracts[genesis.AuthContractAddress] = RegisterAuthContract
}

func GetContractAdmin(native *NativeService, contractAddr []byte) ([]byte, error) {
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

func initContractAdmin(native *NativeService, contractAddr, ontID []byte) (bool, error) {
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
	PutBytes(native, adminKey, ontID)
	return true, nil
}

func InitContractAdmin(native *NativeService) ([]byte, error) {
	param := new(InitContractAdminParam)
	rd := bytes.NewReader(native.Input)
	if err := param.Deserialize(rd); err != nil {
		return nil, err
	}

	/*
		cxt := native.ContextRef.CallingContext()
		if cxt == nil {
			return nil, errors.NewErr("no calling context")
		}
		invokeAddr := cxt.ContractAddress
	*/

	invokeAddr := genesis.OntContractAddress
	ret, err := initContractAdmin(native, invokeAddr[:], param.AdminOntID)
	if err != nil {
		return nil, err
	}
	if !ret {
		return utils.BYTE_FALSE, nil
	}
	return utils.BYTE_TRUE, nil
}

func transfer(native *NativeService, contractAddr, newAdminOntID []byte, keyNo uint32) (bool, error) {
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
	PutBytes(native, adminKey, newAdminOntID)
	return true, nil
}

func Transfer(native *NativeService) ([]byte, error) {
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
 * role -> []ONT ID
 *
 * funcName x ONT ID -> (expiration time, authorized?)
 */
func AssignFuncsToRole(native *NativeService) ([]byte, error) {

	//deserialize input param
	param := new(FuncsToRoleParam)
	rd := bytes.NewReader(native.Input)
	if err := param.Deserialize(rd); err != nil {
		return nil, fmt.Errorf("deserialize failed, caused by %v", err)
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
		return nil, fmt.Errorf("verify admin's signature failed, caused by %d", err)
	}
	if !ret {
		return utils.BYTE_FALSE, nil
	}

	//insert funcnames into role->func Linkedlist
	roleF, err := GetRoleFKey(native, param.ContractAddr, param.Role)
	if err != nil {
		return nil, err
	}
	for _, fn := range param.FuncNames {
		if fn == "" {
			continue
		}
		err = LinkedlistInsert(native, roleF, []byte(fn), null)
		if err != nil {
			return nil, fmt.Errorf("insert into role->[]func failed. fn: %s", fn)
		}
	}
	// insert into 'func x person' table
	roleP, err := GetRolePKey(native, param.ContractAddr, param.Role)
	if err != nil {
		return nil, err
	}
	head, err := LinkedlistGetHead(native, roleP)
	if err != nil {
		return nil, err
	}
	q := head //q is an ont ID
	for {
		if q == nil {
			break
		}
		nodeq, err := LinkedlistGetItem(native, roleP, q)
		if err != nil {
			return nil, err
		}
		for _, fn := range param.FuncNames {
			err = writeAuthToken(native, param.ContractAddr, []byte(fn), q, nodeq.payload)
			if err != nil {
				return nil, err
			}
		}
		q = nodeq.next
	}
	return utils.BYTE_TRUE, nil
}

func assignToRole(native *NativeService, auth []byte) ([]byte, error) {
	param := new(OntIDsToRoleParam)
	rd := bytes.NewReader(native.Input)
	if err := param.Deserialize(rd); err != nil {
		return nil, fmt.Errorf("deserialize failed, caused by %v", err)
	}
	if param.Role == nil {
		return nil, errors.NewErr("role is null")
	}

	roleP, err := GetRolePKey(native, param.ContractAddr, param.Role)
	if err != nil {
		return nil, errors.NewErr("get RoleP failed")
	}
	roleF, err := GetRoleFKey(native, param.ContractAddr, param.Role)
	if err != nil {
		return nil, errors.NewErr("get RoleF failed")
	}
	//insert persons into role->person Linkedlist
	for _, p := range param.Persons {
		if p == nil {
			continue
		}
		err = LinkedlistInsert(native, roleP, p, auth)
		if err != nil {
			return nil, err
		}
	}

	fn, err := LinkedlistGetHead(native, roleF) //fn is a func name
	for {
		if fn == nil {
			break
		}
		nodef, err := LinkedlistGetItem(native, roleF, fn)
		if err != nil {
			return nil, err
		}
		for _, per := range param.Persons {
			err = writeAuthToken(native, param.ContractAddr, fn, per, auth)
		}
		fn = nodef.next
	}
	return utils.BYTE_TRUE, nil
}

func AssignOntIDsToRole(native *NativeService) ([]byte, error) {
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
		return nil, errors.NewErr("get contract admin failed")
	}
	if admin == nil {
		return utils.BYTE_FALSE, nil
	}
	if bytes.Compare(admin, param.AdminOntID) != 0 {
		return nil, errors.NewErr("not invoked by contract admin")
	}
	valid, err := verifySig(native, param.AdminOntID, param.KeyNo)
	if err != nil {
		return nil, fmt.Errorf("verify admin's signature failed, caused by %d", err)
	}
	if !valid {
		return utils.BYTE_FALSE, nil
	}
	future := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	auth, err := (&AuthToken{expireTime: uint32(future.Unix()), level: 10}).serialize()
	if err != nil {
		return nil, err
	}
	ret, err := assignToRole(native, auth)
	return ret, err
}

/*
 * if 'from' has the authority and 'to' has not been authorized 'role',
 * then make changes to storage as follows:
 *   - insert 'to' into the linked list of 'role -> [] persons'
 *   - insert 'to' into the delegate list of 'from'.
 *   - put 'func x to' into the 'func x person' table
 */
func delegate(native *NativeService, contractAddr []byte, from []byte, to []byte,
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
	node_to, err := LinkedlistGetItem(native, roleP, to)
	if err != nil {
		return nil, err
	}
	node_from, err := LinkedlistGetItem(native, roleP, from)
	if err != nil {
		return nil, err
	}

	if node_to != nil || node_from == nil {
		return nil, fmt.Errorf("from has no %x or to already has %x", role, role) //ignore
	}

	auth := new(AuthToken)
	err = auth.deserialize(node_from.payload)
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
		err = LinkedlistInsert(native, deleFrom, to, null)
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

func Delegate(native *NativeService) ([]byte, error) {
	param := &DelegateParam{}
	rd := bytes.NewReader(native.Input)
	err := param.Deserialize(rd)
	if err != nil {
		return nil, err
	}
	return delegate(native, param.ContractAddr, param.From, param.To, param.Role,
		param.Period, param.Level, param.KeyNo)
}

func withdraw(native *NativeService, contractAddr []byte, initiator []byte, delegate []byte,
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
	fn, err := LinkedlistGetHead(native, roleF) //fn is a func name
	funcs := [][]byte{}
	for {
		if fn == nil {
			break
		}
		nodef, err := LinkedlistGetItem(native, roleF, fn) //
		if err != nil {
			return nil, err
		}
		funcs = append(funcs, fn)
		fn = nodef.next
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
	node_dele, err := LinkedlistGetItem(native, fromKey, delegate)
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
		LinkedlistDelete(native, roleP, per)
		// iterate over 'role x person -> [] person' list
		ser_role_per, err := GetDelegateListKey(native, contractAddr, role, per)
		if err != nil {
			return nil, err
		}
		q, err := LinkedlistGetHead(native, ser_role_per)
		if err != nil {
			return nil, err
		}
		for {
			if q == nil {
				break
			}
			wdList = append(wdList, q)
			item, err := LinkedlistGetItem(native, ser_role_per, q)
			if err != nil {
				return nil, err
			}
			_, err = LinkedlistDelete(native, ser_role_per, q)
			if err != nil {
				return nil, err
			}
			q = item.next
		}
	}
	return utils.BYTE_TRUE, nil
}

func Withdraw(native *NativeService) ([]byte, error) {
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

func verifyToken(native *NativeService, contractAddr []byte, caller []byte, fn []byte, keyNo uint32) (bool, error) {
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

func VerifyToken(native *NativeService) ([]byte, error) {
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

func verifySig(native *NativeService, ontID []byte, keyNo uint32) (bool, error) {
	return false, nil
	/* enable signature verification whenever the OntID contract is merged into the master branch
	bf := new(bytes.Buffer)
	if err := serialization.WriteVarBytes(bf, ontID); err != nil {
		return false, err
	}
	if err := serialization.WriteUint32(bf, keyNo); err != nil {
		return false, err
	}
	args := bf.Bytes()
	//ontIDContract := []byte{}
	ret, err := native.ContextRef.AppCall(genesis.AuthContractAddress, "verifySignature", []byte{}, args)
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
	*/
}

func PackKeys(field []byte, items [][]byte) ([]byte, error) {
	w := new(bytes.Buffer)
	for _, item := range items {
		err := serialization.WriteVarBytes(w, item)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[AuthContract] PackKeys failed")
		}
	}
	key := append(field, w.Bytes()...)
	return key, nil
}

//pack data to be used as a key in the kv storage
// key := field || ser_data
func packKey(field []byte, data []byte) ([]byte, error) {
	return PackKeys(field, [][]byte{data})
}

func RegisterAuthContract(native *NativeService) {
	native.Register("initContractAdmin", InitContractAdmin)
	native.Register("assignFuncsToRole", AssignFuncsToRole)
	native.Register("delegate", Delegate)
	native.Register("withdraw", Withdraw)
	native.Register("assignOntIDsToRole", AssignOntIDsToRole)
	native.Register("verifyToken", VerifyToken)
	native.Register("transfer", Transfer)
	//native.Register("test", AuthTest)
}

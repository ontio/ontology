package auth

import (
	"bytes"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	. "github.com/ontio/ontology/smartcontract/service/native"
	"time"
)

var RoleF = []byte{0x01}
var RoleP = []byte{0x02}
var FuncPerson = []byte{0x03}
var DelegateList = []byte{0x04}
var Admin = []byte{0x05}

var BYTE_ONE = []byte{0x01}
var BYTE_ZERO = []byte{0x00}

func init() {
	native.Contracts[genesis.AuthContractAddress] = RegisterAuthContract
}

/*
 * contract admin management
 */

//can be called only once

func initContractAdmin(native *NativeService, contractAddr, ontID []byte) (bool, error) {
	//null := []byte{}
	admin, err := getContractAdmin(native, contractAddr)
	if err != nil {
		return false, err
	}
	if admin != nil {
		//admin is already set.
		return false, nil
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	adminKey, err := packKeys(contract[:], [][]byte{contractAddr, Admin})
	if err != nil {
		return false, err
	}
	writeBytes(native, append(contract[:], adminKey...), ontID)
	return true, nil
}

func InitContractAdmin(native *NativeService) ([]byte, error) {
	cxt := native.ContextRef.CallingContext()
	if cxt == nil {
		log.Error("no calling context")
		return nil, errors.NewErr("no calling context")
	}
	invokeAddr := cxt.ContractAddress
	invokeAddr = genesis.OntContractAddress
	param := new(InitContractAdminParam)
	rd := bytes.NewReader(native.Input)
	param.Deserialize(rd)
	ret, err := initContractAdmin(native, invokeAddr[:], param.AdminOntID)
	if err != nil {
		return BYTE_ZERO, err
	}
	if !ret {
		return BYTE_ZERO, nil
	}
	return BYTE_ONE, nil
}

func getContractAdmin(native *NativeService, contractAddr []byte) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	adminKey, err := packKeys(contract[:], [][]byte{contractAddr, Admin}) //contract.contractAddr.Admin
	if err != nil {
		return nil, err
	}
	val, err := native.CloneCache.Get(scommon.ST_STORAGE, adminKey)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}
	admin, ok := val.(*cstates.StorageItem)
	if !ok {
		return nil, errors.NewErr("")
	}
	return admin.Value, nil
}

func transfer(native *NativeService, contractAddr, newAdminOntID []byte) (bool, error) {
	//null := []byte{}
	admin, err := getContractAdmin(native, contractAddr)
	if err != nil {
		return false, err
	}
	ret, err := verifySig(native, admin)
	if err != nil {
		return false, err
	}
	if !ret {
		return false, nil
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	adminKey, err := packKeys(contract[:], [][]byte{contractAddr, Admin})
	if err != nil {
		return false, err
	}
	writeBytes(native, append(contract[:], adminKey...), newAdminOntID)
	return true, nil
}

func Transfer(native *NativeService) ([]byte, error) {
	param := new(TransferParam)
	rd := bytes.NewReader(native.Input)
	err := param.Deserialize(rd)
	if err != nil {
		return nil, err
	}
	ret, err := transfer(native, param.ContractAddr, param.NewAdminOntID)
	if ret {
		return BYTE_ONE, nil
	} else {
		return BYTE_ZERO, nil
	}
}

//type FuncsToRole struct {
//	FuncsToRoleParam
//	sig []byte
//}
//
//func (this *FuncsToRole) Serialize(bf io.Writer) error {
//	err := this.FuncsToRoleParam.Serialize(bf)
//	if err != nil {
//		return err
//	}
//	err = serialization.WriteVarBytes(bf, this.sig)
//	return err
//}
//
//func (this *FuncsToRole) Deserialize(rd io.Reader) error {
//	err := this.FuncsToRoleParam.Deserialize(rd)
//	if err != nil {
//		return err
//	}
//	this.sig, err = serialization.ReadVarBytes(rd)
//	if err != nil {
//		return err
//	}
//	return nil
//}

func writeBytes(native *NativeService, key []byte, value []byte) {
	native.CloneCache.Add(scommon.ST_STORAGE, key, &cstates.StorageItem{Value: value})
}

func writeAuthToken(native *NativeService, contractAddr, fn, ontID, auth []byte) error {
	contractKey := native.ContextRef.CurrentContext().ContractAddress[:]
	contractKey = append(contractKey, contractAddr...)
	key, err := packKeys(FuncPerson, [][]byte{fn, ontID}) //fn x ontID
	if err != nil {
		return err
	}
	// construct auth token
	writeBytes(native, append(contractKey[:], key...), auth)
	return nil
}

/*
 * role -> []FuncName
 * role -> []ONT ID
 *
 * funcName x ONT ID -> (expiration time, authorized?)
 */

func AssignFuncsToRole(native *NativeService) ([]byte, error) {
	curContract := native.ContextRef.CurrentContext().ContractAddress
	null := []byte{}
	//deserialize input param
	param := new(FuncsToRoleParam)
	rd := bytes.NewReader(native.Input)
	if err := param.Deserialize(rd); err != nil {
		return BYTE_ZERO, err
	}
	if param.Role == nil {
		return BYTE_ZERO, errors.NewErr("")
	}

	//check the caller's permission
	admin, err := getContractAdmin(native, param.ContractAddr)
	if err != nil {
		return nil, err
	}
	if bytes.Compare(admin, param.AdminOntID) != 0 {
		return nil, errors.NewErr("")
	}
	ret, err := verifySig(native, param.AdminOntID)
	if err != nil {
		return nil, err
	}
	if !ret {
		return BYTE_ONE, nil
	}

	//insert funcnames into role->func linkedlist
	contractKey := curContract[:]
	contractKey, err = packKey(contractKey, param.ContractAddr)
	if err != nil {
		return nil, err
	}
	roleF, err := packKey(RoleF, param.Role)
	if err != nil {
		return nil, err
	}
	roleF = append(contractKey, roleF...)
	for _, fn := range param.FuncNames {
		if fn == "" {
			continue
		}
		err = linkedlistInsert(native, roleF, []byte(fn), null)
		if err != nil {
			return nil, err
		}
	}
	// insert into 'func x person' table
	roleP, err := packKey(RoleP, param.Role)
	if err != nil {
		return nil, err
	}
	roleP = append(contractKey, roleP...)
	head, err := linkedlistGetHead(native, roleP)
	if err != nil {
		return nil, err
	}
	q := head ////q is an ont ID
	for {
		if q == nil {
			break
		}
		nodeq, err := linkedlistGetItem(native, roleP, q)
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
	return BYTE_ONE, nil
}

func AssignOntIDsToRole(native *NativeService) ([]byte, error) {
	future := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	auth, err := (&AuthToken{expireTime: uint32(future.Unix()), level: 10}).serialize()
	if err != nil {
		return nil, err
	}
	return assignToRole(native, auth)
}
func assignToRole(native *NativeService, auth []byte) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	//null := []byte{}

	param := new(OntIDsToRoleParam)
	rd := bytes.NewReader(native.Input)
	if err := param.Deserialize(rd); err != nil {
		return nil, err
	}
	if param.Role == nil {
		return nil, errors.NewErr("")
	}

	admin, err := getContractAdmin(native, param.ContractAddr)
	if err != nil {
		return nil, err
	}
	if bytes.Compare(admin, param.AdminOntID) != 0 {
		return nil, errors.NewErr("")
	}
	ret, err := verifySig(native, param.AdminOntID)
	if err != nil {
		return nil, err
	}
	if !ret {
		return BYTE_ONE, nil
	}
	//construct an auth token
	//insert persons into role->person linkedlist
	roleP, err := packKey(RoleP, param.Role)
	if err != nil {
		return nil, err
	}
	contractKey := append(contract[:], param.ContractAddr...)
	roleP = append(contractKey, roleP...)
	for _, p := range param.Persons {
		if p == nil {
			continue
		}
		err = linkedlistInsert(native, roleP, p, auth)
		if err != nil {
			return nil, err
		}
	}
	roleF, err := packKey(RoleF, param.Role)
	if err != nil {
		return nil, err
	}
	roleF = append(contractKey, roleF...)
	head, err := linkedlistGetHead(native, roleF)
	fn := head ////q is a func name
	for {
		if fn == nil {
			break
		}
		nodef, err := linkedlistGetItem(native, roleF, fn) //
		if err != nil {
			return nil, err
		}
		for _, per := range param.Persons {
			err = writeAuthToken(native, param.ContractAddr, fn, per, auth)
		}
		fn = nodef.next
	}
	return BYTE_ONE, nil
}

func Delegate(native *NativeService) ([]byte, error) {
	param := &DelegateParam{}
	rd := bytes.NewReader(native.Input)
	err := param.Deserialize(rd)
	if err != nil {
		return nil, err
	}
	return delegate(native, param.From, param.To, param.ContractAddr, param.Role, param.Period, param.Level)
}

/*
 * if 'from' has the authority and 'to' has not been authorized 'role',
 * then make changes to storage as follows:
 *   - insert 'to' into the linked list of 'role -> [] persons'
 *   - insert 'to' into the delegate list of 'from'.
 *   - put 'func x to' into the 'func x person' table
 */
func delegate(native *NativeService, from []byte, contractAddr []byte, to []byte,
	role []byte, period uint32, level int) ([]byte, error) {
	null := []byte{}
	contract := native.ContextRef.CurrentContext().ContractAddress
	contractKey := append(contract[:], contractAddr...)

	ret, err := verifySig(native, from)
	if err != nil {
		return nil, err
	}
	if !ret {
		return nil, err
	}
	//iterate
	roleP, err := packKey(RoleP, role)
	if err != nil {
		return nil, err
	}
	roleP = append(contractKey, roleP...)

	node_to, err := linkedlistGetItem(native, roleP, to)
	if err != nil {
		return nil, err
	}
	node_from, err := linkedlistGetItem(native, roleP, from)
	if err != nil {
		return nil, err
	}
	if node_to != nil || node_from == nil {
		return nil, err //ignore
	}

	auth := new(AuthToken)
	err = auth.deserialize(node_from.payload)
	if err != nil {
		return nil, err
	}
	//check if 'from' has the permission to delegate
	if auth.level >= 2 && level < auth.level && level > 0 {
		//  put `to` in the delegate list
		//  'role x person' -> [] person is a list of person who gets the auth token
		ser_role_per, err := packKeys(DelegateList, [][]byte{role, from})
		if err != nil {
			return nil, err
		}
		ser_role_per = append(contractKey, ser_role_per...)
		linkedlistInsert(native, ser_role_per, to, null)

		//insert into the `role -> []person` list
		authprime := auth
		authprime.expireTime = native.Time + period
		if authprime.expireTime < native.Time {
			return nil, errors.NewErr("[auth contract] overflow of expire time")
		}
		authprime.level = level
		ser_authprime, err := authprime.serialize()
		if err != nil {
			return nil, err
		}
		persons := &OntIDsToRoleParam{Role: role, Persons: [][]byte{to}}
		bf := new(bytes.Buffer)
		err = persons.Serialize(bf)
		if err != nil {
			return nil, err
		}

		native.Input = bf.Bytes()
		_, err = assignToRole(native, ser_authprime)
		if err != nil {
			return nil, err
		}
	}
	return BYTE_ONE, nil
}

func Withdraw(native *NativeService) ([]byte, error) {
	param := &WithdrawParam{}
	rd := bytes.NewReader(native.Input)
	param.Deserialize(rd)
	return withdraw(native, param.Initiator, param.Delegate, param.ContractAddr, param.Role)
}

func withdraw(native *NativeService, initiator []byte, delegate []byte,
	contractAddr []byte, role []byte) ([]byte, error) {

	contract := native.ContextRef.CurrentContext().ContractAddress
	contractKey := append(contract[:], contractAddr...)

	roleF, err := packKey(RoleF, role)
	if err != nil {
		return nil, err
	}
	roleF = append(contractKey, roleF...)
	fn, err := linkedlistGetHead(native, roleF) //fn is a func name
	funcs := [][]byte{}
	for {
		if fn == nil {
			break
		}
		nodef, err := linkedlistGetItem(native, roleF, fn) //
		if err != nil {
			return nil, err
		}
		funcs = append(funcs, fn)
		fn = nodef.next
	}
	//check if initiator has the right to withdraw
	ser_role_from, err := packKeys(DelegateList, [][]byte{role, initiator})
	if err != nil {
		return nil, err
	}
	ser_role_from = append(contractKey, ser_role_from...)
	node_del, err := linkedlistGetItem(native, ser_role_from, delegate)
	if err != nil {
		return nil, err
	}
	if node_del == nil {
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

		now := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		expireAuth, err := (&AuthToken{expireTime: uint32(now.Unix()), level: 0}).serialize()
		if err != nil {
			return nil, err
		}
		for _, fn := range funcs {
			err = writeAuthToken(native, contractAddr, fn, per, expireAuth)
			if err != nil {
				return nil, err
			}
		}
		// iterate over 'role x person -> [] person' list
		ser_role_per, err := packKeys(DelegateList, [][]byte{role, per})
		ser_role_per = append(contract[:], ser_role_per...)
		q, err := linkedlistGetHead(native, ser_role_per)
		if err != nil {
			return nil, err
		}
		for {
			if q == nil {
				break
			}
			wdList = append(wdList, q)
			item, err := linkedlistGetItem(native, ser_role_per, q)
			if err != nil {
				return nil, err
			}
			_, err = linkedlistDelete(native, ser_role_per, q)
			if err != nil {
				return nil, err
			}
			q = item.next
		}
	}
	return BYTE_ONE, nil
}

/*
 *  VerifyToken(caller []byte, fn []byte, tokenSig []byte) (bool, error)
 *  @caller the ONT ID of the caller
 *  @fn the name of the func to call
 *  @tokenSig the signature on the message
 */

func verifyToken(native *NativeService, caller []byte, contractAddr []byte, fn []byte) (bool, error) {
	//check caller's identity
	ret, err := verifySig(native, caller)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[auth contract] (VerifyToken) verify sig failed")
	}
	if !ret {
		return false, nil
	}

	//check if caller has the auth to call 'fn'
	contract := native.ContextRef.CurrentContext().ContractAddress
	contractKey := append(contract[:], contractAddr...)
	ser_fn_per, err := packKeys(FuncPerson, [][]byte{fn, caller})
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[auth contract] (VerifyToken) pack FuncPerson failed")
	}
	//FIXME
	data, err := native.CloneCache.Get(scommon.ST_STORAGE, append(contractKey, ser_fn_per...))
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[auth contract] (VerifyToken) get storage failed")
	}
	if data == nil {
		return false, nil
	}
	item := data.(*cstates.StorageItem)

	auth := new(AuthToken)
	err = auth.deserialize(item.Value)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[auth contract] (VerifyToken) deserialize auth token failed")
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
		return BYTE_ZERO, err
	}
	ret, err := verifyToken(native, param.Caller, param.ContractAddr, param.Fn)
	if err != nil {
		return BYTE_ZERO, err
	}
	if !ret {
		return BYTE_ZERO, nil
	}
	return BYTE_ONE, nil
}

func verifySig(native *NativeService, ontID []byte) (bool, error) {
	//native.ContextRef.AppCall(genesis., "VerifySignature", []byte{}, args)
	return true, nil
}

func packKeys(field []byte, items [][]byte) ([]byte, error) {
	w := new(bytes.Buffer)
	for _, item := range items {
		err := serialization.WriteVarBytes(w, item)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[AuthContract] packKeys failed")
		}
	}
	key := append(field, w.Bytes()...)
	return key, nil
}

//pack data to be used as a key in the kv storage
// key := field || ser_data
func packKey(field []byte, data []byte) ([]byte, error) {
	return packKeys(field, [][]byte{data})
	//w := new(bytes.Buffer)
	//err := serialization.WriteVarBytes(w, data)
	//if err != nil {
	//	return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[AuthContract] packKey failed")
	//}
	//key := append(field, w.Bytes()...)
	//return key, nil
}

func _InitContractAdmin(native *NativeService) error {
	_, err := InitContractAdmin(native)
	return err
}

func _AssignFuncsToRole(native *NativeService) error {
	_, err := AssignFuncsToRole(native)
	return err
}

func _Delegate(native *NativeService) error {
	_, err := Delegate(native)
	return err
}

func _Withdraw(native *NativeService) error {
	_, err := AssignFuncsToRole(native)
	return err
}
func _AssignOntIDsToRole(native *NativeService) error {
	_, err := AssignOntIDsToRole(native)
	return err
}

func _VerifyToken(native *NativeService) error {
	_, err := VerifyToken(native)
	return err
}
func RegisterAuthContract(native *NativeService) {
	native.Register("initContractAdmin", _InitContractAdmin)
	native.Register("assignFuncsToRole", _AssignFuncsToRole)
	native.Register("delegate", _Delegate)
	native.Register("withdraw", _Withdraw)
	native.Register("assignOntIDsToRole", _AssignOntIDsToRole)
	native.Register("verifyToken", _VerifyToken)
	//native.Register("test", AuthTest)
}

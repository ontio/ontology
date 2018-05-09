package auth

import (
	"bytes"
	"github.com/ontio/ontology/common/serialization"
	//"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"

	. "github.com/ontio/ontology/smartcontract/service/native"
	"io"
	"math"
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
	//native.Contracts[genesis.AuthContractAddress] = RegisterAuthContract
}

//can be called only once
func setContractAdmin(native *NativeService, contractAddr, ontID []byte) (bool, error) {
	//null := []byte{}
	admin, err := get_contract_admin(native, contractAddr)
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

func SetContractAdmin(native *NativeService) ([]byte, error) {
	cxt := native.ContextRef.CallingContext()
	if cxt == nil {
		return nil, errors.NewErr("no calling context")
	}
	invokeAddr := cxt.ContractAddress
	ret, err := setContractAdmin(native, invokeAddr[:], native.Input)
	if err != nil {
		return BYTE_ZERO, err
	}
	if !ret {
		return BYTE_ZERO, nil
	}
	return BYTE_ONE, nil
}

func get_contract_admin(native *NativeService, contractAddr []byte) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	adminKey, err := packKeys(contract[:], [][]byte{contractAddr, Admin})
	if err != nil {
		return nil, err
	}
	val, err := native.CloneCache.Get(scommon.ST_STORAGE, adminKey)
	if err != nil {
		return nil, err
	}
	if val == nil {
		//admin is already set.
		return nil, nil
	}
	admin, ok := val.(*cstates.StorageItem)
	if !ok {
		return nil, errors.NewErr("")
	}
	return admin.Value, nil
}

type TransferParam struct {
	contractAddr  []byte
	newAdminOntID []byte
	sig           []byte
}

func (this *TransferParam) Deserialize(data []byte) error {
	var err error
	rd := bytes.NewReader(data)
	this.contractAddr, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	this.newAdminOntID, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	this.sig, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	return nil
}

func transfer(native *NativeService, contractAddr, newAdminOntID, sig []byte) (bool, error) {
	null := []byte{}
	admin, err := get_contract_admin(native, contractAddr)
	if err != nil {
		return false, err
	}
	msg, err := packKeys(null, [][]byte{contractAddr[:], []byte("Transfer"), newAdminOntID})
	if err != nil {
		return false, err
	}
	ret, err := verifySig(native, admin, msg, sig)
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
	err := param.Deserialize(native.Input)
	if err != nil {
		return nil, err
	}
	ret, err := transfer(native, param.contractAddr, param.newAdminOntID, param.sig)
	if ret {
		return BYTE_ONE, nil
	} else {
		return BYTE_ZERO, nil
	}
}

type PersonsToRole struct {
	role    []byte
	persons [][]byte
}

func (this *PersonsToRole) Serialize() ([]byte, error) {
	bf := new(bytes.Buffer)
	if err := serialization.WriteVarBytes(bf, this.role); err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[PersonsToRole] serialize role error")
	}
	if err := serialization.WriteVarUint(bf, uint64(len(this.persons))); err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[PersonsToRole] serialize persons length error")
	}
	for _, p := range this.persons {
		if err := serialization.WriteVarBytes(bf, p); err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[PersonsToRole] serialize persons error!")
		}
	}
	return bf.Bytes(), nil
}

func (this *PersonsToRole) Deserialize(data []byte) error {
	rd := bytes.NewReader(data)
	role, err := serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	pLen, err := serialization.ReadVarUint(rd, 0)
	if err != nil {
		return err
	}
	var i uint64
	for i = 0; i < pLen; i++ {
		p, err := serialization.ReadVarBytes(rd)
		if err != nil {
			return err
		}
		this.persons = append(this.persons, p)
	}
	this.role = role
	return nil
}

type FuncsToRoleParam struct {
	adminOntID   []byte
	contractAddr []byte
	role         []byte
	funcNames    []string
}

func (this *FuncsToRoleParam) Serialize(bf io.Writer) error {
	if err := serialization.WriteVarBytes(bf, this.adminOntID); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[FuncsToRoleParam] serialize adminOntID error")
	}
	if err := serialization.WriteVarBytes(bf, this.contractAddr); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[FuncsToRoleParam] serialize contractAddr error")
	}
	if err := serialization.WriteVarBytes(bf, this.role); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[FuncsToRoleParam] serialize role error")
	}
	if err := serialization.WriteVarUint(bf, uint64(len(this.funcNames))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[FuncsToRoleParam] serialize funcNames length error")
	}
	for _, fn := range this.funcNames {
		if err := serialization.WriteString(bf, fn); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[FuncsToRoleParam] serialize funcNames error!")
		}
	}
	return nil
}

func (this *FuncsToRoleParam) Deserialize(rd io.Reader) error {
	var err error
	this.adminOntID, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	this.contractAddr, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	this.role, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	fnLen, err := serialization.ReadVarUint(rd, 0)
	if err != nil {
		return err
	}
	this.funcNames = make([]string, fnLen)
	var i uint64
	for i = 0; i < fnLen; i++ {
		fn, err := serialization.ReadString(rd)
		if err != nil {
			return err
		}
		this.funcNames[i] = fn
	}
	return nil
}

type FuncsToRole struct {
	FuncsToRoleParam
	sig []byte
}

func (this *FuncsToRole) Serialize(bf io.Writer) error {
	err := this.FuncsToRoleParam.Serialize(bf)
	if err != nil {
		return err
	}
	err = serialization.WriteVarBytes(bf, this.sig)
	return err
}

func (this *FuncsToRole) Deserialize(rd io.Reader) error {
	err := this.FuncsToRoleParam.Deserialize(rd)
	if err != nil {
		return err
	}
	this.sig, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	return nil
}

type AuthToken struct {
	//ontId []byte
	expireTime uint32
	level      int
}

func (this *AuthToken) Serialize() ([]byte, error) {
	bf := new(bytes.Buffer)
	if err := serialization.WriteVarUint(bf, uint64(this.expireTime)); err != nil {
		return nil, err
	}
	if err := serialization.WriteVarUint(bf, uint64(this.level)); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}
func (this *AuthToken) Deserialize(data []byte) error {
	rd := bytes.NewReader(data)
	expireTime, err := serialization.ReadVarUint(rd, 0)
	if err != nil {
		return err
	}
	level, err := serialization.ReadVarUint(rd, 0)
	if err != nil {
		return err
	}
	this.expireTime = uint32(expireTime)
	this.level = int(level)
	return nil
}

func writeBytes(native *NativeService, key []byte, value []byte) {
	native.CloneCache.Add(scommon.ST_STORAGE, key, &cstates.StorageItem{Value: value})
}

func writeAuthToken(native *NativeService, fn, person, auth []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	key, err := packKeys(FuncPerson, [][]byte{fn, person}) //fn x ontID
	if err != nil {
		return err
	}
	// construct auth token
	writeBytes(native, append(contract[:], key...), auth)
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
	param := new(FuncsToRole)
	rd := bytes.NewReader(native.Input)
	if err := param.Deserialize(rd); err != nil {
		return BYTE_ZERO, err
	}
	if param.role == nil {
		return BYTE_ZERO, errors.NewErr("")
	}

	//check the caller's permission
	admin, err := get_contract_admin(native, param.contractAddr)
	if err != nil {
		return nil, err
	}
	if bytes.Compare(admin, param.adminOntID) != 0 {
		return nil, errors.NewErr("")
	}
	bf := new(bytes.Buffer)
	err = param.FuncsToRoleParam.Serialize(bf)
	if err != nil {
		return nil, err
	}
	msg := bf.Bytes()
	verifySig(native, param.adminOntID, msg, param.sig)

	//insert funcnames into role->func linkedlist
	contractKey := curContract[:]
	contractKey, err = packKey(contractKey, param.contractAddr)
	if err != nil {
		return nil, err
	}
	roleF, err := packKey(RoleF, param.role)
	if err != nil {
		return nil, err
	}
	roleF = append(contractKey, roleF...)
	for _, fn := range param.funcNames {
		if fn == "" {
			continue
		}
		err = linkedlistInsert(native, roleF, []byte(fn), null)
		if err != nil {
			return nil, err
		}
	}
	// insert into 'func x person' table
	roleP, err := packKey(RoleP, param.role)
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
		for _, fn := range param.funcNames {
			err = writeAuthToken(native, []byte(fn), q, nodeq.payload)
			if err != nil {
				return nil, err
			}
		}
		q = nodeq.next
	}
	return BYTE_ONE, nil
}

func assign_ontids_to_role(native *NativeService) error {
	future := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	auth, err := (&AuthToken{expireTime: uint32(future.Unix()), level: 10}).Serialize()
	if err != nil {
		return err
	}
	return assign_to_role(native, auth)
}
func assign_to_role(native *NativeService, auth []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	//null := []byte{}

	persons := new(PersonsToRole)
	if err := persons.Deserialize(native.Input); err != nil {
		return err
	}
	if persons.role == nil {
		return errors.NewErr("")
	}
	//construct an auth token
	//insert persons into role->person linkedlist
	roleP, err := packKey(RoleP, persons.role)
	if err != nil {
		return err
	}
	roleP = append(contract[:], roleP...)
	for _, p := range persons.persons {
		if p == nil {
			continue
		}
		err = linkedlistInsert(native, roleP, p, auth)
		if err != nil {
			return err
		}
	}
	roleF, err := packKey(RoleF, persons.role)
	if err != nil {
		return err
	}
	roleF = append(contract[:], roleF...)
	head, err := linkedlistGetHead(native, roleF)
	fn := head ////q is a func name
	for {
		if fn == nil {
			break
		}
		nodef, err := linkedlistGetItem(native, roleF, fn) //
		if err != nil {
			return err
		}
		for _, per := range persons.persons {
			err = writeAuthToken(native, fn, per, auth)
		}
		fn = nodef.next
	}
	return nil
}

type DelegateParam struct {
	from   []byte
	to     []byte
	role   []byte
	period uint32
	level  int
}

func (this *DelegateParam) Deserialize(data []byte) error {
	var err error
	rd := bytes.NewReader(data)
	this.from, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	this.to, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	this.role, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	period, err := serialization.ReadVarUint(rd, 0)
	if err != nil {
		return err
	}
	level, err := serialization.ReadVarUint(rd, 0)
	if err != nil {
		return err
	}
	if period > math.MaxUint32 || level > math.MaxInt8 {
		return errors.NewErr("[auth contract] delegate deserialize: period or level too large")
	}
	this.period = uint32(period)
	this.level = int(level)
	return nil
}
func delegate(native *NativeService) error {
	param := &DelegateParam{}
	err := param.Deserialize(native.Input)
	if err != nil {
		return err
	}
	return _delegate(native, param.from, param.to, param.role, param.period, param.level)
}

/*
 * if 'from' has the authority and 'to' has not been authorized 'role',
 * then make changes to storage as follows:
 *   - insert 'to' into the linked list of 'role -> [] persons'
 *   - insert 'to' into the delegate list of 'from'.
 *   - put 'func x to' into the 'func x person' table
 */
func _delegate(native *NativeService, from []byte, to []byte, role []byte, period uint32, level int) error {
	null := []byte{}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//iterate
	roleP, err := packKey(RoleP, role)
	if err != nil {
		return err
	}
	roleP = append(contract[:], roleP...)

	node_to, err := linkedlistGetItem(native, roleP, to)
	if err != nil {
		return err
	}
	node_from, err := linkedlistGetItem(native, roleP, from)
	if err != nil {
		return err
	}
	if node_to != nil || node_from == nil {
		return nil //ignore
	}

	auth := new(AuthToken)
	err = auth.Deserialize(node_from.payload)
	if err != nil {
		return err
	}
	//check if 'from' has the permission to delegate
	if auth.level >= 2 && level < auth.level && level > 0 {
		//  put `to` in the delegate list
		//  'role x person' -> [] person is a list of person who gets the auth token
		ser_role_per, err := packKeys(DelegateList, [][]byte{role, from})
		if err != nil {
			return err
		}
		ser_role_per = append(contract[:], ser_role_per...)
		linkedlistInsert(native, ser_role_per, to, null)

		//insert into the `role -> []person` list
		authprime := auth
		authprime.expireTime = native.Time + period
		if authprime.expireTime < native.Time {
			return errors.NewErr("[auth contract] overflow of expire time")
		}
		authprime.level = level
		ser_authprime, err := authprime.Serialize()
		if err != nil {
			return err
		}
		persons := &PersonsToRole{role: role, persons: [][]byte{to}}
		ser_per, err := persons.Serialize()
		if err != nil {
			return err
		}
		native.Input = ser_per
		err = assign_to_role(native, ser_authprime)
		if err != nil {
			return err
		}
	}
	return nil
}

type WithdrawParam struct {
	initiator []byte
	delegate  []byte
	role      []byte
}

func (this *WithdrawParam) Deserialize(data []byte) error {
	var err error
	rd := bytes.NewReader(data)
	this.initiator, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	this.delegate, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	this.role, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	return nil
}

func withdraw(native *NativeService) error {
	param := &WithdrawParam{}
	param.Deserialize(native.Input)
	return _withdraw(native, param.initiator, param.delegate, param.role)
}

func _withdraw(native *NativeService, initiator []byte, delegate []byte, role []byte) error {
	//null := []byte{}
	contract := native.ContextRef.CurrentContext().ContractAddress

	roleF, err := packKey(RoleF, role)
	if err != nil {
		return err
	}
	roleF = append(contract[:], roleF...)
	fn, err := linkedlistGetHead(native, roleF) //fn is a func name
	funcs := [][]byte{}
	for {
		if fn == nil {
			break
		}
		nodef, err := linkedlistGetItem(native, roleF, fn) //
		if err != nil {
			return err
		}
		funcs = append(funcs, fn)
		fn = nodef.next
	}
	//check if initiator has the right to withdraw
	ser_role_from, err := packKeys(DelegateList, [][]byte{role, initiator})
	if err != nil {
		return err
	}
	ser_role_from = append(contract[:], ser_role_from...)
	node_del, err := linkedlistGetItem(native, ser_role_from, delegate)
	if err != nil {
		return err
	}
	if node_del == nil {
		return errors.NewErr("[auth contract] initiator does not have the right")
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
		expireAuth, err := (&AuthToken{expireTime: uint32(now.Unix()), level: 0}).Serialize()
		if err != nil {
			return err
		}
		for _, fn := range funcs {
			err = writeAuthToken(native, fn, per, expireAuth)
			if err != nil {
				return err
			}
		}
		// iterate over 'role x person -> [] person' list
		ser_role_per, err := packKeys(DelegateList, [][]byte{role, per})
		ser_role_per = append(contract[:], ser_role_per...)
		q, err := linkedlistGetHead(native, ser_role_per)
		if err != nil {
			return err
		}
		for {
			if q == nil {
				break
			}
			wdList = append(wdList, q)
			item, err := linkedlistGetItem(native, ser_role_per, q)
			if err != nil {
				return err
			}
			_, err = linkedlistDelete(native, ser_role_per, q)
			if err != nil {
				return err
			}
			q = item.next
		}
	}
	return nil
}

/*
 *  VerifyToken(caller []byte, fn []byte, tokenSig []byte) (bool, error)
 *  @caller the ONT ID of the caller
 *  @fn the name of the func to call
 *  @tokenSig the signature on the message
 */
type VerifyTokenParam struct {
	caller []byte
	//contractAddr []byte
	fn  []byte
	sig []byte
}

func (this *VerifyTokenParam) Deserialize(data []byte) error {
	var err error
	rd := bytes.NewReader(data)
	this.caller, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err //deserialize caller error
	}
	this.fn, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	this.sig, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	return nil
}

func _verify_token(native *NativeService, caller []byte, fn []byte, tokenSig []byte) (bool, error) {
	// function being called is invokeAddr.fn
	null := []byte{}
	invokeAddr := native.ContextRef.CallingContext().ContractAddress
	contractFunc, err := packKeys(null, [][]byte{invokeAddr[:], fn})
	if err != nil {
		return false, err
	}
	//check caller's identity
	ret, err := verifySig(native, caller, contractFunc, tokenSig)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[auth contract] (VerifyToken) verify sig failed")
	}
	if !ret {
		return false, nil
	}

	//check if caller has the auth to call 'fn'
	contract := native.ContextRef.CurrentContext().ContractAddress
	ser_fn_per, err := packKeys(FuncPerson, [][]byte{contractFunc, caller})
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[auth contract] (VerifyToken) pack FuncPerson failed")
	}
	//FIXME
	data, err := native.CloneCache.Get(scommon.ST_STORAGE, append(contract[:], ser_fn_per...))
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[auth contract] (VerifyToken) get storage failed")
	}
	if data == nil {
		return false, nil
	}
	item := data.(*cstates.StorageItem)

	auth := new(AuthToken)
	err = auth.Deserialize(item.Value)
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
	err := param.Deserialize(native.Input)
	if err != nil {
		return BYTE_ZERO, err
	}
	ret, err := _verify_token(native, param.caller, param.fn, param.sig)
	if err != nil {
		return BYTE_ZERO, err
	}
	if !ret {
		return BYTE_ZERO, nil
	}
	return BYTE_ONE, nil
}

func verifySig(native *NativeService, ontID []byte, msg []byte, sig []byte) (bool, error) {

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

func RegisterAuthContract(native *NativeService) {
	native.Register("assignFuncsToRole", AssignFuncsToRole)
	native.Register("delegate", delegate)
	native.Register("withdraw", withdraw)
	native.Register("assignOntIDsToRole", assign_ontids_to_role)
	//native.Register("verifyToken", VerifyToken)
	//native.Register("test", AuthTest)
}

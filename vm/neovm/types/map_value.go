package types

type MapValue struct {
	Data map[string]VmValue
}

func NewMapValue() *MapValue {
	return &MapValue{Data: make(map[string]VmValue)}
}
func (this *MapValue) Set(key, value VmValue) error {
	skey, err := key.GetMapKey()
	if err != nil {
		return err
	}

	this.Data[skey] = value
	return nil
}

func (this *MapValue) Reset() {
	this.Data = make(map[string]VmValue)
}

func (this *MapValue) Remove(key VmValue) error {
	skey, err := key.GetMapKey()
	if err != nil {
		return err
	}

	delete(this.Data, skey)

	return nil
}

func (this *MapValue) Get(key VmValue) (value VmValue, ok bool, err error) {
	skey, e := key.GetMapKey()
	if e != nil {
		err = e
		return
	}

	value, ok = this.Data[skey]
	return
}

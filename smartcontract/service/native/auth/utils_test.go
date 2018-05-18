package auth

import "testing"

//{"a", "b"} == {"b", "a"}
func testEq(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}
	Map := make(map[string]bool)
	for i := range a {
		Map[a[i]] = true
	}
	for _, s := range b {
		_, ok := Map[s]
		if !ok {
			return false
		}
	}
	return true
}
func TestStringSliceUniq(t *testing.T) {
	s := []string{"foo", "foo1", "foo2", "foo", "foo1", "foo2", "foo3"}
	ret := stringSliceUniq(s)
	t.Log(ret)
	if !testEq(ret, []string{"foo", "foo1", "foo2", "foo3"}) {
		t.Fatalf("failed")
	}
}

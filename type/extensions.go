package typex

import "fmt"

type NullableBool struct {
	Value *bool
}

func (nb *NullableBool) Set(s string) error {
	v := (s == "true")
	nb.Value = &v
	return nil
}

func (nb *NullableBool) String() string {
	if nb.Value == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", *nb.Value)
}

func (nb *NullableBool) Val(defaultValue bool) bool {
	if nb.Value == nil {
		return defaultValue
	}
	return *nb.Value
}

func (nb *NullableBool) IsBoolFlag() bool {
	return true
}

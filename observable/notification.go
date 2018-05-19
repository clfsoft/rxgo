package observable

// Notification is the representation of an emission.
type Notification struct {
	Value    interface{}
	HasValue bool
	HasError bool
}

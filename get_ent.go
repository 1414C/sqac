package sqac

// GetEnt interface is provided for retrieval of slices
// containing arbitrary structs.  Exec accepts a handle
// to a PublicDB interface, thereby providing access to
// the DB sqac is connected to.  it is up to the caller
// to code the underlying value / code.  As awful as this
// seems, it eliminates the need to abuse reflection or
// unsafe pointers to achieve the goal of a general
// CRUD API.
type GetEnt interface {
	Exec(handle PublicDB) error
}

package common

// GetParam defines a common structure for
// CRUD GET parameters.
type GetParam struct {
	FieldName    string
	Operand      string
	ParamValue   interface{}
	NextOperator string
}

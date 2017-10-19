package sqac_test

import "testing"
import "github.com/1414C/sqac/common"
import "fmt"

// TestCamelToSnake
//
// Check that function CamelToSnake is correct
func TestCamelToSnake(t *testing.T) {

	var errs []error
	type pairs struct {
		camel string
		snake string
	}

	var p pairs
	tList := make([]pairs, 0)
	p.camel = "material"
	p.snake = "material"
	tList = append(tList, p)
	p.camel = "materialNum"
	p.snake = "material_num"
	tList = append(tList, p)
	p.camel = "testCamelCaseIBMPowerEdge"
	p.snake = "test_camel_case_ibm_power_edge"
	tList = append(tList, p)
	p.camel = "IBMOneTwo"
	p.snake = "ibm_one_two"
	tList = append(tList, p)
	p.camel = "IDOneTwo"
	p.snake = "id_one_two"
	tList = append(tList, p)
	p.camel = "IOneTwo"
	p.snake = "i_one_two"
	tList = append(tList, p)

	for _, v := range tList {
		res := common.CamelToSnake(v.camel)
		if res != v.snake {
			errs = append(errs, fmt.Errorf("CamelToSnake('%s' expected '%s' - got '%s'", v.camel, v.snake, res))
		}
	}

	if len(errs) > 0 {
		es := ""
		for _, e := range errs {
			es = es + e.Error() + "\n"
		}
		t.Errorf(es)
	}
}

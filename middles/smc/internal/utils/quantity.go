package utils

import (
	"fmt"
	"strconv"
	"strings"
)

type Quantity struct {
	Amend int64  // Integer value
	Unit  string // Unit
}

/**
 * Scaling parameters for measurement units
 */
type Scale struct {
	Unit     string
	Multiple int64
}

/**
 * String representation, e.g: 100m,1K,1000
 */
func (q *Quantity) String() string {
	return fmt.Sprint(q.Amend) + q.Unit
}

/**
 * Kubernetes-style representation, e.g: 1Ki, 1Mi
 */
func (q *Quantity) K8sString() string {
	if q.Unit == "" {
		return fmt.Sprint(q.Amend)
	}
	if strings.ContainsRune("KMGTPE", rune(q.Unit[0])) {
		return fmt.Sprint(q.Amend) + q.Unit + "i"
	}
	return fmt.Sprint(q.Amend) + q.Unit
}

/**
 * Parse resource quantity from string representation
 */
func (q *Quantity) Parse(num string) error {
	if num == "" {
		q.Amend = 0
		q.Unit = ""
		return nil
	}
	if len(num) <= 1 {
		q.Amend, _ = strconv.ParseInt(num, 10, 64)
		q.Unit = ""
		return nil
	}
	uch := num[len(num)-1]
	digit := num[:len(num)-1]
	if !strings.ContainsRune("mKMGTPE", rune(uch)) {
		if !strings.ContainsRune("0123456789", rune(uch)) {
			return fmt.Errorf("invalid format")
		}
		digit = num
		uch = 0
	}
	size, err := strconv.ParseInt(digit, 10, 64)
	if err != nil {
		return err
	}
	q.Amend = size
	if uch == 0 {
		q.Unit = ""
	} else {
		q.Unit = string(uch)
	}
	return nil
}

/**
 * Increase resource quantity
 */
func (q *Quantity) Plus(rhs Quantity) error {
	if err := alignUnit(q, &rhs); err != nil {
		return err
	}
	q.Amend += rhs.Amend
	return nil
}

/**
 * Decrease resource quantity
 */
func (q *Quantity) Minus(rhs Quantity) error {
	if err := alignUnit(q, &rhs); err != nil {
		return err
	}
	q.Amend -= rhs.Amend
	return nil
}

/**
 * Scaling ladder table
 * Each level to next level (larger units) requires division by current level factor
 * Each level to previous level (smaller units) requires multiplication by previous level factor
 */
var scales []Scale = []Scale{
	{"m", 1000},
	{"", 1024},
	{"K", 1024},
	{"M", 1024},
	{"G", 1024},
	{"T", 1024},
	{"P", 1024},
	{"E", 1024},
}

/**
 * Optimize for better display
 */
func (q *Quantity) Optimize() Quantity {
	if q.Amend == 0 {
		return *q
	}
	val := q.Amend
	ok := false
	for _, scale := range scales {
		if scale.Unit == q.Unit {
			ok = true
		}
		if !ok {
			continue
		}
		if val%scale.Multiple != 0 {
			q.Amend = val
			q.Unit = scale.Unit
			return *q
		}
		val = val / scale.Multiple
	}
	return *q
}

/**
 * Change display unit of resource quantity
 */
func (q *Quantity) ChangeUnit(unit string) error {
	s := -1 // Source level
	e := -1 // Target level
	for i, scale := range scales {
		if scale.Unit == unit {
			e = i
		}
		if scale.Unit == q.Unit {
			s = i
		}
	}
	if s == -1 || e == -1 {
		return fmt.Errorf("invalid format")
	}
	var val int64 = q.Amend
	if s == e {
		return nil
	} else if s > e { // Larger unit to smaller unit
		for i := s - 1; i >= e; i-- {
			val = val * scales[i].Multiple
		}
	} else { // Smaller unit to larger unit
		for i := s; i < e; i++ {
			if val%scales[i].Multiple == 0 {
				val = val / scales[i].Multiple
			} else {
				return fmt.Errorf("failed")
			}
		}
	}
	q.Amend = val
	q.Unit = unit
	return nil
}

/**
 * Create new resource quantity object
 */
func NewQuantity(amend int64, unit string) (Quantity, error) {
	for _, s := range scales {
		if unit == s.Unit {
			return Quantity{Amend: amend, Unit: unit}, nil
		}
	}
	return Quantity{}, fmt.Errorf("invalid quantity: %d%s", amend, unit)
}

/**
 * Parse resource quantity from string
 */
func QuantityParse(qstr string) (Quantity, error) {
	var qn Quantity
	if err := qn.Parse(qstr); err != nil {
		return Quantity{}, err
	}
	return qn, nil
}

/**
 * Add two resource quantities
 */
func QuantityPlus(lhs, rhs Quantity) (Quantity, error) {
	if err := lhs.Plus(rhs); err != nil {
		return Quantity{}, err
	}
	return lhs, nil
}

/**
 * Subtract two resource quantities
 */
func QuantityMinus(lhs, rhs Quantity) (Quantity, error) {
	if err := lhs.Minus(rhs); err != nil {
		return Quantity{}, err
	}
	return lhs, nil
}

/**
 * Compare two resource quantities
 */
func QuantityCompare(lhs, rhs Quantity) (int, error) {
	if err := alignUnit(&lhs, &rhs); err != nil {
		return -1, err
	}
	val := lhs.Amend - rhs.Amend
	if val < 0 {
		return -1, nil
	} else if val > 0 {
		return 1, nil
	} else {
		return 0, nil
	}
}

/**
 * Align units for operations between two resource quantities
 */
func alignUnit(lhs, rhs *Quantity) error {
	l := -1
	r := -1
	for i, scale := range scales {
		if scale.Unit == lhs.Unit {
			l = i
		}
		if scale.Unit == rhs.Unit {
			r = i
		}
	}
	if l == -1 || r == -1 {
		return fmt.Errorf("invalid format")
	}
	if l == r {
		return nil
	} else if l < r { // Left side has smaller units (mKMGTPE)
		return rhs.ChangeUnit(lhs.Unit)
	} else { // Right side has smaller units
		return lhs.ChangeUnit(rhs.Unit)
	}
}

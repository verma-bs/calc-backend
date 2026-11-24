package calculator

import (
	"errors"
	"math"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name string
		op   Operation
		a, b float64
		want float64
		err  error
	}{
		{"addition", Add, -2, 5, 3, nil},
		{"subtraction order", Subtract, 2, 5, -3, nil},
		{"multiplication", Multiply, -4, 2.5, -10, nil},
		{"fractional division", Divide, 3, 2, 1.5, nil},
		{"power", Power, 2, 3, 8, nil},
		{"fractional power", Power, 9, 0.5, 3, nil},
		{"zero numerator", Divide, 0, 4, 0, nil},
		{"zero denominator", Divide, 4, 0, 0, ErrDivisionByZero},
		{"negative zero denominator", Divide, 4, math.Copysign(0, -1), 0, ErrDivisionByZero},
		{"unknown operation", "modulo", 2, 3, 0, ErrUnsupportedOperation},
		{"oversized operand", Multiply, math.MaxFloat64, 2, 0, ErrOperandOutOfRange},
		{"infinite operand", Add, math.Inf(1), 2, 0, ErrNonFiniteOperand},
		{"NaN operand", Add, 1, math.NaN(), 0, ErrNonFiniteOperand},
		{"operand above exact integer range", Add, 9007199254740992, 1, 0, ErrOperandOutOfRange},
		{"result above exact integer range", Add, 9007199254740991, 1, 0, ErrNonFiniteResult},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calculate(tt.op, tt.a, tt.b)
			if !errors.Is(err, tt.err) || got != tt.want {
				t.Fatalf("got (%v, %v), want (%v, %v)", got, err, tt.want, tt.err)
			}
		})
	}
}

func TestCalculateInfix(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       float64
		err        error
	}{
		{"addition chain", "1+4+3", 8, nil},
		{"decimal addition precision", "27.02+20", 47.02, nil},
		{"large addition preserves small increment", "2700000000000000+20", 2700000000000020, nil},
		{"input above exact integer range", "270000000000000000000+20", 0, ErrOperandOutOfRange},
		{"result above exact integer range", "9007199254740991+1", 0, ErrNonFiniteResult},
		{"trailing operator", "1+4+3-", 0, ErrInvalidExpression},
		{"consecutive signs", "1+2*3-+2^2", 0, ErrInvalidExpression},
		{"consecutive signs with spaces", "1- +2", 0, ErrInvalidExpression},
		{"mixed precedence", "1+2*3-2^2", 3, nil},
		{"right associative power", "2^3^2", 512, nil},
		{"negative exponent", "2^-2", 0.25, nil},
		{"unary minus", "-2^2", -4, nil},
		{"parentheses", "(1+2)*3", 9, nil},
		{"decimal operands", "7/2", 3.5, nil},
		{"scientific notation", "1e-3*2", 0.002, nil},
		{"division by zero", "1/0", 0, ErrDivisionByZero},
		{"missing operand", "1+", 0, ErrInvalidExpression},
		{"invalid token", "1+x", 0, ErrInvalidExpression},
		{"unclosed parenthesis", "(1+2", 0, ErrInvalidExpression},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateInfix(tt.expression)
			if !errors.Is(err, tt.err) || math.Abs(got-tt.want) > 1e-12 {
				t.Fatalf("got (%v, %v), want (%v, %v)", got, err, tt.want, tt.err)
			}
		})
	}
}

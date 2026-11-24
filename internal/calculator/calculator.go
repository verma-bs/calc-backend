package calculator

import (
	"errors"
	"math"
	"strconv"
	"unicode"
)

type Operation string

const (
	Add      Operation = "add"
	Subtract Operation = "subtract"
	Multiply Operation = "multiply"
	Divide   Operation = "divide"
	Power    Operation = "power"
)

var (
	ErrUnsupportedOperation = errors.New("unsupported operation")
	ErrDivisionByZero       = errors.New("division by zero")
	ErrNonFiniteOperand     = errors.New("operands must be finite numbers")
	ErrOperandOutOfRange    = errors.New("number exceeds the supported range (maximum absolute value is 9007199254740991)")
	ErrNonFiniteResult      = errors.New("result exceeds the supported numeric range")
	ErrInvalidExpression    = errors.New("invalid expression")
)

const resultPrecision = 12
const maxExactInteger = 9007199254740991.0

func Calculate(op Operation, a, b float64) (float64, error) {
	if !finite(a) || !finite(b) {
		return 0, ErrNonFiniteOperand
	}
	if math.Abs(a) > maxExactInteger || math.Abs(b) > maxExactInteger {
		return 0, ErrOperandOutOfRange
	}

	var result float64
	switch op {
	case Add:
		result = a + b
	case Subtract:
		result = a - b
	case Multiply:
		result = a * b
	case Divide:
		if b == 0 {
			return 0, ErrDivisionByZero
		}
		result = a / b
	case Power:
		result = math.Pow(a, b)
	default:
		return 0, ErrUnsupportedOperation
	}

	if !finite(result) || math.Abs(result) > maxExactInteger {
		return 0, ErrNonFiniteResult
	}
	return result, nil
}

func CalculateInfix(expression string) (float64, error) {
	input := []rune(expression)
	if hasConsecutiveSigns(input) {
		return 0, ErrInvalidExpression
	}
	parser := expressionParser{input: input}
	result, err := parser.parseAddSubtract()
	parser.skipSpaces()
	if err != nil {
		return 0, err
	}
	if parser.position != len(parser.input) {
		return 0, ErrInvalidExpression
	}
	rounded, err := strconv.ParseFloat(strconv.FormatFloat(result, 'g', resultPrecision, 64), 64)
	if err != nil {
		return 0, ErrNonFiniteResult
	}
	ulp := math.Nextafter(math.Abs(result), math.Inf(1)) - math.Abs(result)
	if math.Abs(result-rounded) <= ulp {
		return rounded, nil
	}
	return result, nil
}

func hasConsecutiveSigns(input []rune) bool {
	previousSign := false
	for _, character := range input {
		if unicode.IsSpace(character) {
			continue
		}
		currentSign := character == '+' || character == '-'
		if previousSign && currentSign {
			return true
		}
		previousSign = currentSign
	}
	return false
}

type expressionParser struct {
	input    []rune
	position int
}

func (p *expressionParser) parseAddSubtract() (float64, error) {
	left, err := p.parseMultiplyDivide()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpaces()
		if !p.consume('+') && !p.consume('-') {
			return left, nil
		}
		operator := p.input[p.position-1]
		right, err := p.parseMultiplyDivide()
		if err != nil {
			return 0, err
		}
		op := Add
		if operator == '-' {
			op = Subtract
		}
		left, err = Calculate(op, left, right)
		if err != nil {
			return 0, err
		}
	}
}

func (p *expressionParser) parseMultiplyDivide() (float64, error) {
	left, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpaces()
		if !p.consume('*') && !p.consume('/') {
			return left, nil
		}
		operator := p.input[p.position-1]
		right, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		op := Multiply
		if operator == '/' {
			op = Divide
		}
		left, err = Calculate(op, left, right)
		if err != nil {
			return 0, err
		}
	}
}

func (p *expressionParser) parseUnary() (float64, error) {
	p.skipSpaces()
	if p.consume('+') {
		return p.parseUnary()
	}
	if p.consume('-') {
		value, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		return -value, nil
	}
	return p.parsePower()
}

func (p *expressionParser) parsePower() (float64, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return 0, err
	}
	p.skipSpaces()
	if !p.consume('^') {
		return left, nil
	}
	right, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	return Calculate(Power, left, right)
}

func (p *expressionParser) parsePrimary() (float64, error) {
	p.skipSpaces()
	if p.consume('(') {
		value, err := p.parseAddSubtract()
		if err != nil {
			return 0, err
		}
		p.skipSpaces()
		if !p.consume(')') {
			return 0, ErrInvalidExpression
		}
		return value, nil
	}

	start := p.position
	for p.position < len(p.input) && (unicode.IsDigit(p.input[p.position]) || p.input[p.position] == '.') {
		p.position++
	}
	if p.position < len(p.input) && (p.input[p.position] == 'e' || p.input[p.position] == 'E') {
		p.position++
		if p.position < len(p.input) && (p.input[p.position] == '+' || p.input[p.position] == '-') {
			p.position++
		}
		exponentStart := p.position
		for p.position < len(p.input) && unicode.IsDigit(p.input[p.position]) {
			p.position++
		}
		if exponentStart == p.position {
			return 0, ErrInvalidExpression
		}
	}
	if start == p.position {
		return 0, ErrInvalidExpression
	}
	value, err := strconv.ParseFloat(string(p.input[start:p.position]), 64)
	if err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
		return 0, ErrNonFiniteOperand
	}
	if math.Abs(value) > maxExactInteger {
		return 0, ErrOperandOutOfRange
	}
	return value, nil
}

func (p *expressionParser) skipSpaces() {
	for p.position < len(p.input) && unicode.IsSpace(p.input[p.position]) {
		p.position++
	}
}

func (p *expressionParser) consume(character rune) bool {
	if p.position < len(p.input) && p.input[p.position] == character {
		p.position++
		return true
	}
	return false
}

func finite(n float64) bool {
	return !math.IsNaN(n) && !math.IsInf(n, 0)
}

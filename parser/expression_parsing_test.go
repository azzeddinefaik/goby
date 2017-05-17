package parser

import (
	"github.com/goby-lang/goby/ast"
	"github.com/goby-lang/goby/lexer"
	"testing"
)

func TestNilExpression(t *testing.T) {
	input := `
	nil
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	_, ok := stmt.Expression.(*ast.NilExpression)

	if !ok {
		t.Fatalf("Expect expression to be NilExpression. got=%T", stmt.Expression)
	}
}

func TestNamespaceConstant(t *testing.T) {
	input := `
	Foo::Bar
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	exp := stmt.Expression.(*ast.InfixExpression)

	if exp.Operator != "::" {
		t.Fatalf("Expect infix operator to be resolution operator. got=%s", exp.Operator)
	}

	namespace, ok := exp.Left.(*ast.Constant)

	if !ok {
		t.Fatalf("Expect left side of resolution operator is a constant. got=%T", exp.Left)

		if namespace.Value != "Foo" {
			t.Fatalf("Expect namespace to be %s. got=%s", "Foo", namespace.Value)
		}
	}

	constant, ok := exp.Right.(*ast.Constant)

	if !ok {
		t.Fatalf("Expect right side of resolution operator is a constant. got=%T", exp.Right)

		if namespace.Value != "Bar" {
			t.Fatalf("Expect namespace to be %s. got=%s", "Bar", constant.Value)
		}
	}
}

func TestHashExpression(t *testing.T) {
	tests := []struct {
		input            string
		expectedElements map[string]int
	}{
		{`{}`, map[string]int{}},
		{
			`{ test: 123 }`,
			map[string]int{
				"test": 123,
			},
		},
		{

			`{ another_string: 456 }`,
			map[string]int{
				"another_string": 456,
			},
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program has not enough statments. expect 1, got=%d", len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)

		if !ok {
			t.Fatalf("program.Statments[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		hash, ok := stmt.Expression.(*ast.HashExpression)

		for key := range hash.Data {
			testIntegerLiteral(t, hash.Data[key], tt.expectedElements[key])
		}
	}
}

func TestHashAccessExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{}["123"]`, "123"},
		{`{ test: 1 }["test"]`, "test"},
		{`{ foo: "123" }["foo"]`, "foo"},
		{`{ bar: true }["bar"]`, "bar"},
		{`{ bar: true }[var]`, "var"},
	}

	for i, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program has not enough statments. expect 1, got=%d", len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)

		if !ok {
			t.Fatalf("program.Statments[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		hashAccess, ok := stmt.Expression.(*ast.CallExpression)

		_, ok = hashAccess.Receiver.(*ast.HashExpression)

		if !ok {
			t.Fatalf("expect method call's receiver to be a hash. got=%T", hashAccess.Receiver)
		}

		if i < 4 {
			testStringLiteral(t, hashAccess.Arguments[0], tt.expected)
		} else {
			testIdentifier(t, hashAccess.Arguments[0], tt.expected)
		}

	}
}

func TestArrayExpression(t *testing.T) {
	tests := []struct {
		input            string
		expectedElements []int
	}{
		{`[]`, []int{}},
		{`[1]`, []int{1}},
		{`[1,2,4,5]`, []int{1, 2, 4, 5}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program has not enough statments. expect 1, got=%d", len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)

		if !ok {
			t.Fatalf("program.Statments[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		arr, ok := stmt.Expression.(*ast.ArrayExpression)

		for i, elem := range arr.Elements {
			testIntegerLiteral(t, elem, tt.expectedElements[i])
		}
	}
}

func TestArrayIndexExpression(t *testing.T) {
	tests := []struct {
		input         string
		expectedIndex interface{}
	}{
		{`[][1]`, 1},
		{`[1][0]`, 0},
		{`[1,2,4,5][3]`, 3},
		{`[1,2,4,5][foo]`, "foo"},
		{`test[bar]`, "bar"},
		{`test[1]`, 1},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program has not enough statments. expect 1, got=%d", len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)

		if !ok {
			t.Fatalf("program.Statments[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		arrIndex, ok := stmt.Expression.(*ast.CallExpression)

		switch expected := tt.expectedIndex.(type) {
		case int:
			testIntegerLiteral(t, arrIndex.Arguments[0], expected)
		case string:
			testIdentifier(t, arrIndex.Arguments[0], expected)
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := `foobar;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statments. expect 1, got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)

	if !ok {
		t.Fatalf("program.Statments[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
	testIdentifier(t, ident, "foobar")

}

func TestConstantExpression(t *testing.T) {
	input := `Person;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statments. expect 1, got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)

	if !ok {
		t.Fatalf("program.Statments[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	constant, ok := stmt.Expression.(*ast.Constant)
	testConstant(t, constant, "Person")

}

func TestIntegerLiteralExpression(t *testing.T) {
	input := `5;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("first program statement is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	testIntegerLiteral(t, literal, 5)
}

func TestStringLiteralExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: `"testString";`, expected: "testString"},
		{input: `'test_string';`, expected: "test_string"},
		{input: `'!@#!@!$123';`, expected: "!@#!@!$123"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("first program statement is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		literal, ok := stmt.Expression.(*ast.StringLiteral)
		testStringLiteral(t, literal, tt.expected)
	}
}

func TestParsingPrefixExpression(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		expected interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!true;", "!", true},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("expect %d statements. got=%d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("statement is not ast.Expression. got=%T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("expression is not a PrefixExpression. got=%T", stmt.Expression)
		}
		if exp.Operator != tt.operator {
			t.Fatalf("expression's operator is not '-'. got=%s", exp.Operator)
		}
		if !testLiteralExpression(t, exp.Right, tt.expected) {
			return
		}
	}
}

func TestParsingPostfixExpression(t *testing.T) {
	tests := []struct {
		input            string
		operator         string
		expectedReceiver interface{}
	}{
		{"1++", "++", 1},
		{"2--", "--", 2},
		{"foo++", "++", "foo"},
		{"bar--", "--", "bar"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("expect %d statements. got=%d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("statement is not ast.Expression. got=%T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.CallExpression)
		if !ok {
			t.Fatalf("expression is not a CallExpression. got=%T", stmt.Expression)
		}
		if exp.Method != tt.operator {
			t.Fatalf("expression's operator is not %s. got=%s", tt.operator, exp.Method)
		}
		switch r := tt.expectedReceiver.(type) {
		case int:
			testIntegerLiteral(t, exp.Receiver, r)
		case string:
			testIdentifier(t, exp.Receiver, r)
		}
	}
}

func TestParsingInfixExpression(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  int
		operator   string
		rightValue int
	}{
		{"4 + 1;", 4, "+", 1},
		{"3 - 2;", 3, "-", 2},
	}

	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("expect %d statements. got=%d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statments[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}
		testInfixExpression(t, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue)
	}
}

func TestIfExpression(t *testing.T) {
	input := `
	if x < y
	  x + 5
	else
	  y + 4
	end
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expect program's statements to be 1. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)

	if !ok {
		t.Fatalf("expect program.Statements[0] to be *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)

	if !ok {
		t.Fatalf("expect statement to be an IfExpression. got=%T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("should be only one consequence. got=%d\n", len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)

	if !ok {
		t.Errorf("expect consequence should be an expression statement. got=%T", exp.Consequence.Statements[0])
	}

	if !testInfixExpression(t, consequence.Expression, "x", "+", 5) {
		return
	}

	alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)

	if !ok {
		t.Errorf("expect alternative should be an expression statement. got=%T", exp.Alternative.Statements[0])
	}

	if !testInfixExpression(t, alternative.Expression, "y", "+", 4) {
		return
	}
}

func TestMethodParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "def add(x, y); end", expectedParams: []string{"x", "y"}},
		{input: `
		def print(x)
		end
		`, expectedParams: []string{"x"}},
		{input: "def test(x, y, z); end", expectedParams: []string{"x", "y", "z"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		methodStatement := program.Statements[0].(*ast.DefStatement)

		if len(methodStatement.Parameters) != len(tt.expectedParams) {
			t.Errorf("expect %d parameters. got=%d", len(tt.expectedParams), len(methodStatement.Parameters))
		}

		for i, expectedParam := range tt.expectedParams {
			testIdentifier(t, methodStatement.Parameters[i], expectedParam)
		}
	}
}

func TestCallExpression(t *testing.T) {
	input := `
		p.add(1, 2 * 3, 4 + 5)
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	callExpression := stmt.Expression.(*ast.CallExpression)

	if !testIdentifier(t, callExpression.Receiver, "p") {
		return
	}

	testMethodName(t, callExpression, "add")

	if len(callExpression.Arguments) != 3 {
		t.Fatalf("expect %d arguments. got=%d", 3, len(callExpression.Arguments))
	}

	testIntegerLiteral(t, callExpression.Arguments[0], 1)
	testInfixExpression(t, callExpression.Arguments[1], 2, "*", 3)
	testInfixExpression(t, callExpression.Arguments[2], 4, "+", 5)
}

func TestSelfCallExpression(t *testing.T) {
	input := `
		self.add(1, 2 * 3, 4 + 5);
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	callExpression := stmt.Expression.(*ast.CallExpression)

	self, ok := callExpression.Receiver.(*ast.SelfExpression)
	if !ok {
		t.Fatalf("expect receiver to be SelfExpression. got=%T", callExpression.Receiver)
	}

	if self.TokenLiteral() != "self" {
		t.Fatalf("expect SelfExpression's token literal to be 'self'. got=%s", self.TokenLiteral())
	}

	testMethodName(t, callExpression, "add")

	if len(callExpression.Arguments) != 3 {
		t.Fatalf("expect %d arguments. got=%d", 3, len(callExpression.Arguments))
	}

	testIntegerLiteral(t, callExpression.Arguments[0], 1)
	testInfixExpression(t, callExpression.Arguments[1], 2, "*", 3)
	testInfixExpression(t, callExpression.Arguments[2], 4, "+", 5)
}

func TestCallExpressionWithBlock(t *testing.T) {
	input := `
	[1, 2, 3, 4].each do |i|
	  puts(i)
	end
	`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	callExpression := stmt.Expression.(*ast.CallExpression)

	receiver := callExpression.Receiver
	if _, ok := receiver.(*ast.ArrayExpression); !ok {
		t.Fatalf("Expect receiver to be an Array. got=%T", receiver)
	}

	testMethodName(t, callExpression, "each")
	testIdentifier(t, callExpression.BlockArguments[0], "i")

	block := callExpression.Block
	exp := block.Statements[0].(*ast.ExpressionStatement).Expression
	testMethodName(t, exp, "puts")
}

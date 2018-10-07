package main

import "fmt"
import (
	"io/ioutil"
	"strings"
	"strconv"
	"errors"
)

var debugMode = 0

type Token struct {
	typ   string
	sval  string
}

type Ast struct {
	typ     string
	ival    int
	operand *Ast
	left    *Ast
	right   *Ast
}

var tokens []*Token
var tokenIndex int
var source string
var sourceInex int

func readFile(filename string) string {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func getc() (byte,error) {
	if sourceInex >= len(source) {
		return 0, errors.New("EOF")
	}
	r := source[sourceInex]
	sourceInex++
	return r, nil
}

func ungetc() {
	sourceInex--
}

func is_number(c byte) bool {
	debugPrint(fmt.Sprintf("is_numeric %c", c))
	return '0' <= c && c  <= '9'
}


func read_number(c1 byte) string {
	var chars = []byte{c1}
	for {
		c,err := getc()
		debugPrint("read_number")
		if err != nil {
			return string(chars)
		}
		if is_number(c) {
			chars = append(chars, c)
			continue
		} else {
			ungetc()
			return string(chars)
		}
	}
}

func is_space(c byte) bool {
	return  c == ' ' || c == '\t'
}

func skip_space() {
	for {
		c,err:= getc()
		if err != nil {
			return
		}
		if is_space(c) {
			continue
		} else {
			ungetc()
			return
		}
	}
}

func tokinize(s string) []*Token {
	var r []*Token
	s = strings.Trim(s, "\n")
	source = s
	for  {
		c, err := getc()
		if err != nil {
			return r
		}
		var tok *Token
		switch  {
		case c == 0:
			return r
		case is_number(c):
			val := read_number(c)
			tok = &Token{typ: "number", sval: val}
		case c == ' ' || c == '\t' :
			skip_space()
			tok = &Token{typ: "space", sval: " "}
		case c == '+':
			tok = &Token{typ: "punct", sval: fmt.Sprintf("%c", c)}
		default:
			fmt.Printf("c='%c'\n", c)
			panic("unknown char")
		}
		r = append(r, tok)
	}

	return r
}

func readToken() *Token {

	if tokenIndex <= len(tokens)-1 {
		r := tokens[tokenIndex]
		tokenIndex++
		return r
	}
	return nil
}

func parseUnaryExpr() *Ast {
	tok := readToken()
	if tok.typ == "space" {
		tok = readToken()
	}
	ival, _ := strconv.Atoi(tok.sval)
	return &Ast{
		typ: "uop",
		operand: &Ast{
			typ:  "int",
			ival: ival,
		},
	}
}

func parseExpr() *Ast {
	ast := parseUnaryExpr()
	for {
		tok := readToken()
		if tok == nil {
			return ast
		}
		if tok.typ == "space" {
			continue
		}
		if tok.typ != "punct" {
			return ast
		}
		if tok.sval == "+" {
			right := parseUnaryExpr()
			debugAst("right", right)
			return &Ast{
				typ:   "binop",
				left:  ast,
				right: right,
			}
		} else {
			fmt.Printf("unknown token=%v\n", tok)
			debugToken(tok)
			panic("internal error")
		}
	}

	return ast
}

func generate(ast *Ast) {
	fmt.Println("\t.globl	main")
	fmt.Println("main:")
	emitAst(ast)
	fmt.Println("\tret")
}

func emitAst(ast *Ast) {
	if ast.typ == "uop" {
		fmt.Printf("\tmovl	$%d, %%eax\n", ast.operand.ival)
	} else if ast.typ == "binop" {
		fmt.Printf("\tmovl	$%d, %%ebx\n", ast.left.operand.ival)
		fmt.Printf("\tmovl	$%d, %%eax\n", ast.right.operand.ival)
		fmt.Printf("\taddl	%%ebx, %%eax\n")
	} else {
		panic(fmt.Sprintf("unexpected ast type %s", ast.typ))
	}
}

func debugPrint(s string) {
	fmt.Printf("# %s\n", s)
}

func debugPrintVar(name string, v interface{}) {
	debugPrint(fmt.Sprintf("%s=%v", name, v))
}

func debugToken(tok *Token) {
	debugPrint(fmt.Sprintf("tok: type= %7s, sval=\"%s\"", tok.typ, tok.sval))
}

func debugTokens(tokens []*Token) {
	for _, tok := range tokens {
		debugToken(tok)
	}
}

func debugAst(name string, ast *Ast) {
	if ast.typ == "int" {
		debugPrintVar(name, fmt.Sprintf("%d", ast.ival))
	} else if ast.typ == "uop" {
		debugAst(name, ast.operand)
	} else if ast.typ == "binop" {
		debugPrintVar("ast.binop", ast.typ)
		debugAst("left", ast.left)
		debugAst("right", ast.right)
	}
}

func main() {
	debugMode = 1
	s := readFile("/dev/stdin")
	tokens = tokinize(s)
	if len(tokens) == 0 {
		panic("no token")
	}
	debugPrint("==== Start Dump Tokens ===")
	tokenIndex = 0
	debugTokens(tokens)
	debugPrint("==== End Dump Tokens ===")
	ast := parseExpr()
	debugPrint("==== Dump Ast ===")
	debugAst("root", ast)
	generate(ast)
}

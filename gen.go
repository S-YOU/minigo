package main

import "fmt"

func emit(format string, v ...interface{})  {
	fmt.Printf("\t" + format + "\n", v...)
}

func emitLabel(format string, v ...interface{})  {
	fmt.Printf(format + "\n", v...)
}

func emitDataSection() {
	emit(".data")

	// put dummy label
	emitLabel(".L0:")
	emit(".string \"%%d\\n\"")
}

func emitFuncMainPrologue() {
	emit(".text")
	emit(".globl	main")
	emitLabel("main:")
	emit("push %%rbp")
	emit("mov %%rsp, %%rbp")
}

func emitFuncMainEpilogue() {
	emit("leave")
	emit("ret")
}

func emitExpr(ast *Ast) {
	if ast.typ == "uop" {
		emit("movl	$%d, %%eax", ast.operand.ival)
	} else if ast.typ == "binop" {
		emit("movl	$%d, %%eax", ast.left.operand.ival)
		emit("movl	$%d, %%ebx", ast.right.operand.ival)
		if ast.op == "+" {
			emit("addl	%%ebx, %%eax")
		} else if ast.op == "-" {
			emit("subl	%%ebx, %%eax")
		} else if ast.op == "*" {
			emit("imul	%%ebx, %%eax")
		}
	} else if ast.typ == "string" {
		emit("lea .%s(%%rip), %%rax", ast.label)
	} else if ast.typ == "funcall" {
		emitFuncall(ast)
	} else {
		panic(fmt.Sprintf("unexpected ast type %s", ast.typ))
	}
}

var regs = []string{"rdi", "rsi"}

func emitFuncall(funcall *Ast) {
	fname := funcall.fname
	args := funcall.args
	for i, _ := range args {
		emit("push %%%s", regs[i])
	}

	for _, arg := range args {
		emitExpr(arg)
		emit("push %%rax")
	}

	for i, _ := range args {
		j := len(args) - 1 - i
		emit("pop %%%s", regs[j])
	}
	emit("mov $0, %%rax")
	emit("call %s", fname)

	for i, _ := range args {
		j := len(args) - 1 - i
		emit("pop %%%s", regs[j])
	}
}

func generate(expr *Ast) {
	emitDataSection()
	emitFuncMainPrologue()

	// call printf("%d\n", expr)
	astString := &Ast{
		typ: "string",
		label: "L0",
	}
	funcall := &Ast{
		typ:"funcall",
		fname : "printf",
		args: []*Ast{astString, expr},
	}

	emitExpr(funcall)

	emit("mov $0, %%eax") // return 0
	emitFuncMainEpilogue()
}

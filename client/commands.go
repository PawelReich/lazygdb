package client

import (
	"fmt"
	"strconv"
	"strings"
)

func (gdb *GdbClient) Interrupt() error {
	return gdb.gdb.Interrupt()
}

type Instruction struct {
	Address string `mapstructure:"address"`
	Inst    string `mapstructure:"inst"`
	Func    string `mapstructure:"func"`
}

type GdbAsmDisassemblyPayload struct {
	AsmInsns []Instruction `mapstructure:"asm_insns"`
}

func (gdb *GdbClient) DisassembleAroundPC(byteRange int) <-chan AsyncDecodedResult[GdbAsmDisassemblyPayload] {
	return SendDecodeAsync[GdbAsmDisassemblyPayload](gdb, "data-disassemble", "-s", fmt.Sprintf("$pc-%d", byteRange/2), "-e", fmt.Sprintf("$pc+%d", byteRange/2), "--", "0")
}

type GdbStackFrame struct {
	Address  string `mapstructure:"address"`
	Function string `mapstructure:"func"`
	FilePath string `mapstructure:"fullname"`
	FileLine string `mapstructure:"line"`
}

type GdbStackFramePayload struct {
	Frame GdbStackFrame `mapstructure:"frame"`
}

func (gdb *GdbClient) GetCurrentStackFrame() <-chan AsyncDecodedResult[GdbStackFramePayload] {
	return SendDecodeAsync[GdbStackFramePayload](gdb, "stack-info-frame")
}

type GdbRegisterNamesPayload struct {
	RegisterNames []string `mapstructure:"register-names"`
}

type RegisterValue struct {
	Index string `mapstructure:"number"`
	Value string `mapstructure:"value"`
}

type GdbRegisterValuesPayload struct {
	RegisterValues []RegisterValue `mapstructure:"register-values"`
}

type Register struct {
	Name   string
	Number string
	Value  string
}

func (gdb *GdbClient) GetRegisters() <-chan AsyncDecodedResult[[]Register] {
	ch := make(chan AsyncDecodedResult[[]Register], 1)

	go func() {

		namesFut := SendDecodeAsync[GdbRegisterNamesPayload](gdb, "data-list-register-names")
		valuesFut := SendDecodeAsync[GdbRegisterValuesPayload](gdb, "data-list-register-values", "x")

		names := <-namesFut
		values := <-valuesFut

		if names.Error != nil {
			ch <- AsyncDecodedResult[[]Register]{Error: names.Error}
			return
		}
		if values.Error != nil {
			ch <- AsyncDecodedResult[[]Register]{Error: values.Error}
			return
		}

		length := len(values.Result.RegisterValues)
		registers := make([]Register, 0, length)

		for _, register := range values.Result.RegisterValues {
			registerIndex, err := strconv.Atoi(register.Index)
			if err != nil {
				panic(err)
			}

			registers = append(registers, Register{
				Name:  names.Result.RegisterNames[registerIndex],
				Value: register.Value,
			})
		}

		ch <- AsyncDecodedResult[[]Register]{Result: registers}
	}()

	return ch
}

func (gdb *GdbClient) GetCachedSymbol(sym string) <-chan AsyncDecodedResult[string] {
	ch := make(chan AsyncDecodedResult[string], 1)

	if resolvedSymbol, ok := gdb.symbolLookup[sym]; ok {
		ch <- AsyncDecodedResult[string]{Result: resolvedSymbol}
		close(ch)
		return ch
	}

	go func() {
		defer close(ch)

		res := <-gdb.GetSymbol(sym)
		if res.Error == nil {
			gdb.symbolLookup[sym] = res.Result
		}
		ch <- res
	}()

	return ch
}

func (gdb *GdbClient) GetSymbol(sym string) <-chan AsyncDecodedResult[string] {
	ch := make(chan AsyncDecodedResult[string], 1)

	go func() {
		defer close(ch)
		res := <-gdb.SendConsoleCommandAsync("info symbol " + sym)
		if res.Error != nil {
			ch <- res
			return
		}

		if strings.Contains(res.Result, "No symbol matches") {
			res.Result = ""
		}
		res.Result = strings.Split(res.Result, " in ")[0]
		ch <- res
	}()

	return ch
}

type GdbStackListFramesEntry struct {
	Frame GdbStackListFramesFrame `mapstructure:"frame"`
}

type GdbStackListFramesPayload struct {
	Stack []GdbStackListFramesEntry `mapstructure:"stack"`
}

type GdbStackFrameArguments struct {
	Level string                       `mapstructure:"level"`
	Args  []GdbStacktraceFrameArgument `mapstructure:"args"`
}

type GdbStackListArgumentsEntry struct {
	Frame GdbStackFrameArguments `mapstructure:"frame"`
}

type GdbStackListArgumentsPayload struct {
	StackArgs []GdbStackListArgumentsEntry `mapstructure:"stack-args"`
}

type GdbStacktraceFrameArgument struct {
	Name  string `mapstructure:"name"`
	Value string `mapstructure:"value"`
}

type GdbStackListFramesFrame struct {
	Level     string                       `mapstructure:"level"`
	Address   string                       `mapstructure:"addr"`
	Function  string                       `mapstructure:"func"`
	FileName  string                       `mapstructure:"file"`
	FilePath  string                       `mapstructure:"fullname"`
	FileLine  string                       `mapstructure:"line"`
	Arch      string                       `mapstructure:"arch"`
	Arguments []GdbStacktraceFrameArgument `mapstructure:"args"`
}

func (gdb *GdbClient) GetStacktrace() <-chan AsyncDecodedResult[[]GdbStackListFramesFrame] {
	ch := make(chan AsyncDecodedResult[[]GdbStackListFramesFrame], 1)

	go func() {
		defer close(ch)

		framesFut := SendDecodeAsync[GdbStackListFramesPayload](gdb, "stack-list-frames")
		argsFut := SendDecodeAsync[GdbStackListArgumentsPayload](gdb, "stack-list-arguments", "2")

		frames := <-framesFut
		args := <-argsFut

		if frames.Error != nil {
			ch <- AsyncDecodedResult[[]GdbStackListFramesFrame]{Error: frames.Error}
			return
		}
		if args.Error != nil {
			ch <- AsyncDecodedResult[[]GdbStackListFramesFrame]{Error: args.Error}
			return
		}

		argumentsByLevel := make(map[string][]GdbStacktraceFrameArgument, len(args.Result.StackArgs))
		for _, entry := range args.Result.StackArgs {
			argumentsByLevel[entry.Frame.Level] = entry.Frame.Args
		}

		stacktrace := make([]GdbStackListFramesFrame, 0, len(frames.Result.Stack))
		for _, entry := range frames.Result.Stack {
			frame := entry.Frame
			frame.Arguments = argumentsByLevel[frame.Level]
			stacktrace = append(stacktrace, frame)
		}

		ch <- AsyncDecodedResult[[]GdbStackListFramesFrame]{Result: stacktrace}
	}()

	return ch
}

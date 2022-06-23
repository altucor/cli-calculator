package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

/*
Data types:
	 2t - tb - Binary
	 8t - to - Octal
	10t - td - Decimal
	16t - th - Hexadecimal - 0x \x

Out flags:
	b - binary
	o - octal
	d - decimal
	h - hexadecimal

	u - unsigned
	s - signed
	f - format
	c - color
	p - padding
	8,16,32,64 - bitmask
*/

func readSymbolsFromStream(input *strings.Reader, count uint64) (string, error) {
	buffer := make([]byte, count)
	var err error = nil
	var readed int = 0
	readed, err = input.Read(buffer)
	if err == io.EOF {
		return "", err
	}
	if readed != int(count) {
		return "", errors.New("read not full")
	}
	return string(buffer), err
}

type OperandType string

const (
	OperandTypeBinary      OperandType = "b"
	OperandTypeOctal       OperandType = "o"
	OperandTypeDecimal     OperandType = "d"
	OperandTypeHexadecimal OperandType = "h"
)

type StateMachineType int

const (
	StateMachineTypeUnknown StateMachineType = 0
)

type OperationType int

const (
	OperationTypeUnknown  OperationType = 0
	OperationTypeAdd      OperationType = 1
	OperationTypeSubtract OperationType = 2
	OperationTypeMultiply OperationType = 3
	OperationTypeDivide   OperationType = 4
)

func collectValue(input *strings.Reader) string {
	var alphabet string = "0123456789abcdefABCDEF+-" // - and + here for parsing values with sign
	var value string = ""
	for {
		symbol, err := readSymbolsFromStream(input, 1)
		if err == io.EOF {
			input.UnreadByte()
			return value
		}
		if value != "" && (symbol == "+" || symbol == "-") {
			input.UnreadByte()
			return value
		}
		if symbol == "" {
			return value
		}
		if strings.Contains(alphabet, symbol) {
			value += symbol
		} else {
			input.UnreadByte()
			return value
		}
	}
}

func readValue(reader *strings.Reader, opType OperandType) int64 {
	valueString := collectValue(reader)
	var value int64 = 0
	var err error = nil
	switch opType {
	case "b":
		value, err = strconv.ParseInt(valueString, 2, 64)
	case "o":
		value, err = strconv.ParseInt(valueString, 8, 64)
	case "d":
		value, err = strconv.ParseInt(valueString, 10, 64)
	case "h":
		value, err = strconv.ParseInt(valueString, 16, 64)
	}
	if err != nil {
		fmt.Errorf("Error parsing value")
	}
	//fmt.Println("Value:", value)
	return value
}

func doSingleOperation(op1 uint64, opType OperationType, op2 uint64) uint64 {
	switch opType {
	case OperationTypeAdd:
		return op1 + op2
	case OperationTypeSubtract:
		return op1 - op2
	case OperationTypeMultiply:
		return op1 * op2
	case OperationTypeDivide:
		return op1 / op2
	}
	return 0
}

type OutFormat struct {
	prefix       string
	padding      bool
	fmt          string
	sign         bool
	format       bool
	color        bool
	bitmaskWidth uint64
}

func collectOutputFmt(reader *strings.Reader) OutFormat {
	var bitmaskStr string = ""

	stru := OutFormat{
		prefix:       "",
		padding:      false,
		fmt:          "",
		sign:         false,
		format:       false,
		color:        false,
		bitmaskWidth: 64,
	}

	for {
		symbol, err := readSymbolsFromStream(reader, 1)
		if err == io.EOF {
			break
		}
		switch symbol {
		case "b":
			stru.prefix = "0b"
			stru.fmt = "b"
		case "o":
			stru.fmt = "o"
		case "d":
			stru.fmt = "d"
		case "h":
			stru.prefix = "0x"
			stru.fmt = "X"
		case "s":
			stru.sign = true
		case "f":
			stru.format = true
		case "c":
			stru.color = true
		case "p":
			stru.padding = true
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			bitmaskStr += symbol
		}
	}

	if bitmaskStr != "" {
		temp, _ := strconv.ParseInt(bitmaskStr, 10, 64)
		stru.bitmaskWidth = uint64(temp)
	}

	return stru
}

func formatValue(input uint64, cfg OutFormat) string {
	var padding string = ""
	if cfg.padding {
		switch cfg.fmt {
		case "b":
			padding = "0" + fmt.Sprintf("%d", cfg.bitmaskWidth)
		case "o":
			padding = "0" + fmt.Sprintf("%d", cfg.bitmaskWidth)
		case "d":
			padding = "0" + fmt.Sprintf("%d", cfg.bitmaskWidth)
		case "X":
			padding = "0" + fmt.Sprintf("%d", cfg.bitmaskWidth/4)
		}
	}
	var i int64 = 0
	var bitmask uint64 = 0
	for i < int64(cfg.bitmaskWidth) {
		bitmask |= 1 << i
		i++
	}
	var outData string = fmt.Sprintf(cfg.prefix+"%"+padding+cfg.fmt, (input & bitmask))
	return outData
}

func evaluateExpr(input string) string {
	input = strings.ReplaceAll(input, " ", "")
	reader := strings.NewReader(input)
	var accumulator uint64 = 0
	var operationType OperationType = OperationTypeUnknown
	var formattedData = ""
	for {
		symbol, err := readSymbolsFromStream(reader, 1)
		if err == io.EOF {
			break
		}
		switch symbol {
		case "+":
			operationType = OperationTypeAdd
		case "-":
			operationType = OperationTypeSubtract
		case "*":
			operationType = OperationTypeMultiply
		case "/":
			operationType = OperationTypeDivide
		case "t":
			opType, err := readSymbolsFromStream(reader, 1)
			if err == io.EOF {
				break
			}
			tempValue := uint64(readValue(reader, OperandType(opType)))
			if operationType != OperationTypeUnknown {
				accumulator = doSingleOperation(accumulator, operationType, tempValue)
				operationType = OperationTypeUnknown
			} else {
				accumulator = tempValue
			}
		case "=":
			//outFormat, err := readSymbolsFromStream(reader, 1)
			//if err == io.EOF {
			//	break
			//}
			fmt := collectOutputFmt(reader)
			formattedData = formatValue(accumulator, fmt)
			break
		}
	}
	fmt.Println(formattedData)
	return ""
}

func main() {
	fmt.Println(evaluateExpr("td10+td20+td50=hfp16"))
}

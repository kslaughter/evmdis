package evmdis

import (
    "fmt"
    "math/big"
)

/*type Analysis interface {
    func Process(*Instruction) Analysis
    func Combine(Analysis) Analysis
    func Copy() Analysis
    func Equals(Analysis) bool
}*/

type Instruction struct {
    Op          OpCode
    Arg         *big.Int
    Annotations *TypeMap
    //analyses map[reflect.Type]Analysis
}

func (self *Instruction) String() string {
    if self.Arg != nil {
        return fmt.Sprintf("%v 0x%x", self.Op, self.Arg)
    } else {
        return self.Op.String()
    }
}

type BasicBlock struct {
    Instructions    []Instruction
    Offset          int
    Next            *BasicBlock
    Annotations     *TypeMap
}

type Program struct {
    Blocks              []*BasicBlock
    JumpDestinations    map[int]*BasicBlock
    //Instructions map[int]*Instruction
}

func NewProgram(bytecode []byte) *Program {
    program := &Program{
        JumpDestinations:   make(map[int]*BasicBlock),
    }

    currentBlock := &BasicBlock{
        Offset: 0,
        Annotations: NewTypeMap(),
    }

    for i := 0; i < len(bytecode); i++ {
        op := OpCode(bytecode[i])
        size := op.OperandSize()
        var arg *big.Int
        if size > 0 {
            arg = big.NewInt(0)
            for j := 1; j <= size; j++ {
                arg.Lsh(arg, 8)
                if i + j < len(bytecode) {
                    arg.Or(arg, big.NewInt(int64(bytecode[i + j])))
                }
            }
        }

        if op == JUMPDEST {
            if len(currentBlock.Instructions) > 0 {
                program.Blocks = append(program.Blocks, currentBlock)
                newBlock := &BasicBlock{
                    Offset: i,
                    Annotations: NewTypeMap(),
                }
                currentBlock.Next = newBlock
                currentBlock = newBlock
            }
            currentBlock.Offset += 1
            program.JumpDestinations[i] = currentBlock
        } else {
            instruction := Instruction{
                Op: op,
                Arg: arg,
                Annotations: NewTypeMap(),
            }
            currentBlock.Instructions = append(currentBlock.Instructions, instruction)

            if op.IsJump() {
                program.Blocks = append(program.Blocks, currentBlock)
                newBlock := &BasicBlock{
                    Offset: i + size + 1,
                    Annotations: NewTypeMap(),
                }
                currentBlock.Next = newBlock
                currentBlock = newBlock
            }
        }
        i += size
    }

    if len(currentBlock.Instructions) > 0 || program.JumpDestinations[currentBlock.Offset] != nil {
        program.Blocks = append(program.Blocks, currentBlock)
    } else {
        program.Blocks[len(program.Blocks) - 1].Next = nil
    }

    return program
}

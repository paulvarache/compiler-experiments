# Compiler experiments

This repo contains the source code for a compiler written in golang as a learning exercise

## Setup

The compiler has a basic CLI. 

Use `compiler build files/main.c -o files/assembly.s` to build the assembly

Use `compiler print-ast files/main.c` to print a JSON representation of the AST

You can use `docker run --rm -it -v <srcdir>:/src gcc` to assemble the binary

```
gcc -m64 -g /src/files/main.c -o /src/out
/src/out
echo $?
```

This will print the returned value of the main function in `main.c`

## Debugging

Use `docker run -it --rm --entrypoint=/bin/sh -v <srcdir>:/src --cap-add=SYS_PTRACE --security-opt seccomp=unconfined chrahunt/docker-gdb` to run a container and used gdb

Start gdb with the binary using `gdb -tui /src/out` then use `layout asm` and `layout regs` to display the asm source and registers values

Use `start` to start debugging and `si` or `stepi` to step through the assembly code

## Structure

`buffer` takes a `io.Reader` and is used to move a cursor through the string. `Peek` looks at the following characters, `Move` moves the pointer, `Shift` takes the current string slice and returns a `[]byte`

`lexer` uses the buffer to look at the next character and move te cursor to grab the characters for a token. It also detects the token type based on the characters scanned. The next token can be grabbed from the stream by calling `Next`

`ast` defines the AST node types and utility functions to build the nodes based on tokens. An interface is used for the `Statement` and `Expression` nodes. Type assertion is used to generate the nodes

`parser` takes one token at a time and constructs the AST with `Program` as root by respecting a defined grammar

`generator` takes a program and generates assembly code for it.


## TODO

 - This is missing proper Unicode support. `PeekRune` can be used for that
 - Only very basic operations and types are supported

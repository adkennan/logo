package main

import (
	"bufio"
	"fmt"
	"github.com/adkennan/logo"
	"os"
)

func main() {

	source := ` 

TO fib :n 
	MAKE "a 0
	MAKE "b 1
	REPEAT :n [
		MAKE "c :a
		MAKE "a :b
		MAKE "b SUM :b :c

		PRINT :a
	]
END

TO blah :n 
	IFELSE EQUALP :n 1 [	
		PRINT "BLAH
	] [
		TYPE "BLAH\ 
		BLAH DIFFERENCE :n 1
	]
END
`
	scope, err := logo.ParseNonInteractiveScope(logo.CreateBuiltInScope(), source)
	if err != nil {
		panic(err)
	}

	bio := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("? ")
		line, _, err := bio.ReadLine()
		err = logo.Evaluate(scope, string(line))
		if err != nil {
			fmt.Println(err)
		}
	}
}

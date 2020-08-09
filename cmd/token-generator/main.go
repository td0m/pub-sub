package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/td0m/pub-sub/pkg/auth"
)

func main() {
	auth.Init(os.Getenv("JWT_SECRET"))
	user := readLine("user")
	fmt.Println("Please enter permissions separated by a comma")
	read, write := readLine("read permissions"), readLine("write permissions")
	token, err := auth.CreateToken(user, arr(read), arr(write))
	if err != nil {
		fmt.Println("error: ", err)
		return
	}
	fmt.Println("token: " + token)
}

func arr(s string) []string {
	return strings.Split(s, ",")
}

func readLine(field string) string {
	r := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", field)
	text, _ := r.ReadString('\n')
	return text[:len(text)-1]
}

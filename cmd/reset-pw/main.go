package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/users"
	"golang.org/x/crypto/ssh/terminal"
)

var errPasswordMismatch = errors.New("passwords did not match")

func main() {
	if err := configs.ReloadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	username, err := resolveUsername(os.Args[1:], os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to determine username: %v\n", err)
		os.Exit(1)
	}

	userRecord, err := findUser(username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to locate user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Resetting password for user %q.\n", userRecord.Username)

	newPassword, err := promptForPassword(os.Stdin, os.Stdout, int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read password: %v\n", err)
		os.Exit(1)
	}

	if err := userRecord.SetPassword(newPassword); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set password: %v\n", err)
		os.Exit(1)
	}

	if err := users.SaveUser(*userRecord); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Password updated for %q.\n", userRecord.Username)
}

func resolveUsername(args []string, stdin *os.File, stdout *os.File) (string, error) {
	if len(args) > 0 {
		username := strings.TrimSpace(args[0])
		if username == "" {
			return "", errors.New("username cannot be empty")
		}
		return username, nil
	}

	reader := bufio.NewReader(stdin)
	fmt.Fprint(stdout, "Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return "", errors.New("username cannot be empty")
	}

	return username, nil
}

func findUser(username string) (*users.UserRecord, error) {
	userRecord, err := users.LoadUser(username, true)
	if err != nil {
		return nil, err
	}

	return userRecord, nil
}

func promptForPassword(stdin *os.File, stdout *os.File, fd int) (string, error) {
	var reader *bufio.Reader
	if !terminal.IsTerminal(fd) {
		reader = bufio.NewReader(stdin)
	}

	fmt.Fprint(stdout, "New password: ")
	first, err := readPassword(stdin, fd, reader)
	if err != nil {
		return "", err
	}

	fmt.Fprintln(stdout)
	fmt.Fprint(stdout, "Confirm new password: ")
	second, err := readPassword(stdin, fd, reader)
	if err != nil {
		return "", err
	}
	fmt.Fprintln(stdout)

	if first != second {
		return "", errPasswordMismatch
	}

	return first, nil
}

func readPassword(stdin *os.File, fd int, reader *bufio.Reader) (string, error) {
	if terminal.IsTerminal(fd) {
		passwordBytes, err := terminal.ReadPassword(fd)
		if err != nil {
			return "", err
		}
		return string(passwordBytes), nil
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

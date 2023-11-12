package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// fzf
func fzf(cmd string) string {
	c := exec.Command("sh", "-c", cmd)
	c.Stdin = os.Stdin
	// c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	out, err := c.Output()
	if err != nil {
		fmt.Println(err)
	}

	return string(out)
}

func System(cmd string) int {
	c := exec.Command("sh", "-c", cmd)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()

	if err == nil {
		return 0
	}

	// Figure out the exit code
	if ws, ok := c.ProcessState.Sys().(syscall.WaitStatus); ok {
		if ws.Exited() {
			return ws.ExitStatus()
		}

		if ws.Signaled() {
			return -int(ws.Signal())
		}
	}

	return -1

}

func is_stdin_open() bool {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// fmt.Println("data is being piped to stdin")
		return true
	} else {
		// fmt.Println("stdin is from a terminal")
		return false
	}
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		if response == "" {
			return true
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" || response == "\n" || response == "" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

func main() {
	fmt.Println()
	out := fzf(`paru --color=always -Sl | sed -E 's: :/:; s/ (\x1b\[[0-9;]*m)?unknown-version/\1/' | fzf -m --ansi --preview='paru -Si {1} | grep --color=never -v "v " | bat -p -l yaml --color=always'`)
	output := strings.Split(out, "\n")
	paru, err := exec.LookPath("paru")
	if err != nil {
		fmt.Println(err)
		paru = "sudo pacman"
	}
	args := "-S"

	var pkgs []string
	for x := range output {
		pkg := output[x]
		strip := strings.Split(pkg, " ")
		// fmt.Println(strip[0])
		pkg = strip[0]
		pkgs = append(pkgs, pkg)
	}

	fmt.Println(pkgs)
	cmdstr := strings.Join(pkgs, " ")
	cmd := fmt.Sprintf("%s %s %s", paru, args, cmdstr)
	fmt.Println(cmd)
	if askForConfirmation("Install these packages?: ") {
		System(cmd)
	}
}

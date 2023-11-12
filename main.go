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

func statCmd() string {
	paru, err := exec.LookPath("paru")
	if err == nil {
		return paru
	} else {
		paru, err = exec.LookPath("pacman")
		sudo, _ := exec.LookPath("sudo")
		return fmt.Sprintf("%s %s", sudo, paru)
	}
}

var (
	args string
	aur  string
	tail []string
)

var usage = `USAGE:
paruz <paru-opts>
a FZF terminal UI for paru or pacman
sudo is invoked automatically if needed.
Multiple packages can be selected.

OPTIONS:
	-h | --help    show this help message
	-S | -Syu      pass these args to paru
	-R | -Rns      remove packages

** NOTE ** The first arguments are automatically passed to paru / pacman
paruz <args-to-be-passed>
if none are passed "-S" is used automatically

FZF-KEYS:
	TAB            Select/Deselect
	Shift+TAB      Deselect
	ENTER          Install/remove selected packages

Examples:
	paruz -Syu --nocleanafter
	parus -Rns

	`

func init() {
	// default is to install
	// args = "-S"
	if len(os.Args) >= 2 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			fmt.Println(usage)
			os.Exit(0)
		}

		if os.Args[1] != "" {
			// args = os.Args[1]
			args = strings.Join(os.Args[1:], " ")
		} else {
			args = "-S"
		}
	}
}

func main() {
	out := fzf(`paru --color=always -Sl | sed -E 's: :/:; s/ (\x1b\[[0-9;]*m)?unknown-version/\1/' | fzf -m --ansi --preview='paru -Si {1} | grep --color=never -v "v " | bat -p -l yaml --color=always'`)
	if out == "" {
		fmt.Println("Nothing selected")
		os.Exit(0)
	}

	output := strings.Split(out, "\n")

	paru := statCmd()
	if paru == "" {
		fmt.Println("Couldnt find paru or pacman")
		os.Exit(1)
	}

	var pkgs []string
	for x := range output {
		pkg := output[x]
		strip := strings.Split(pkg, " ")
		pkg = strip[0]
		pkgs = append(pkgs, pkg)
	}

	cmdstr := strings.Join(pkgs, " ")
	cmd := fmt.Sprintf("%s %s %s", paru, args, cmdstr)

	fmt.Println(pkgs)
	fmt.Println(cmd)

	if strings.Contains(args, "-R") {
		if askForConfirmation("Remove these packages?: ") {
			System(cmd)
		}
	}

	if strings.Contains(args, "-S") {
		if askForConfirmation("Install these packages?: ") {
			System(cmd)
		}
	}
}

package main

import (
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	cmd := exec.Command("niri", "msg", "event-stream")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		output := scanner.Text()
		lines := strings.Split(output, "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			line := lines[i]
			isWindowOpen, _ := regexp.MatchString("^Window opened or changed:.*", line)
			if isWindowOpen {
				tryCenterWindow()
			}
			isWindowClose, _ := regexp.MatchString("^Window closed: ", line)
			if isWindowClose {
				tryCenterWindow()
			}
		}
	}
}

func tryCenterWindow() {
	getwinCmd := exec.Command("niri", "msg", "windows")
	getwinOutputData, _ := getwinCmd.CombinedOutput()
	emptyLineRegexp := regexp.MustCompile(`\n\n+`)
	winList := emptyLineRegexp.Split(string(getwinOutputData), -1)

	// find focused
	for i := len(winList) - 1; i >= 0; i-- {
		curWin := winList[i]

		// "Window ID 1344: (focused)"
		isFocusedWin, _ := regexp.MatchString(`^Window\sID\s\d+:\s\(focused\)`, curWin)
		if isFocusedWin {
			curWinLines := strings.Split(curWin, "\n")
			workspaceIDLine := ""
			// find Workspace ID line
			for _, line := range curWinLines {
				isWorkspaceLine, _ := regexp.MatchString(`Workspace\sID:\s\d+$`, line)
				if isWorkspaceLine {
					workspaceIDLine = line
					break
				}
			}
			count := getWorkspaceWindowCount(workspaceIDLine)
			if count == 1 {
				exec.Command("niri", "msg", "action", "center-column").Start()
				break
			}
		}
	}
}

func getWorkspaceWindowCount(workspaceIDLine string) int {
	getWinsCmd := exec.Command("niri", "msg", "windows")
	grep := exec.Command("grep", workspaceIDLine)
	wc := exec.Command("wc", "-l")

	grep.Stdin, _ = getWinsCmd.StdoutPipe()
	wc.Stdin, _ = grep.StdoutPipe()
	var out bytes.Buffer
	wc.Stdout = &out

	wc.Start()
	grep.Start()
	getWinsCmd.Run()
	grep.Wait()
	wc.Wait()

	count, _ := strconv.Atoi(strings.TrimRight(out.String(), "\n"))
	return count
}

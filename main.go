package main

import (
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var operatedList = make([]string, 0)

func main() {
	cmd := exec.Command("niri", "msg", "event-stream")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		// "Window opened or changed: ..."
		output := scanner.Text()
		lines := strings.Split(output, "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			line := lines[i]
			isWindowOpen, _ := regexp.MatchString("^Window opened or changed:.*", line)
			if isWindowOpen {
				tryCenterWindow()
			}
		}
	}
}

func tryCenterWindow() {

	getWinCmd := exec.Command("niri", "msg", "windows")
	getWinOutputData, _ := getWinCmd.CombinedOutput()

	emptyLineRegexp := regexp.MustCompile(`\n\n+`)
	winList := emptyLineRegexp.Split(string(getWinOutputData), -1)

	// find focused
	// "Window ID 1344: (focused)"
	for i := len(winList) - 1; i >= 0; i-- {
		curWin := winList[i]

		isFocusedWin, _ := regexp.MatchString(`^Window\sID\s\d+:\s\(focused\)`, curWin)
		if isFocusedWin {

			curWinLines := strings.Split(curWin, "\n")
			head := curWinLines[0]
			lineParts := strings.Split(head, " ")
			winID := lineParts[2][:len(lineParts[2])-1]

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
			lineParts = strings.Split(workspaceIDLine, " ")
			if count == 1 && slices.Contains(operatedList, winID) == false {
				// operatedList = append(operatedList, winID)
				exec.Command("niri", "msg", "action", "center-column").Start()
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

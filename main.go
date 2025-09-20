package main

import (
	"bufio"
	"os/exec"
	"time"

	"github.com/buger/jsonparser"
)

var debounce = NewDebounce(time.Millisecond * 50)
var count = 0

func main() {
	cmd := exec.Command("niri", "msg", "--json", "event-stream")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		_, _, _, err := jsonparser.Get(scanner.Bytes(), "WindowOpenedOrChanged", "window", "id")
		if err == nil {
			tryCenterWindow()
		}
		_, _, _, err = jsonparser.Get(scanner.Bytes(), "WindowClosed")
		if err == nil {
			tryCenterWindow()
		}
	}
}

func tryCenterWindow() {
	if debounce.isLock() {
		return
	} else {
		debounce.TryLock()
	}
	// count += 1
	// fmt.Println("try...", count)

	focusedWindowData, _ := exec.Command("niri", "msg", "--json", "focused-window").CombinedOutput()
	focusedWindowworkspaceID, _ := jsonparser.GetInt(focusedWindowData, "workspace_id")
	count := 0

	windowsData, _ := exec.Command("niri", "msg", "--json", "windows").CombinedOutput()
	jsonparser.ArrayEach(windowsData, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		workspaceID, _ := jsonparser.GetInt(value, "workspace_id")
		if workspaceID == focusedWindowworkspaceID {
			count += 1
		}
		// if count > 1 {}
	})
	if count == 1 {
		exec.Command("niri", "msg", "action", "center-column").Start()
	}

}

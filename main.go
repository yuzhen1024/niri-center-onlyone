package main

import (
	"bufio"
	"os/exec"
	"time"

	"github.com/tidwall/gjson"
)

var debounce = NewDebounce(time.Millisecond * 50)
var count = 0

func main() {
	cmd := exec.Command("niri", "msg", "--json", "event-stream")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		result := gjson.GetBytes(scanner.Bytes(), "WindowOpenedOrChanged.window.id")
		if result.Index != 0 {
			tryCenterWindow()
		}
		result = gjson.GetBytes(scanner.Bytes(), "WindowClosed")
		if result.Index != 0 {
			tryCenterWindow()
		}
		// _, _, _, err = jsonparser.Get(scanner.Bytes(), "WindowLayouts")
		// if err == nil {
		// 	tryCenterWindow()
		// }
	}
}

func tryCenterWindow() {
	if debounce.isLock() {
		return
	} else {
		debounce.TryLock()
	}
	count += 1
	// fmt.Println("try...", count)

	focusedWindowData, _ := exec.Command("niri", "msg", "--json", "focused-window").CombinedOutput()
	if gjson.GetBytes(focusedWindowData, "is_floating").Bool() {
		return
	}
	focusedWindowworkspaceID := gjson.GetBytes(focusedWindowData, "workspace_id").Int()

	count := 0
	windowsData, _ := exec.Command("niri", "msg", "--json", "windows").CombinedOutput()
	gjson.ParseBytes(windowsData).ForEach(func(key, value gjson.Result) bool {
		workspaceID := gjson.Get(value.Raw, "workspace_id").Int()
		if workspaceID == focusedWindowworkspaceID {
			count += 1
		}
		if count > 1 {
			return false
		}
		return true
	})
	if count == 1 {
		go exec.Command("niri", "msg", "action", "center-column").Run()
	}
}

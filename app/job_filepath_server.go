package app

//

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

type ModJobFilePathServerStruct struct{}

var ModJobFilePathServer ModJobFilePathServerStruct

func (m ModJobFilePathServerStruct) JobCmdName() string {
	return "filepath-server"
}

func (m ModJobFilePathServerStruct) handleGetPath(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid Request Method", http.StatusMethodNotAllowed)
		fmt.Println("handleGetPath: Invalid Request Method")
		return
	}

	var req PathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to parse request", http.StatusBadRequest)
		fmt.Println("handleGetPath: Failed to Parse Request")
		return
	}

	// Gemini 交互，鉴权

}

func (m ModJobFilePathServerStruct) listenRequest() {
	http.HandleFunc("/getpath", m.handleGetPath)
}

func (m ModJobFilePathServerStruct) Run() {
	go listenRequest()

}

func (m ModJobFilePathServerStruct) ParseJob(filePathServerCmd *cobra.Command) *cobra.Command {
	filePathServerCmd.Run = func(_ *cobra.Command, _ []string) {
		m.Run()
	}
	return filePathServerCmd
}

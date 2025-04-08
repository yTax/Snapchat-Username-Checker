package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sqweek/dialog"
	"github.com/ytax/snapchat-username-checker/modules/requestparser"
	"google.golang.org/protobuf/proto"
)

func spoofRequest(username string) ([]string, uint32, error) {
	req := &requestparser.SuggestUsernameRequest{
		Username: &requestparser.SuggestUsernameRequest_NameWrapper{
			Name: username,
		},
		Locale:        "",
		SomethingFlag: 0,
		DeviceId:      "c798e85f-4511-66b0-889a-ef303fa6bfab",
		SessionId:     "6687bd20-731d-387c-e3b9-d47c5a90f410",
	}

	body, err := proto.Marshal(req)
	if err != nil {
		return nil, 0, fmt.Errorf("error marshalling request: %v", err)
	}

	payload := append([]byte{0}, uint32ToBytes(uint32(len(body)))...)
	payload = append(payload, body...)

	headers := map[string]string{
		"Content-Type":            "application/grpc",
		"TE":                      "trailers",
		"Grpc-Accept-Encoding":    "identity, deflate, gzip",
		"Grpc-Timeout":            "3S",
		"User-Agent":              "Snapchat/13.21.0.43 (moto g play (2021); Android 11#e00ca2#30; gzip) V/MUSHROOM grpc-c++/1.48.0 grpc-c/26.0.0 (android; cronet_http)",
		"Allow-Recycled-Username": "true",
		"X-Request-Id":            "63adac91-301f-46d3-a576-44c28d302153",
	}

	url := "https://aws.api.snapchat.com/snapchat.activation.api.SuggestUsernameService/SuggestUsername"
	reqHttp, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, fmt.Errorf("error creating HTTP request: %v", err)
	}

	for key, value := range headers {
		reqHttp.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(reqHttp)
	if err != nil {
		return nil, 0, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("error reading response body: %v", err)
	}

	if len(bodyResp) <= 5 {
		return nil, 0, fmt.Errorf("response body too short")
	}
	bodyResp = bodyResp[5:]

	var response requestparser.SuggestUsernameResponse
	err = proto.Unmarshal(bodyResp, &response)
	if err != nil {
		return nil, 0, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return response.GetSuggestions(), response.GetSuccessCode(), nil
}

func uint32ToBytes(n uint32) []byte {
	return []byte{
		byte(n >> 24),
		byte(n >> 16),
		byte(n >> 8),
		byte(n),
	}
}

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Cyan = "\033[36m"

func checkID(id string) int {
	if len(id) < 3 || len(id) > 15 {
		return 0
	}
	if !regexp.MustCompile(`^[a-zA-Z]`).MatchString(id) {
		return 0
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9._-]+$`).MatchString(id) {
		return 0
	}
	if strings.Count(id, "-") > 1 || strings.Count(id, "_") > 1 || strings.Count(id, ".") > 1 {
		return 0
	}
	if strings.HasSuffix(id, "-") || strings.HasSuffix(id, "_") || strings.HasSuffix(id, ".") {
		return 0
	}
	if strings.Contains(id, " ") {
		return 0
	}

	suggestions, successCode, err := spoofRequest(id)
	if err != nil {
		fmt.Println(Red+"[ "+Reset+strconv.Itoa(int(successCode))+Red+" ] "+Reset+"- "+Red+"Request error:"+Reset, err)
		fmt.Printf(Yellow+"RATELIMITED: Retrying %s in 10 seconds...\n"+Reset, id)
		return 2
	}
	if suggestions[0] == id {
		return 1
	}
	return 0
}

func pauseTerminal() {
	fmt.Println("\nPress Enter to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func getAllSessions() ([]string, string) {
	sessionsDir := "sessions"
	files, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, ""
	}

	var sessions []string
	var latestSession string
	maxSession := 0

	for _, file := range files {
		if file.IsDir() {
			sessionName := file.Name()
			if num, err := strconv.Atoi(strings.TrimPrefix(sessionName, "SESSION_")); err == nil {
				if num > maxSession {
					maxSession = num
					latestSession = sessionName
				}
				sessions = append(sessions, sessionName)
			}
		}
	}
	return sessions, latestSession
}

func getSessionPath(sessionName string) string {
	return filepath.Join("sessions", sessionName)
}

func readTargets(filename string) ([]string, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	var ids []string
	scanner := bufio.NewScanner(file)
	progress := 0

	if scanner.Scan() {
		firstLine := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(firstLine, "Progress:") {
			p := strings.TrimSpace(strings.TrimPrefix(firstLine, "Progress:"))
			if n, err := strconv.Atoi(p); err == nil {
				progress = n
			}
		} else if firstLine != "" {
			ids = append(ids, firstLine)
		}
	}

	for scanner.Scan() {
		id := strings.TrimSpace(scanner.Text())
		if id != "" {
			ids = append(ids, id)
		}
	}

	return ids, progress, scanner.Err()
}

func updateProgress(filename string, progress int, ids []string) error {
	lines := []string{fmt.Sprintf("Progress: %d", progress)}
	lines = append(lines, ids...)
	return os.WriteFile(filename, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

func showSplash() {
	fmt.Println(Blue + `
------------------------------------------------------------------------------------------------------------------------

  █████████                                          █████                 █████   
 ███░░░░░███                                        ░░███                 ░░███    
░███    ░░░  ████████    ██████   ████████   ██████  ░███████    ██████   ███████  
░░█████████ ░░███░░███  ░░░░░███ ░░███░░███ ███░░███ ░███░░███  ░░░░░███ ░░░███░   
 ░░░░░░░░███ ░███ ░███   ███████  ░███ ░███░███ ░░░  ░███ ░███   ███████   ░███    
 ███    ░███ ░███ ░███  ███░░███  ░███ ░███░███  ███ ░███ ░███  ███░░███   ░███ ███
░░█████████  ████ █████░░████████ ░███████ ░░██████  ████ █████░░████████  ░░█████ 
 ░░░░░░░░░  ░░░░ ░░░░░  ░░░░░░░░  ░███░░░   ░░░░░░  ░░░░ ░░░░░  ░░░░░░░░    ░░░░░   
                                  ░███                                   
                                  █████                                  
                                 ░░░░░                                                                                           
   █████████  █████                        █████                         
  ███░░░░░███░░███                        ░░███                          
 ███     ░░░  ░███████    ██████   ██████  ░███ █████  ██████  ████████  
░███          ░███░░███  ███░░███ ███░░███ ░███░░███  ███░░███░░███░░███ 
░███          ░███ ░███ ░███████ ░███ ░░░  ░██████░  ░███████  ░███ ░░░  
░░███     ███ ░███ ░███ ░███░░░  ░███  ███ ░███░░███ ░███░░░   ░███      
 ░░█████████  ████ █████░░██████ ░░██████  ████ █████░░██████  █████     
  ░░░░░░░░░  ░░░░ ░░░░░  ░░░░░░   ░░░░░░  ░░░░ ░░░░░  ░░░░░░  ░░░░░           
                                                                                                                        
------------------------------------------------------------------------------------------------------------------------` + Cyan + `
SNAPCHAT USERNAME AVAILABILITY CHECKER — by ytax - https://oguser.com/clarke

Send suggestions or report bugs at:` + Blue + ` https://github.com/ytax/snapchat-username-checker` + Cyan + `

This software will check for usernames inside "` + Blue + `targets.txt` + Cyan + `" feel free to replace the content of this file with a list of users
you want to check!

By default targets.txt is loaded with some shitty semi-og usernames.

I also recommend you to run the targets.txt file through a randomizer so you're not checking 
the usernames in alphabetic order` + Blue + `
------------------------------------------------------------------------------------------------------------------------` + Reset)
}

func main() {
	showSplash()

	sessions, latestSession := getAllSessions()

	fmt.Println(Cyan + "\n-> Existing sessions:" + Reset)
	for _, session := range sessions {
		if session == latestSession {
			fmt.Printf("  - %s"+Blue+" (LATEST SESSION)\n"+Reset, session)
		} else {
			fmt.Printf("  - %s\n", session)
		}
	}

	fmt.Println(Cyan + `
+-----------------------+
|` + Blue + ` 1. Start New Session` + Cyan + `  |
|` + Blue + ` 2. Resume Session` + Cyan + `     |
|` + Blue + ` 3. Exit` + Cyan + `               |
+-----------------------+` + Reset)
	fmt.Print(Cyan + "\n-> Choose an option" + Reset + ": ")

	var choice string
	fmt.Scanln(&choice)

	var sessionPath, targetsPath, outputPath string
	var isNewSession bool

	switch choice {
	case "1":

		newSessionName := "SESSION_" + strconv.Itoa(len(sessions)+1)
		sessionPath = getSessionPath(newSessionName)
		targetsPath = filepath.Join(sessionPath, "targets.txt")
		outputPath = filepath.Join(sessionPath, "output.txt")
		isNewSession = true
	
		
	
		// uses dialogue's fancy ass file picker instead of just reading from targets.txt
		fmt.Println(Cyan + "-> Choose your targets .txt file..." + Reset)
		time.Sleep(1 * time.Second) // lil delay so the user understands whats happening
		targetFile, err := dialog.File().Filter("Text Files", "txt").Title("Select the target .txt file for checking").Load()
		if err != nil {
			fmt.Println(Red+"Error selecting file:"+Reset, err)
			pauseTerminal()
			return
		}
	

		// create a new session folder
		if err := os.MkdirAll(sessionPath, os.ModePerm); err != nil {
			fmt.Println(Red+"Error creating session directory:"+Reset, err)
			pauseTerminal()
			return
		}

		// copy this to the respective session so we can resume later
		input, err := os.ReadFile(targetFile)
		if err != nil {
			fmt.Println(Red+"Error reading selected file:"+Reset, err)
			pauseTerminal()
			return
		}
		if err := os.WriteFile(targetsPath, input, 0644); err != nil {
			fmt.Println(Red+"Error copying selected file to session:"+Reset, err)
			pauseTerminal()
			return
		}
	
	case "2":
		fmt.Print(Cyan + "-> Enter the session name (e.g. SESSION_1): " + Reset)
		var chosenSession string
		fmt.Scanln(&chosenSession)
		sessionPath = getSessionPath(chosenSession)
		targetsPath = filepath.Join(sessionPath, "targets.txt")
		outputPath = filepath.Join(sessionPath, "output.txt")
		isNewSession = false
	case "3":
		fmt.Println(Red + "Exiting program. Hope you found some good users!" + Reset)
		os.Exit(0)
	default:
		fmt.Println(Red + "Invalid choice. Please restart the program." + Reset)
		return
	}

	if isNewSession {
		if err := os.MkdirAll(sessionPath, os.ModePerm); err != nil {
			fmt.Println(Red+"Error creating session directory:"+Reset, err)
			pauseTerminal()
			return
		}

		input, err := os.ReadFile("targets.txt")
		if err != nil {
			fmt.Println(Red+"Error reading targets.txt:"+Reset, err)
			pauseTerminal()
			return
		}
		if err := os.WriteFile(targetsPath, input, 0644); err != nil {
			fmt.Println(Red+"Error copying targets.txt:"+Reset, err)
			pauseTerminal()
			return
		}
	}

	ids, progress, err := readTargets(targetsPath)
	if err != nil {
		fmt.Println(Red+"Error reading targets file:"+Reset, err)
		pauseTerminal()
		return
	}

	if progress >= len(ids) {
		fmt.Println(Green + "All users have already been checked!" + Reset)
		pauseTerminal()
		return
	}

	file, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(Red+"Error creating/opening output file:"+Reset, err)
		pauseTerminal()
		return
	}
	defer file.Close()

	fmt.Println(Cyan + "\nChecking snapchat usernames...\n" + Reset)

	var checkTimestamps []time.Time

	for i := progress; i < len(ids); i++ {
		id := ids[i]
		now := time.Now()
		checkTimestamps = append(checkTimestamps, now)

		// Clean up old timestamps older than 60 seconds
		validTime := now.Add(-1 * time.Minute)
		for len(checkTimestamps) > 0 && checkTimestamps[0].Before(validTime) {
			checkTimestamps = checkTimestamps[1:]
		}

		cpm := len(checkTimestamps)

		switch checkID(id) {
		case 0:
			fmt.Printf(Red+"Not available: %s"+Reset+" | CPM: %d\n", id, cpm)
		case 1:
			fmt.Printf(Green+"Available: %s"+Reset+" | CPM: %d\n", id, cpm)
			file.WriteString(id + "\n")
		case 2:
			time.Sleep(10 * time.Second)
			i--
		}

		if err := updateProgress(targetsPath, i+1, ids); err != nil {
			fmt.Println(Red+"Failed to update progress:"+Reset, err)
		}
	}

	fmt.Println(Green + "\nCheck completed. Available users saved to " + outputPath + Reset)
	pauseTerminal()
}

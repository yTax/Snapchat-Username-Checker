package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"resty.dev/v3"
)

var scrapper = resty.New().SetTimeout(5 * time.Second)

// silly terminal colors
var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Magenta = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

func checkID(id string) int {
	url := "https://www.snapchat.com/add/" + id
	// here im adding some checks to reduce false positives, code is a bit effy and could probably be combined into a single expression but fuck it

	if len(id) < 3 || len(id) > 15 {
		return 0
	}

	// must start with a letter
	if !regexp.MustCompile(`^[a-zA-Z]`).MatchString(id) {
		return 0
	}

	// can only have letters, numbers, underscore, dots and -
	if !regexp.MustCompile(`^[a-zA-Z0-9._-]+$`).MatchString(id) {
		return 0
	}

	// max of 1 special character
	if strings.Count(id, "-") > 1 || strings.Count(id, "_") > 1 || strings.Count(id, ".") > 1 {
		return 0
	}

	// make sure that the special chars arent at the end
	if strings.HasSuffix(id, "-") || strings.HasSuffix(id, "_") || strings.HasSuffix(id, ".") {
		return 0
	}

	// obviously spaces arent allowed
	if strings.Contains(id, " ") {
		return 0
	}

	resp, err := scrapper.R().Get(url)
	if err != nil {
		fmt.Println(Red+"Request error:"+Reset, err)
		fmt.Printf(Yellow+"RATELIMITED: Retrying %s in 10 seconds.. \n"+Reset, id)

		return 2
	}

	switch resp.StatusCode() {
	case 404:
		return 1
	case 200:
		return 0
	default:
		return 0
	}
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

 ███████████   ███              █████                                █████   
░░███░░░░░███ ░░░              ░░███                                ░░███    
 ░███    ░███ ████  ████████   ███████   ████████   ██████   █████  ███████  
 ░██████████ ░░███ ░░███░░███ ░░░███░   ░░███░░███ ███░░███ ███░░  ░░░███░   
 ░███░░░░░░   ░███  ░███ ░███   ░███     ░███ ░░░ ░███████ ░░█████   ░███    
 ░███         ░███  ░███ ░███   ░███ ███ ░███     ░███░░░   ░░░░███  ░███ ███
 █████        █████ ████ █████  ░░█████  █████    ░░██████  ██████   ░░█████ 
░░░░░        ░░░░░ ░░░░ ░░░░░    ░░░░░  ░░░░░      ░░░░░░  ░░░░░░     ░░░░░  
                                                                             
                                                                     
   █████████  █████                        █████                             
  ███░░░░░███░░███                        ░░███                              
 ███     ░░░  ░███████    ██████   ██████  ░███ █████  ██████  ████████      
░███          ░███░░███  ███░░███ ███░░███ ░███░░███  ███░░███░░███░░███     
░███          ░███ ░███ ░███████ ░███ ░░░  ░██████░  ░███████  ░███ ░░░      
░░███     ███ ░███ ░███ ░███░░░  ░███  ███ ░███░░███ ░███░░░   ░███          
 ░░█████████  ████ █████░░██████ ░░██████  ████ █████░░██████  █████         
  ░░░░░░░░░  ░░░░ ░░░░░  ░░░░░░   ░░░░░░  ░░░░ ░░░░░  ░░░░░░  ░░░░░          
                                                                                                                        
------------------------------------------------------------------------------------------------------------------------` + Cyan + `
PINTREST USERNAME AVAILABILITY CHECKER — by ytax - https://oguser.com/clarke

Send suggestions or report bugs at:` + Blue + ` https://github.com/ytax/pintrest-username-checker` + Cyan + `

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
			fmt.Println(Red+"Error creating session directory (this is really weird, make sure controlled folder access isnt blocking the program or that you arent running from a place where the program doesnt have permission to write.):"+Reset, err)
			pauseTerminal()
			return
		}

		input, err := os.ReadFile("targets.txt")
		if err != nil {
			fmt.Println(Red+"Error reading targets.txt (this is really weird, make sure controlled folder access isnt blocking the program or that you arent running from a place where the program doesnt have permission to write.):"+Reset, err)
			pauseTerminal()
			return
		}
		if err := os.WriteFile(targetsPath, input, 0644); err != nil {
			fmt.Println(Red+"Error copying targets.txt (this is really weird, make sure controlled folder access isnt blocking the program or that you arent running from a place where the program doesnt have permission to write.):"+Reset, err)
			pauseTerminal()
			return
		}
	}

	ids, progress, err := readTargets(targetsPath)
	if err != nil {
		fmt.Println(Red+"Error reading targets file (this is really weird, make sure controlled folder access isnt blocking the program or that you arent running from a place where the program doesnt have permission to write.):"+Reset, err)
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
		fmt.Println(Red+"Error creating/opening output file (this is really weird, make sure controlled folder access isnt blocking the program or that you arent running from a place where the program doesnt have permission to write.):"+Reset, err)
		pauseTerminal()
		return
	}
	defer file.Close()

	fmt.Println(Cyan + "\nChecking pintrest usernames...\n" + Reset)

	for i := progress; i < len(ids); i++ {
		id := ids[i]

		switch checkID(id) {
		case 0:
			fmt.Printf(Red+"Not available: %s\n"+Reset, id)
		case 1:
			fmt.Printf(Green+"Available: %s\n"+Reset, id)
			file.WriteString(id + "\n")
		case 2:
			time.Sleep(10 * time.Second)
			i--
		}

	}

	fmt.Println(Green + "\nCheck completed. Available users saved to " + outputPath + Reset)
	pauseTerminal()
}

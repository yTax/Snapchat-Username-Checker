package main

import (
	"bufio"
	"bytes"

	//"encoding/hex" // enable for debugging
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sqweek/dialog"
	"github.com/ytax/snapchat-username-checker/modules/requestparser"
	"golang.org/x/net/http2"
	"golang.org/x/net/proxy"
	"google.golang.org/protobuf/proto"
)


var proxyToggle int

var currentProxyIndex = 0

func validateProxies(proxies []string) []string {

	choice := 0
    
    fmt.Println(Cyan + "\n-> Do you want to validate your proxies? (This is recommended as the validation will also remove slow proxies):" + Reset)
    fmt.Println("1. Yes")
    fmt.Println("2. No")
    
    for {
        fmt.Print(Cyan + "Enter your choice: " + Reset)
        _, err := fmt.Scanln(&choice)
        if err != nil {
            fmt.Println(Red+"ERROR: "+Reset+"Invalid input")
            continue
        }
        
        if choice == 1 || choice == 2 {
            break
        }
        fmt.Println(Red+"ERROR: "+Reset+"You must choose between 1 and 2")
    }

	if choice == 2 {
		return proxies // basically if the user doesnt want to check the proxies it'll return the unchecked list
	}


    var validProxies []string
    for _, proxy := range proxies {
        if isProxyValid(proxy) {
			fmt.Println(Green+"Valid proxy:"+Reset, proxy)
            validProxies = append(validProxies, proxy)
        } else {
            fmt.Println(Red+"Invalid proxy or too slow:"+Reset, proxy)
        }
    }
    return validProxies
}

func isProxyValid(proxyAddr string) bool {
	// timeout for the proxies, you can set this to wtv you want but i think 5 secs is a good value
	// if your proxies take longer than this to reply they're way too slow to make this efficient
	// optimally your proxies should take no longer than 3 seconds as that makes the request way too slow, but i understand that some people want to use free proxies
	// so i'd say 5 seconds is a good middleground
	// when i release the V2 of the checker this will be a configurable variabl instead of something that's hardcoded
	// in the meantime, if you have really fast projects feel free to change this timeout and compile the program yourself to make it faster


	// update on this: fuck slow proxy support, it makes no sense
	// if your proxies take more than 2-3 seconds to contact the server bruteforcing through the ratelimit will yield better results than using them..

    timeout := 2 * time.Second
    netDialer := &net.Dialer{
        Timeout: timeout,
    }

    // this is the equivalent of a pointer in go, pretty cool
	// we will use this to exfil data from the goroutine that checks the proxies without blocking the main thread
	// this is SLIGHTLY more efficient because it sets an absolute timeout
    result := make(chan bool)


    go func() {
        dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, netDialer)
        if err != nil {
            result <- false
            return
        }

        conn, err := dialer.Dial("tcp", "aws.api.snapchat.com:443")
        if err != nil {
            result <- false
            return
        }

        conn.Close()
        result <- true
    }()

    // wait until the connection is finished, if it takes more than 3 seconds automatically assume as dead and skip to next
	// fast proxies or nothing

    select {
    case res := <-result: 
        return res
    case <-time.After(timeout): 
        return false
    }
}


func readProxies(filename string) ([]string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var proxies []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        proxy := strings.TrimSpace(scanner.Text())
        if proxy != "" {
            proxies = append(proxies, proxy)
        }
    }
    return proxies, scanner.Err()
}


func proxyPrompt() int {
    choice := 0
    
    fmt.Println(Cyan + "\n-> Choose an option:" + Reset)
    fmt.Println(Cyan+"1."+Reset+" Proceed without proxies\n")
    fmt.Println(Cyan+"2."+Reset+" Use SOCKS5 proxies")
	fmt.Println(Red+"WARNING: "+Reset+"Your proxies NEED to be SOCKS5 and the must be in a IP:PORT format WITHOUT having 'socks5://' behind them.")
	fmt.Println(Red+"WARNING: "+Reset+"If your proxies are slow RUN PROXYLESS.This tool was NOT designed with slow proxies in mind.")
    fmt.Println(Red+"WARNING: "+Reset+"It is faster to run without proxies than with slow proxies.")

    for {
        fmt.Print(Cyan + "Enter your choice: " + Reset)
        _, err := fmt.Scanln(&choice)
        if err != nil {
            fmt.Println(Red+"ERROR: "+Reset+"Invalid input")
            continue
        }
        
        if choice == 1 || choice == 2 {
            break
        }
        fmt.Println(Red+"ERROR: "+Reset+"You must choose between 1 and 2")
    }
    
    return choice
}

type RequestConfig struct { // finally made a struct for this shit
    Username    string
    UseProxies  bool
    Proxies     []string
}


// this code is fucking horrendous btw ill rewrite it for the V2 of the project
// these functions are probably the worst offenders when it comes to dogshit code
// not only is the logic behind them very stupid but its also kinda poorly written and has lots of bandaid fixes
// it works tho so fuck it we ball
// this is mostly my fault because instead of rewriting everything from scratch im using most of the logic from my old checkers
// and the logic of a PoC i wrote using a different endpoint. This lead to me being forced to make a bunch of bandaid fixes for some of the code.
// still its efficient enough and it works, just not the prettiest to look at nor the best way to write it.

func spoofRequest(config RequestConfig) ([]string, uint32, error) {
	username := config.Username

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

func proxiedSpoofRequest(config RequestConfig) ([]string, uint32, error)  {
	username := config.Username


	// creates a dialer that we will use to proxy the requests made
	proxies := config.Proxies


	// even tho we're using % to keep it in bound im adding in this check because of some weird behaviour i saw during testing
	// i have no fucking clue how it sometimes somehow gets out of bounds
	// ill rewrite this entire structure once i rewrite the code for V2
	// nvm im autistic
	// if currentProxyIndex >= len(proxies){ 
	// 	currentProxyIndex = 0
	// }

	proxyURL := proxies[currentProxyIndex] // this gets the current proxy that we should use in case of a ratelimit or invalid proxy this gets cycled
	dialer, err := proxy.SOCKS5("tcp", proxyURL, nil, proxy.Direct)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating SOCKS5 dialer: %v", err)
	}
	
	transport := &http.Transport{
		Dial: dialer.Dial,
	}
	err = http2.ConfigureTransport(transport)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to configure HTTP/2 transport: %v", err)
	}


	client := &http.Client{
		Transport: transport,
		Timeout: 3 * time.Second, // if your proxies take longer than this forget about using them 
	}

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

		// // Print full request (headers + body)
		// fmt.Println("=== REQUEST DUMP ===")
		// fmt.Printf("POST %s\n", url)
		// fmt.Println("Headers:")
		// for key, value := range reqHttp.Header {
		// 	fmt.Printf("%s: %s\n", key, value)
		// }

		// fmt.Println("\nBody (hex dump):")
		// fmt.Println(hex.Dump(payload)) // useful for raw binary gRPC body

		if err != nil {
			return nil, 0, fmt.Errorf("error creating HTTP request: %v", err)
		}

		for key, value := range headers {
			reqHttp.Header.Set(key, value)
		}

		resp, err := client.Do(reqHttp)
		if err != nil {
			return nil, 0, fmt.Errorf("error sending request: %v", err)
		}
		defer resp.Body.Close()

		bodyResp, err := io.ReadAll(resp.Body)

		// // === Print response headers === this is for debugging
		// fmt.Println("=== RESPONSE HEADERS ===")
		// for key, value := range resp.Header {
		// 	fmt.Printf("%s: %s\n", key, value)
		// }
		// fmt.Println("=== RAW RESPONSE BODY (hex dump) ===")
		// fmt.Println(hex.Dump(bodyResp))

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

func checkID(config RequestConfig) int {
	id := config.Username
	proxyToggle := config.UseProxies

	// this clusterfuck of ifs will basically pre-check for invalid IDs and return them as not avaliable before sending it through to the api
	// this not only spares requests on your proxies, but also prevents wasting time on a request we already know will return negative
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


	// the proxying is toggled on and off, this is intentional because by default it will always be fast to send an unproxied request.
	// basically the program will send unproxied requests until it hits it's first ratelimit, then it'll send proxied requests for around 60 seconds
	// after those 60 seconds it will default back to unproxied, this not only saves proxy bandwith but also is WAY more efficient overall
	// this kinda makes slow proxies that take 2-3 seconds to fulfill the request usable



	if proxyToggle {
		suggestions, successCode, err := proxiedSpoofRequest(config)
		if err != nil {
			fmt.Println(Red+"[ "+Reset+strconv.Itoa(int(successCode))+Red+" ] "+Reset+"- "+Red+"Request error:"+Reset, err)
			
			return 2
		}
		if suggestions[0] == id {
			return 1
		}
		return 0
	}

	
	suggestions, successCode, err := spoofRequest(config)
	if err != nil {
		fmt.Println(Red+"[ "+Reset+strconv.Itoa(int(successCode))+Red+" ] "+Reset+"- "+Red+"Request error:"+Reset, err)
		
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


	switch choice {
	case "1":

		newSessionName := "SESSION_" + strconv.Itoa(len(sessions)+1)
		sessionPath = getSessionPath(newSessionName)
		targetsPath = filepath.Join(sessionPath, "targets.txt")
		outputPath = filepath.Join(sessionPath, "output.txt")

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


	case "3":
		fmt.Println(Red + "Exiting program. Hope you found some good users!" + Reset)
		os.Exit(0)
	default:
		fmt.Println(Red + "Invalid choice. Please restart the program." + Reset)
		return
	}

	// handles the proxy/proxyless choice
	proxyToggle = proxyPrompt()

	var proxies []string
	var proxyChoice bool
	var proxyList [] string

	if proxyToggle == 2 {
		proxyChoice = true

		fmt.Println(Cyan + "-> Choose your proxies .txt file..." + Reset)
		time.Sleep(1 * time.Second) // Delay for user clarity
		proxyFile, err := dialog.File().Filter("Text Files", "txt").Title("Select the proxies .txt file").Load()
		if err != nil {
			fmt.Println(Red+"Error selecting file:"+Reset, err)
			pauseTerminal()
			return
		}
	
		proxies, err = readProxies(proxyFile)
		if err != nil {
			fmt.Println(Red+"Error reading proxies file:"+Reset, err)
			pauseTerminal()
			return
		}

		proxyList = validateProxies(proxies) // will ask if the user wants to validate their proxies (u should always do this btw)
		

	} else {
		proxyChoice = false
		proxyList = nil
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



	userCfg := RequestConfig{Username: "empty", UseProxies: proxyChoice, Proxies: proxyList}

	proxySwitch := false // this way we can start it proxyless and then move to proxied once we hit a limit


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

		userCfg.Username = id

		// if the user wants to run fully proxyless we will never attempt to use proxies
		if !proxyChoice {
			switch checkID(userCfg) {
			case 0:
				fmt.Printf(Red+"Not available: %s"+Reset+" | CPM: %d\n", id, cpm)
			case 1:
				fmt.Printf(Green+"Available: %s"+Reset+" | CPM: %d\n", id, cpm)
				file.WriteString(id + "\n")
			case 2:
				fmt.Printf(Yellow+"RATELIMITED: Retrying %s in 10 seconds...\n"+Reset, id)
				time.Sleep(10 * time.Second) 
				i--
			}
		} else{ 
			// however if the user wants to use proxies we will do the following:
			// the UseProxies inside the config will be toggled on an off, first we will begin a proxyless check as this is faster than running proxied checks
			// once we hit our first ratelimit we will run proxyless for 60 seconds, after those 60 seconds we will return to proxyless
			// this logic makes checking MUCH faster

			userCfg.UseProxies = proxySwitch // updates to our toggle that changes over execution depending if we're ratelimited or not

			switch checkID(userCfg) {
			case 0:
				fmt.Printf(Red+"Not available: %s"+Reset+" | CPM: %d\n", id, cpm)
			case 1:
				fmt.Printf(Green+"Available: %s"+Reset+" | CPM: %d\n", id, cpm)
				file.WriteString(id + "\n")
			case 2:

				// this codebase needs a rewrite LMAO
				
				if userCfg.UseProxies{
					currentProxyIndex = (currentProxyIndex + 1) % len(userCfg.Proxies)
	
					fmt.Printf(Yellow+"RATELIMITED or INVALID PROXY: Retrying %s in 1 seconds...\n"+Reset, id)
					fmt.Println(Yellow + "Rotating to proxy: " + userCfg.Proxies[currentProxyIndex] + Reset)

					// i made it even more efficient nvm
					//time.Sleep(1 * time.Second) // i know the delay seems pointless but if it isnt here the speed ironically goes down due to anti spam
					i--
				} else {


					fmt.Printf(Yellow+"RATELIMITED: Retrying %s in 1 seconds...\n"+Reset, id)
					fmt.Println(Yellow + "Ratelimit detected activating proxy: " + userCfg.Proxies[currentProxyIndex] + Reset)
					//time.Sleep(1 * time.Second)

					proxySwitch = true // activates the proxies

					go func() { // run this under a routine so it (obviously) doesnt stop the checking
						time.Sleep(30 * time.Second)
						proxySwitch = false // disables proxying
						userCfg.UseProxies = false // just to make sure
						fmt.Println(Green+"[ "+Reset+"!"+Green+" ] "+Reset+"- "+Green+"30 seconds passed. Proxy disabled returning to proxyless for speed until the next ratelimit."+Reset)
					}()

					i-- 
				}
			}



		}
		

		if err := updateProgress(targetsPath, i+1, ids); err != nil {
			fmt.Println(Red+"Failed to update progress:"+Reset, err)
		}
	}

	fmt.Println(Green + "\nCheck completed. Available users saved to " + outputPath + Reset)
	print(proxyToggle)
	pauseTerminal()
}

# Pintrest Username Availability Checker

This is a command-line tool written in Go that allows you to check the availability of pintrest usernames.
This tool also has a session feature that allows you to save your progress so you dont have to recheck usernames you already checked previously.

## Features

- **Check Pintrest Username availability**: Check if a Github Username is available or already claimed.
- **Session management**: Create new sessions or resume existing ones.
- **Progress tracking**: The tool remembers where it left off in case of interruptions.
- **Proxyless**: This tool does not require you to provide proxies.

## Installation

You can download the latest release of the Pintrest Username Checker from the [releases page](https://github.com/yTax/Pintrest-Username-Checker/releases) on GitHub. Simply go to the page, choose the latest version, and download the zip.


## Build Instructions

1. Clone the repository:

   ```bash
   git clone https://github.com/yTax/Pintrest-Username-Checker.git
   ```

2. Navigate to the project folder:

   ```bash
   cd pintrest-username-checker
   ```

3. Build the project:

   ```bash
   go build
   ```

4. Run the program:

   ```bash
   ./main.exe
   ```

## Usage

When you run the tool, you'll be greeted with a menu to select your desired action. You can:

1. **Start a New Session**: Create a new session and begin checking usernames from a `targets.txt` file.
2. **Resume an Existing Session**: Choose an existing session to continue checking usernames.
3. **Exit**: Exit the program.

The program will check each username from your `targets.txt` file. If an username is available, it will be saved to the `output.txt` file. The program will also track your progress, so you can stop and resume at any time.

### Example Session Workflow

1. **Start a New Session**:
   - If no session exists, the tool will create a new session automatically and read from the `targets.txt` file.
   - By default, this file contains some random semi-og usernames, you can replace the content of this file with whatever you want the software to check.
   - I also HEAVILY recommend you to run your list through a randomizer so that you arent checking the targets in alphabetic order.
2. **Resume a Session**:
   - If sessions exist, you can select one to resume from where you left off.
   - You will be shown the available sessions, and you can write it's name to select it.

## Known Issues

- The tool may take time depending on the number of usernames being checked. This is due to the fact that the size of the request is quite big, a possible way to fix this would be implementing some form of concurrency.
- If you delete the targets.txt file the tool will stop working because it wont be able to read the targets.

## To-do List

- [ ] Add an option to generate targets (3c, 3l, 4c and 4l usernames).
- [ ] Add discord webhook support.
- [ ] Add some comments because i was too lazy to do it. I think the code is very easy to understand though.
- [ ] Add concurrency as this is probably the only checker that i've made so far which would greatly benefit from it.

## Credits

This tool was created by [ytax](https://github.com/ytax).

Feel free to contribute or open issues if you encounter any bugs or have suggestions for new features.

---

Enjoy checking your usernames and have fun!
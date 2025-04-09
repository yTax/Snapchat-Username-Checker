# 👻 Snapchat Username Checker

A terminal-based Snapchat username availability checker with SOCKS5 proxy support, session saving, a proxyless mode and real-time CPM tracking. 
It leverages Snapchat's private gRPC API (acquired from the mobile app) to deliver fast, accurate results — **proxyless or proxied**.
It's a bit ugly but incredibly fast and efficient — V2 rewrite coming soon™.

> Written in Go. Reverse engineered straight from the Snapchat mobile app.

**Created by [ytax](https://github.com/ytax)**

---

## 📚 Table of Contents

### ⚙️ Core
- [Features](#features)
- [Usage](#usage)
- [File Formats](#file-formats)
- [Build Instructions](#build-instruction)

### 🔍 Under the Hood
- [How It Works](#how-it-works-and-how-it-was-built)
- [Warnings](#warnings)
- [Proto & gRPC](#proto--grpc)

### 📌 Help & Info
- [FAQ](#faq)
- [What Makes a Valid Username?](#what-makes-a-valid-username)
- [Support](#support)
- [Known Issues / Limitations](#known-issues--limitations)

### 💻 Dev & Roadmap
- [Developer Notes](#developer-notes)
- [Roadmap (V2 Preview)](#roadmap-v2-preview)
- [Contributing](#contributing)
- [Disclaimer](#disclaimer)

### 📸 Extras
- [Screenshots](#screenshots)


---

## ⚙️ Features

- 🔍 Fastest Snapchat checker you'll find (gRPC based)
- ⚡ Average of **260 checks per minute**
- 🔌 Proxyless execution support
- 🧠 Built-in username filtering for invalid formats
- 🌐 SOCKS5 Proxy support (recommended for ratelimit bypass)
- 💾 Auto session saving — resume your checks anytime
- 📈 CPM (Checks Per Minute) calculation
- 🔁 **Smart proxy switching** — proxyless > proxied > proxyless for speed
- 📄 Thoroughly documented and commented code. Feel free to use this endpoint or fork this repo!

---

## 🚀 Usage

You can download the latest release of the Snapchat Username Checker from the [releases page](https://github.com/yTax/snapchat-username-checker/releases) on GitHub. Simply go to the page, choose the latest version, and download the zip.

After opening you'll be prompted with:

- New Session or Resume Session
- Use proxies or go proxyless after choosing whether to start a new session or resume an existing one
- Select your target usernames file (via GUI dialog)
- (If using proxies) select your proxy list (SOCKS5 only and make sure they are fast)
- If you want to validate your proxies (you should always do this)

### ✍️ File formats

#### `targets.txt` (input)
```
username1
username2
...
```

#### `sessions/SESSION_X/targets.txt` (Saved Sessions)

```
Progress: 2
username1
username2
...
```

> `Progress` line helps resume exactly where you left off.

#### `proxies.txt` (optional, SOCKS5 only due to Snapchat's API only working with gRPC and HTTP2)
```
127.0.0.1:1080
88.198.24.108:1080
...
```

---

## 📦 Build Instruction

> Requires **Go 1.20+** and `protoc` to regenerate gRPC modules (optional).

### 1. Clone the repo

```bash
git clone https://github.com/ytax/snapchat-username-checker.git
cd snapchat-username-checker
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Compile

```bash
go build
```

### 4.1 (Optional) Compile your own proto file
```bash
# installs the go modules for protoc and protobuf in case you dont have them
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest 
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 4.2 (Optional) Build your Go module with the proto file
```bash
# this command will create modules/requestparser/suggest_username.pb.go
protoc suggest_username.proto --go_out=.
```

---

## 🧠 How it works and How it was built

This tool was built by reverse-engineering Snapchat's mobile API. After bypassing the app’s SSL pinning, traffic was intercepted via a debug proxy, allowing access to all the internal gRPC endpoints.

Alongside the endpoint discovery, the tool includes a fairly precise recreation of Snapchat’s .proto files — ensuring that request and response structures are fully compatible with Snapchat's gRPC backend. This makes every check accurate and native, just like the app itself.

It specifically uses: `https://aws.api.snapchat.com/snapchat.activation.api.SuggestUsernameService/SuggestUsername`

This is the same endpoint the Snapchat app uses to suggest alternative usernames during signup. The logic is simple and reliable:

> If the **first suggestion returned matches the username you entered**, the username is **available**.

### ✅ Why This Matters

- **No headless browsers**  
- **No terrible web endpoints that yield inaccurate results**  
- **Incredibly fast and efficient**

Just raw, direct gRPC requests to Snapchat’s infrastructure — resulting in **fast**, **accurate**, and **efficient** username checks.


---

## ⚠️ Warnings

- **Your proxies MUST be SOCKS5**, in `IP:PORT` format. No `socks5://` prefix. This is due to SOCKS5 being the only proxies that can communicate in HTTP2 with gRPC.
- **Slow proxies is a TERRIBLE bad idea.** Proxyless is faster than proxied with slow proxies, some free proxies like the ones from [proxifly's github](https://github.com/proxifly/free-proxy-list) **may** work.
- **Rate-limits are automatically handled** by switching to proxies for a short cooldown.
- **Yes, the code is messy.** It's functional, but a rewrite is planned for V2.

---

## 🛠 Proto & gRPC

Again, if you want to regen your own gRPC or Go module:

```bash
protoc suggest_username.proto --go_out=. --go-grpc_out=.
```

Output module will be in `./modules/requestparser`.

---

## ❓ FAQ

**Q: Why is this checker so fast and what makes it different?**  
A: It talks directly to Snapchat's internal gRPC endpoint used by the mobile app. This results in faster responses and less overall ratelimits.

**Q: Can I use HTTP proxies?**  
A: No. Only SOCKS5. HTTP proxies won't work with HTTP/2 & gRPC.

**Q: It’s not checking properly, what’s wrong?**  
A: Check your proxies. Or just run proxyless. Most issues come from slow/dead proxies.

**Q: I'm using proxies, but it looks like some checks are still proxyless. Why?**
A: This is intentional. Proxyless requests are significantly faster because they avoid the overhead of routing through a proxy.

When you allow the tool to use proxies, the checker starts by sending requests **without a proxy** to maximize speed. When it hits a rate limit, it automatically switches to **proxy mode**, using your SOCKS5 proxies to continue checking. Once the rate limit on the proxyless connection expires, it switches back to proxyless mode to take advantage of it's speed.

If a proxy is rate-limited or fails, the tool rotates to the next one in your list.

This smart switching system ensures the checks are as fast as possible, that it works great even with average-speed proxies and an efficient use of your proxies and their bandwidth.

---

## 🔍 What Makes a Valid Username?

Snapchat requires usernames that:
- Are 3–15 characters long
- Start with a letter
- Contain only letters, numbers, dots, dashes or underscores
- Do **not** end with a symbol
- Do **not** contain multiple symbols in a row

> This tool filters invalid usernames before even sending a request — saving time, proxies and ratelimits.

---

## 👾 Support

Open an issue on [GitHub](https://github.com/ytax/snapchat-username-checker/issues).

---

## 🧪 Known Issues / Limitations

- Some SOCKS5 proxies may not support HTTP/2 and will be skipped.
- Running large lists with poor proxies may significantly slow down checks.
- Concurrency isn't implemented yet.
- The proxy validator is abysmally, concurrency needs to be implemented here asap.

---

## 🧱 Developer Notes

- The `suggest_username.proto` file was reverse engineered from the mobile app and defines the full request/response schema for the SuggestUsername gRPC call.
- You can replace this proto file with others if you reverse additional endpoints.
- The tool currently sends an empty locale (`""`) in gRPC requests. This matches Snapchat’s current default behavior but this could be a problem in the future.

---

## 🔮 Roadmap (V2 Preview)

- [ ] Cleaner, modular codebase
- [ ] Parallel checking with goroutines or some other form of concurrency
- [ ] Optional web GUI
- [ ] Cleaner UI and better user experience

---

## 🤝 Contributing

Pull requests are welcome! Feel free to fork the project and submit improvements, bug fixes, or feature ideas.

---

## ⚖️ Disclaimer

This tool is intended for research purposes only.  
Use responsibly.

---

## 🖼️ Screenshots

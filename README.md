# Pub Sub

[![Contribute with Gitpod](https://img.shields.io/badge/Contribute%20with-Gitpod-908a85?logo=gitpod)](https://gitpod.io/#https://github.com/Nojus0/pub-sub)

Scalable pub-sub server written in go. Can serve 1 million websocket connections while using about gigabyte of ram. Meant to be deployed on Linux.

# Set up
Build the project with `GOOS` set to linux and the architecture that you want to build for. And you **must** run the `setup.sh` script with `sudo` on the target server, if you don't run the script the amount of connections you can have will be significantly lower.

# Building
`go build -o pubsub` inside the `src` directory this will build the application.

# Developing on Windows
This application uses linux syscalls which are not available on windows so vscode shows that those specific functions dont exist, they exist all you need to do is set `GOOS` to linux `go env -w GOOS=linux` use this command to set it, reopen vscode and type `wsl` in the terminal this will open windows subsytem for linux this allows you to run linux application on windows without needing to create a virtual machine. If you haven't installed `wsl` you easily install it via `Windows Features`.



Project based of https://github.com/eranyanay/1m-go-websockets

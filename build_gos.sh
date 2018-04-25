basePath=$(pwd)/vendor

# Build AuthApp
export GOPATH=$basePath:$(pwd)/AuthApp:$(pwd)/GosLib
go install AuthApp

# Build ConnectApp
export GOPATH=$basePath:$(pwd)/ConnectApp:$(pwd)/GosLib:$(pwd)/GameApp
go install ConnectApp

# Build GameApp
export GOPATH=$basePath:$(pwd)/GameApp:$(pwd)/GosLib
go install GameApp

# Build WorldApp
export GOPATH=$basePath:$(pwd)/WorldApp:$(pwd)/GosLib
go install WorldApp

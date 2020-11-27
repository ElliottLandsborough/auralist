# Auralist

## Download

```bash
go get -v github.com/ElliottLandsborough/auralist
```

## Config

config.yml in root of project

```yaml
mysqlDatabase: "auralist"
mysqlHost: "localhost"
mysqlUser: "root"
mysqlPass: "password"
searchDirectory: "/home/username/Music"
```

## Setup

```
cd $GOPATH/src/github.com/ElliottLandsborough/auralist
make deps && make run
```

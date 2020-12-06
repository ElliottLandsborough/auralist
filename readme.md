# Auralist

## Download

```bash
go get -u -v github.com/ElliottLandsborough/auralist
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

## Remove useless files

```bash

find ./ -type f -name ".DS_Store" -delete;
find ./ -type f -name "._.DS_Store" -delete;
find ./ -type f -name "._*" -delete;
find ./ -type f -name ".Spotlight-V100" -delete;
find ./ -type f -name ".Trashes" -delete;
find ./ -type f -name "Thumbs.db" -delete;
find ./ -type f -name "desktop.ini" -delete;
find ./ -type f -name "thumbs.db" -delete;
find ./ -type f -name "ethumbs.db" -delete;
```

```sql
DELETE FROM `files` WHERE `file_name` = ".DS_Store";
DELETE FROM `files` WHERE `file_name` = "._.DS_Store";
DELETE FROM `files` WHERE `file_name` = ".Spotlight-V100";
DELETE FROM `files` WHERE `file_name` = ".Trashes";
DELETE FROM `files` WHERE `file_name` = "Thumbs.db";
DELETE FROM `files` WHERE `file_name` = "desktop.ini";
DELETE FROM `files` WHERE `file_name` = "thumbs.db";
DELETE FROM `files` WHERE `file_name` = "ethumbs.db";
```

```bash
find . -name __MACOSX;
```

```sql
DELETE FROM `files` WHERE `path` LIKE '%__MACOSX%'
```

```sql
SELECT `path` FROM `files` WHERE `path` LIKE "._%"
```

# ibex

ibex, running scripts on large scale machines

# compile

```
# export GOPROXY=https://goproxy.cn
make
```

# run server

```
mysql < sql/ibex.sql
./ibex server
```

# run agentd

```
./ibex agentd
```

# test

```
# create task
curl --location --request POST 'localhost:10090/ibex/v1/tasks' \
-u ibex:ibex \
--header 'Content-Type: application/json' \
--data-raw '{
    "title": "just a echo",
    "account": "root",
    "batch": 0,
    "tolerance": 0,
    "timeout": 10,
    "pause": "",
    "script": "#!/bin/sh\necho hello;date > nice.date;echo world",
    "action": "start",
    "creator": "qinxiaohui",
    "hosts": ["bogon"]
}'
```

# TODO

[] task done flag

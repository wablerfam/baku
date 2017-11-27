# baku
baku is file base api cron scheduler

[baku](https://github.com/wablerfam/baku/blob/master/image/baku.png)

## Download & Install
Download from github release

[download](https://github.com/wablerfam/baku/releases)

Installation is unnecessary.  
If you specify the config file and execute it, it will start immediately
## Setting
Need to create a TOML base config file
### Server Setting  
Set api port number for [server] section
```
[server]
port = 8900
```
### Database Setting  
Set database file for [Database] section
```
[database]
path = "./data/baku.db"
```
### Job Setting
#### Major Job Setting
Set Major job name and description for [job] section
```
[job]
name = "greet"
description = "greet aroud the world"
```
#### Job Group Setting
Set job group name and description for [[job.group]] section
```
[[job.group]]
name = "hello"
description = "greet type hello"
```
#### Job Group Task Setting
Set job group name, description, timing and command for [[job.group.task]] section
```
[[job.group.task]]
name = "bear"
description = "greet hello bear"
timing = "0 0 7 * * *"
command = "echo 'hello bear'"
```
#### Multiple Job Group and Task
Can set multiple Job groups and Tasks
```
[[job.group]]
name = "hello"
description = "greet type hello"

[[job.group.task]]
name = "bear"
description = "greet hello bear"
timing = "0 0 7 * * *"
command = "echo 'hello bear'"

[[job.group.task]]
name = "bird"
description = "greet hello bird"
timing = "0 0 7 * * *"
command = "echo 'hello bird'"

[[job.group]]
name = "goodbye"
description = "greet type goodbye"

[[job.group.task]]
name = "bear"
description = "greet hello bear"
timing = "0 0 21 * * *"
command = "echo 'hello bear'"
```
### Example Setting
Example Config file setting is here

[ExampleConfigFile](https://github.com/wablerfam/baku/blob/master/etc/baku.toml)
## Starting
Specify the config file with -c option
```
# baku -c etc/baku.toml
2017-11-26T18:46:24.625+0900	info	baku.main	start
2017-11-26T18:46:24.628+0900	info	baku.config	load etc/baku.toml
2017-11-26T18:46:24.645+0900	info	hello.bear	task initialize
2017-11-26T18:46:24.651+0900	info	hello.bird	task initialize
2017-11-26T18:46:24.657+0900	info	goodbye.bear	task initialize
2017-11-26T18:46:24.664+0900	info	baku.server	up 8900
```
## API
### Get Running Status
Check the running status of job task
```
# curl -G http://baku:8900/api/status | jq
{
  "job": {
    "name": "greet",
    "description": "greet aroud the world",
    "groups": [
      {
        "name": "hello",
        "description": "greet type hello",
        "tasks": [
          {
            "name": "bear",
            "description": "greet hello bear",
            "status": "succeeded",
            "exec_time": 0.005296595,
            "last_start_time": "2017-11-26 18:54:40.011386608 +0900 JST",
            "last_end_time": "2017-11-26 18:54:40.016682205 +0900 JST"
          }
        ]
      }
    ]
  }
}
```
#### status list  
|status|description|
----|----
|succeeded|job task succeeded state|
|ruuning|job task ruuning state|
|abended|job task abended state|
|aborted(running)|baku got down while job task was running|
|initialized|between initial registration and first execution|
|refreshed|abort or abend status refreshed|
### Immediate Exec Task
Execute the job task immediately via the API
```
# curl -X POST "http://localhost:8900/api/actions/exec?group=hello&task=bear" | jq
{
  "action": "exec",
  "groupname": "hello",
  "taskname": "bear",
  "result": "success",
  "message": ""
}
```
### Immediate Kill Task
Kill the job task immediately via the API
```
# curl -X POST "http://localhost:8900/api/actions/kill?group=hello&task=bear" | jq
{
  "action": "kill",
  "groupname": "hello",
  "taskname": "bear",
  "result": "success",
  "message": ""
}
```
### Refresh Task
Refresh the job task via the API
```
# curl -X POST "http://localhost:8900/api/actions/kill?group=hello&task=bear" | jq
{
  "action": "refresh",
  "groupname": "hello",
  "taskname": "bear",
  "result": "success",
  "message": ""
}
```
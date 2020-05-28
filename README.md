# Core

## Structure

```yaml
├── cmd # executive sources witch be compiled in binary
│   └── example1 # name of binary
│       └── main.go 
├── components # packages with business logic. should not have deps on other components.
│   └── example1 # name of component
│       └── *.go
├── libs # packages for common usage and low-middle level communication code. packages should not have any deps on packages from 'components'
│   ├── base # fx logic and external deps
│   │   └── *.go
│   ├── tester # helpers for testing
│   │   └── *.go
│   └── util # common usage code
│       └── *.go
├── misc # unsorted files
├── res
│   ├── cfg # config files (naming conventions below)
│   │   └── core-local.yml
│   └── migrations # migration files(naming conventions in golang-migrate repo)
│       ├── ch_dbname 
│       │   ├── 1_a.down.sql
│       │   ├── 1_a.up.sql
│       │   ├── 2_b.down.sql
│       │   └── 2_b.up.sql
│       └── pg_dbname
│           ├── 1_a.down.sql
│           ├── 1_a.up.sql
│           ├── 2_b.down.sql
│           └── 2_b.up.sql
├── scripts # scripts for code maintenance 
│   ├── build.sh
│   ├── lint.sh
│   └── test.sh
├── go.mod
├── go.sum
└── README.md
```

## Code maintenance
#### Dependencies
```bash
go mod tidy
```
#### Lint
```bash
bash scripts/lint.sh 
```
#### Test
```bash
bash scripts/test.sh 
```
#### Build
```bash
bash scripts/build.sh 
```

#### Format
###### JetBrains
- Enable checkbox "Reformat code" in commit dialog
- Enable checkbox "Optimize imports" in commit dialog
###### Other IDE
Use gofmt, then goimports

## Code style
###### Main guidance
<https://github.com/uber-go/guide/blob/master/style.md>  
<https://golang.org/doc/effective_go.html>  
<https://github.com/golang/go/wiki/CodeReviewComments>  

###### Code documentation
1. All package requires package level comment, more code and complexity in package more comments
2. All package in 'components' directory requires full documentation package level comment
3. All exported methods and structs in packages from 'components' directory requires mandatory comment  
4. Package in 'libs' directory should have comments on all unobvious exported functions. Any ambiguous func or structs name must have comments
5. Do not require comments functions and structs with obvious purpose. Setters, getters, dialers, constructors or other stuff.  
But only if this functions do not have any side effects. In that case comment is mandatory. 

6. Functions comment should begin with '// FuncName ...' as such as method,variables and types.   
Package comment should begin with '// Package packname ...'.    

7. Comment inside complex functions with many logic branches are appreciated.
8. Comment on exported unobvious global variables and constants are appreciated too.
9. Commit message should have short information about committed code. Example:
```
packname1: fix deadlock in examle.Start();
packname2: add validation for struct example2; 
```

###### Notes
<https://github.com/uber-go/guide/blob/master/style.md#reduce-scope-of-variables>  
Use inline err assertion only if there is no other way keep it simple. In other cases write two line. It improve code readability(clean).  
<https://github.com/uber-go/guide/blob/master/style.md#format-strings-outside-printf>  
If liner do not see you Printf functions setup linter(res/.golangci.yml)  


## Naming conventions

###### Source code files
sneak_case.go
###### Configs files
binaryName-environment.yml
###### Migration files
<https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md>

## Documentation for external libs  
##### DI
<https://github.com/uber-go/fx>  
##### Logger
<https://github.com/uber-go/zap>  
##### DB drivers
<https://github.com/jackc/pgx>  
<https://github.com/ClickHouse/clickhouse-go> 
<https://github.com/influxdata/influxdb>
##### Metrics
<https://github.com/uber-go/tally>
##### Migrations
<https://github.com/golang-migrate/migrate>  
##### Kafka
<https://github.com/segmentio/kafka-go>  
##### Tests
<https://github.com/stretchr/testify>  
<https://github.com/bouk/monkey>  
##### Linter
<https://github.com/golangci/golangci-lint>  
##### Utils
<https://github.com/jinzhu/now>  
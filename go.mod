module github.com/testbook/app-search-sync

go 1.16

require (
	github.com/BurntSushi/toml v1.0.0
	github.com/golang/snappy v0.0.3 // indirect
	github.com/rwynn/gtm v1.0.1-0.20191119151623-081995b34c9c
	github.com/serialx/hashring v0.0.0-20200727003509-22c0c7ab6b1b // indirect
	github.com/testbook/app-search-sync/client v1.0.0
	go.mongodb.org/mongo-driver v1.8.4
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

replace github.com/testbook/app-search-sync/client => ../app-search-sync/client

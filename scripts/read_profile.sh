# go tool pprof -edgefraction 0 -nodefraction 0 -nodecount 100000 ../test.prof
go tool pprof -http=localhost:$1 $2.prof

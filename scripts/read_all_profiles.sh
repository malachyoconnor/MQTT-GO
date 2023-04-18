go tool pprof -http=localhost:5000 TCP.prof &
go tool pprof -http=localhost:5001 UDP.prof &
go tool pprof -http=localhost:5002 QUIC.prof

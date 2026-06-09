module github.com/Lokee86/space-rocks/server

go 1.26.3

require github.com/gorilla/websocket v1.5.3 

require github.com/Lokee86/space-rocks/player-data v0.0.0

replace github.com/Lokee86/space-rocks/player-data => ../player-data

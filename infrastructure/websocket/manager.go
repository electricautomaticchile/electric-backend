package websocket

var GlobalHub *Hub

func InitializeHub() *Hub {
	GlobalHub = NewHub()
	go GlobalHub.Run()
	return GlobalHub
}

func GetHub() *Hub {
	return GlobalHub
}

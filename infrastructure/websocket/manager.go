package websocket

var GlobalHub *Hub

func InitializeHub() *Hub {
	GlobalHub = NewHub()
	return GlobalHub
}

func GetHub() *Hub {
	return GlobalHub
}

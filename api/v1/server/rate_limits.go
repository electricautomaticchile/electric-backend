package server

// rateLimits retorna la configuración de rate limiting por endpoint.
func rateLimits() map[string]int {
	return map[string]int{
		"POST:/api/auth/login":                  5,
		"POST:/api/auth/login/empresa":          5,
		"POST:/api/auth/registro-empresa":       3,
		"POST:/api/auth/solicitar-recuperacion": 3,
		"POST:/api/auth/restablecer-password":   3,
		"POST:/api/auth/cambiar-password":       10,
		"POST:/api/auth/refresh-token":          20,

		"GET:/api/clientes":        60,
		"POST:/api/clientes":       20,
		"PUT:/api/clientes/:id":    30,
		"DELETE:/api/clientes/:id": 10,
		"GET:/api/clientes/:id":    100,

		"GET:/api/dispositivos":          120,
		"POST:/api/dispositivos":         20,
		"PUT:/api/dispositivos/:id":      60,
		"POST:/api/dispositivos/lectura": 300,
		"POST:/api/dispositivos/control": 30,

		"GET:/api/alertas":        100,
		"POST:/api/alertas":       50,
		"PUT:/api/alertas/:id":    40,
		"DELETE:/api/alertas/:id": 20,

		"GET:/api/boletas":     60,
		"POST:/api/boletas":    10,
		"GET:/api/boletas/:id": 100,

		"GET:/api/tickets":                60,
		"POST:/api/tickets":               10,
		"POST:/api/tickets/:id/responder": 20,
		"PUT:/api/tickets/:id":            30,

		"GET:/api/dashboard/empresa": 120,
		"GET:/api/dashboard/cliente": 120,
		"GET:/api/estadisticas/*":    100,

		"GET:/api/reportes/*": 10,

		"GET:/api/mapa/dispositivos":  60,
		"GET:/api/mapa/ubicacion/:id": 100,
		"POST:/api/mapa/actualizar":   30,

		"GET:/api/tarifas":     30,
		"POST:/api/tarifas":    5,
		"PUT:/api/tarifas/:id": 10,

		"POST:/api/imagenes-perfil/upload": 10,
		"GET:/api/imagenes-perfil/:id":     100,
		"DELETE:/api/imagenes-perfil/:id":  10,

		"POST:/api/iot/lectura":           3000,
		"POST:/api/iot/comando-ejecutado": 500,
	}
}

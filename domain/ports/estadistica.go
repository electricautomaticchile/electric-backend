package ports

import (
"context"
)

type PortEstadistica interface {
	ObtenerConsumoCliente(ctx context.Context, clienteID string) (map[string]interface{}, error)
	ObtenerEstadisticasGlobales(ctx context.Context) (map[string]interface{}, error)
}

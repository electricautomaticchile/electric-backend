package eventbus

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Publisher publica eventos en Redis Pub/Sub para que el servicio
// websocket-electric los entregue a los clientes conectados.
//
// La API ya no habla directamente con el Hub: solo publica eventos. Si Redis
// no está disponible, las publicaciones se descartan silenciosamente (los
// eventos en tiempo real son best-effort y no deben bloquear la lógica REST).
type Publisher struct {
	client *redis.Client
}

// NewPublisher crea un publisher. client puede ser nil (modo no-op).
func NewPublisher(client *redis.Client) *Publisher {
	return &Publisher{client: client}
}

// PublishToCliente envía un mensaje a todos los clientes de un cliente.
func (p *Publisher) PublishToCliente(clienteID string, msg Message) {
	p.publish(ChannelEvents, Event{Scope: ScopeCliente, TargetID: clienteID, Message: msg})
}

// PublishToEmpresa envía un mensaje a todos los clientes de una empresa.
func (p *Publisher) PublishToEmpresa(empresaID string, msg Message) {
	p.publish(ChannelEvents, Event{Scope: ScopeEmpresa, TargetID: empresaID, Message: msg})
}

// PublishToAll hace broadcast global a todos los clientes conectados.
func (p *Publisher) PublishToAll(msg Message) {
	p.publish(ChannelBroadcast, Event{Scope: ScopeAll, Message: msg})
}

func (p *Publisher) publish(channel string, event Event) {
	if p == nil || p.client == nil {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("⚠️ eventbus: error serializando evento: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := p.client.Publish(ctx, channel, data).Err(); err != nil {
		log.Printf("⚠️ eventbus: error publicando en %s: %v", channel, err)
	}
}

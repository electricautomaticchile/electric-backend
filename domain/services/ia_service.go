package services

import (
	"context"
	"math"
	"sort"
)

type IAService struct{}

func NewIAService() *IAService {
	return &IAService{}
}

type BoletaAnalisisIA struct {
	Periodo        string
	ConsumoTotal   float64
	Monto          float64
	ConsumoPorHora []ConsumoPorHoraIA
}

type ConsumoPorHoraIA struct {
	Hora    int
	Consumo float64
}

type ResultadoAnalisis struct {
	Consejos        []string        `json:"consejos"`
	AhorroEstimado  float64         `json:"ahorroEstimado"`
	HorasPico       []int           `json:"horasPico"`
	Recomendaciones []Recomendacion `json:"recomendaciones"`
}

type Recomendacion struct {
	Titulo      string `json:"titulo"`
	Descripcion string `json:"descripcion"`
	Impacto     string `json:"impacto"`
}

func (s *IAService) AnalizarConsumo(ctx context.Context, boletas interface{}) (*ResultadoAnalisis, error) {
	boletasSlice, ok := boletas.([]interface{})
	if !ok {
		return s.generarAnalisisBasico(300), nil
	}

	if len(boletasSlice) < 3 {
		return s.generarAnalisisBasico(300), nil
	}

	var consumoTotal float64
	var consumos []float64
	horasPicoMap := make(map[int]float64)

	for _, b := range boletasSlice {
		boletaMap, ok := b.(map[string]interface{})
		if !ok {
			continue
		}

		if consumo, ok := boletaMap["consumoTotal"].(float64); ok {
			consumoTotal += consumo
			consumos = append(consumos, consumo)
		}

		if consumoPorHora, ok := boletaMap["consumoPorHora"].([]interface{}); ok {
			for _, ch := range consumoPorHora {
				if chMap, ok := ch.(map[string]interface{}); ok {
					hora := int(chMap["hora"].(float64))
					consumo := chMap["consumo"].(float64)
					horasPicoMap[hora] += consumo
				}
			}
		}
	}

	consumoPromedio := consumoTotal / float64(len(consumos))
	
	variacion := 0.0
	if len(consumos) > 1 {
		ultimoConsumo := consumos[len(consumos)-1]
		variacion = ((ultimoConsumo - consumoPromedio) / consumoPromedio) * 100
	}

	horasPico := s.identificarHorasPico(horasPicoMap)

	return s.generarAnalisisLocal(consumoPromedio, variacion, horasPico), nil
}

func (s *IAService) identificarHorasPico(horasPicoMap map[int]float64) []int {
	if len(horasPicoMap) == 0 {
		return []int{18, 19, 20, 21, 22}
	}

	type horaConsumo struct {
		hora    int
		consumo float64
	}

	var horas []horaConsumo
	var totalConsumo float64

	for hora, consumo := range horasPicoMap {
		horas = append(horas, horaConsumo{hora, consumo})
		totalConsumo += consumo
	}

	promedioConsumo := totalConsumo / float64(len(horas))
	umbralPico := promedioConsumo * 1.2

	var horasPico []int
	for _, hc := range horas {
		if hc.consumo > umbralPico {
			horasPico = append(horasPico, hc.hora)
		}
	}

	sort.Ints(horasPico)
	return horasPico
}

func (s *IAService) generarConsejos(consumoPromedio, variacion float64, horasPico []int) []string {
	consejos := []string{}

	if variacion > 10 {
		consejos = append(consejos, "Tu consumo ha aumentado un "+formatFloat(variacion)+"% respecto al promedio de los últimos meses")
	} else if variacion < -10 {
		consejos = append(consejos, "¡Excelente! Has reducido tu consumo en un "+formatFloat(math.Abs(variacion))+"% respecto al promedio")
	}

	if consumoPromedio > 400 {
		consejos = append(consejos, "Tu consumo mensual promedio de "+formatFloat(consumoPromedio)+" kWh está por encima del promedio nacional (250 kWh)")
	} else if consumoPromedio < 200 {
		consejos = append(consejos, "Tu consumo mensual promedio de "+formatFloat(consumoPromedio)+" kWh es eficiente comparado con el promedio nacional")
	}

	if len(horasPico) > 5 {
		consejos = append(consejos, "Se detectaron "+string(rune(len(horasPico)))+" horas con consumo elevado. Considera redistribuir el uso de electrodomésticos")
	}

	return consejos
}

func (s *IAService) generarRecomendaciones(consumoPromedio, variacion float64, horasPico []int) []Recomendacion {
	recomendaciones := []Recomendacion{}

	if len(horasPico) > 0 {
		recomendaciones = append(recomendaciones, Recomendacion{
			Titulo:      "Evitar horas pico de consumo",
			Descripcion: "Programa el uso de lavadora, secadora y lavavajillas fuera de las horas pico detectadas",
			Impacto:     "alto",
		})
	}

	if consumoPromedio > 300 {
		recomendaciones = append(recomendaciones, Recomendacion{
			Titulo:      "Optimizar temperatura de climatización",
			Descripcion: "Ajusta el termostato 2°C: 24°C en verano y 20°C en invierno puede ahorrar hasta 15%",
			Impacto:     "alto",
		})
	}

	recomendaciones = append(recomendaciones, Recomendacion{
		Titulo:      "Desconectar aparatos en standby",
		Descripcion: "TV, computadores y cargadores en standby consumen hasta 10% de tu energía mensual",
		Impacto:     "medio",
	})

	recomendaciones = append(recomendaciones, Recomendacion{
		Titulo:      "Usar iluminación LED",
		Descripcion: "Reemplaza ampolletas incandescentes por LED para ahorrar hasta 80% en iluminación",
		Impacto:     "medio",
	})

	if variacion > 15 {
		recomendaciones = append(recomendaciones, Recomendacion{
			Titulo:      "Revisar electrodomésticos antiguos",
			Descripcion: "Electrodomésticos con más de 10 años pueden consumir hasta 50% más energía",
			Impacto:     "alto",
		})
	}

	return recomendaciones
}

func (s *IAService) generarAnalisisLocal(consumoPromedio, variacion float64, horasPico []int) *ResultadoAnalisis {
	consejos := s.generarConsejos(consumoPromedio, variacion, horasPico)
	recomendaciones := s.generarRecomendaciones(consumoPromedio, variacion, horasPico)
	ahorroEstimado := s.calcularAhorroEstimado(consumoPromedio, horasPico)

	return &ResultadoAnalisis{
		Consejos:        consejos,
		AhorroEstimado:  ahorroEstimado,
		HorasPico:       horasPico,
		Recomendaciones: recomendaciones,
	}
}

func (s *IAService) calcularAhorroEstimado(consumoPromedio float64, horasPico []int) float64 {
	ahorroBase := consumoPromedio * 0.10

	if len(horasPico) > 5 {
		ahorroBase += consumoPromedio * 0.05
	}

	if consumoPromedio > 400 {
		ahorroBase += consumoPromedio * 0.05
	}

	return math.Round(ahorroBase*100) / 100
}

func (s *IAService) generarAnalisisBasico(consumoPromedio float64) *ResultadoAnalisis {
	return &ResultadoAnalisis{
		Consejos: []string{
			"Análisis básico generado con datos limitados",
			"Registra más boletas para obtener recomendaciones personalizadas con IA",
		},
		AhorroEstimado: consumoPromedio * 0.10,
		HorasPico:      []int{18, 19, 20, 21, 22},
		Recomendaciones: []Recomendacion{
			{
				Titulo:      "Desconectar aparatos en standby",
				Descripcion: "Los aparatos en modo standby pueden consumir hasta 10% de tu energía",
				Impacto:     "medio",
			},
		},
	}
}

func formatFloat(f float64) string {
	return string(rune(int(math.Round(f))))
}

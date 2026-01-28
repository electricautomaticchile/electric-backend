package services

import (
	"context"
	"electric-backend/config"
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type IAService struct {
	geminiClient *genai.Client
}

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

	if config.AppConfig.GeminiAPIKey == "" {
		return s.generarAnalisisLocal(consumoPromedio, variacion, horasPico), nil
	}

	resultado, err := s.analizarConGemini(ctx, consumoPromedio, variacion, horasPico, consumos)
	if err != nil {
		return s.generarAnalisisLocal(consumoPromedio, variacion, horasPico), nil
	}

	return resultado, nil
}

func (s *IAService) analizarConGemini(ctx context.Context, consumoPromedio, variacion float64, horasPico []int, consumos []float64) (*ResultadoAnalisis, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.AppConfig.GeminiAPIKey))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	model.SetTemperature(0.7)
	model.SetTopP(0.95)
	model.SetTopK(40)
	model.SetMaxOutputTokens(2048)

	prompt := fmt.Sprintf(`Eres un experto en eficiencia energética en Chile. Analiza estos datos de consumo eléctrico y genera recomendaciones específicas en español.

Datos del cliente:
- Consumo promedio mensual: %.2f kWh
- Variación respecto al promedio: %.2f%%
- Horas pico detectadas: %v
- Consumos de últimos meses: %v

Genera una respuesta en formato JSON con esta estructura exacta:
{
  "consejos": ["consejo1", "consejo2", "consejo3"],
  "ahorroEstimado": 50.5,
  "horasPico": [18, 19, 20, 21],
  "recomendaciones": [
    {
      "titulo": "Título de la recomendación",
      "descripcion": "Descripción detallada",
      "impacto": "alto"
    }
  ]
}

Reglas:
- Genera 3-5 consejos específicos y accionables
- El ahorro estimado debe ser en kWh/mes
- Las recomendaciones deben tener impacto: "alto", "medio" o "bajo"
- Considera el contexto chileno (tarifas, clima, horarios)
- Sé específico con números y porcentajes
- Responde SOLO con el JSON, sin texto adicional`, 
		consumoPromedio, variacion, horasPico, consumos)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("respuesta vacía de Gemini")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	
	responseText = cleanJSONResponse(responseText)

	var resultado ResultadoAnalisis
	if err := json.Unmarshal([]byte(responseText), &resultado); err != nil {
		return nil, fmt.Errorf("error parseando respuesta de Gemini: %w", err)
	}

	if len(resultado.HorasPico) == 0 {
		resultado.HorasPico = horasPico
	}

	return &resultado, nil
}

func cleanJSONResponse(text string) string {
	start := 0
	end := len(text)
	
	for i := 0; i < len(text); i++ {
		if text[i] == '{' {
			start = i
			break
		}
	}
	
	for i := len(text) - 1; i >= 0; i-- {
		if text[i] == '}' {
			end = i + 1
			break
		}
	}
	
	return text[start:end]
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

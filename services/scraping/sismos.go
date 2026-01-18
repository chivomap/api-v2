package scraping

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Estructura para almacenar los datos del sismo
type Sismo struct {
	Fecha        string  `json:"fecha"`
	Fases        int     `json:"fases"`
	Latitud      float64 `json:"latitud"`
	Longitud     float64 `json:"longitud"`
	Profundidad  float64 `json:"profundidad"`
	Magnitud     float64 `json:"magnitud"`
	Localizacion string  `json:"localizacion"`
	RMS          float64 `json:"rms"`
	Estado       string  `json:"estado"`
}

type eventoSignalR struct {
	GMTOT       string  `json:"gmtot"`
	Fases       int     `json:"fases"`
	Latitud     float64 `json:"latitud"`
	Longitud    float64 `json:"longitud"`
	Profundidad float64 `json:"profundidad"`
	M           float64 `json:"m"`
	Region      string  `json:"region"`
	RMS         float64 `json:"rms"`
	Estado      string  `json:"estado"`
}

type negotiateResponse struct {
	ConnectionID        string `json:"connectionId"`
	AvailableTransports []struct {
		Transport       string   `json:"transport"`
		TransferFormats []string `json:"transferFormats"`
	} `json:"availableTransports"`
}

type signalRMessage struct {
	Type      int               `json:"type"`
	Target    string            `json:"target"`
	Arguments []json.RawMessage `json:"arguments"`
}

func ScrapeSismos() ([]Sismo, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	// 1. Negociar conexión
	negResp, err := client.Post("https://srt.snet.gob.sv/rtsismos/seiscomphub/negotiate", "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("error negociando: %w", err)
	}
	defer negResp.Body.Close()

	var negotiate negotiateResponse
	if err := json.NewDecoder(negResp.Body).Decode(&negotiate); err != nil {
		return nil, fmt.Errorf("error decodificando negociación: %w", err)
	}

	connID := negotiate.ConnectionID

	// 2. Conectar a SSE
	sseURL := fmt.Sprintf("https://srt.snet.gob.sv/rtsismos/seiscomphub?id=%s", connID)
	req, _ := http.NewRequest("GET", sseURL, nil)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	sseResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error conectando SSE: %w", err)
	}
	defer sseResp.Body.Close()

	// 3. Enviar invoke SendEvento en goroutine
	go func() {
		time.Sleep(500 * time.Millisecond)
		invokeURL := fmt.Sprintf("https://srt.snet.gob.sv/rtsismos/seiscomphub?id=%s", connID)
		payload := `{"arguments":[],"target":"SendEvento","type":1}` + "\x1E"
		resp, err := http.Post(invokeURL, "text/plain;charset=UTF-8", bytes.NewBufferString(payload))
		if err != nil {
			fmt.Printf("Error invocando: %v\n", err)
		} else {
			resp.Body.Close()
			fmt.Printf("Invoke enviado correctamente\n")
		}
	}()

	// 4. Leer eventos SSE
	reader := bufio.NewReader(sseResp.Body)
	timeout := time.After(20 * time.Second)
	lineCount := 0

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout esperando eventos (leídas %d líneas)", lineCount)
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("conexión cerrada después de %d líneas", lineCount)
				}
				return nil, fmt.Errorf("error leyendo: %w", err)
			}

			lineCount++
			line = strings.TrimSpace(line)
			
			if line == "" || line == ":" {
				continue
			}

			fmt.Printf("SSE[%d]: %s\n", lineCount, line)

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "" || data == "{}" {
				continue
			}

			// Parsear mensaje SignalR
			var msg signalRMessage
			if err := json.Unmarshal([]byte(data), &msg); err != nil {
				fmt.Printf("Error parseando mensaje: %v\n", err)
				continue
			}

			fmt.Printf("Mensaje tipo %d, target: %s\n", msg.Type, msg.Target)

			// Buscar EventSignal
			if msg.Target == "EventSignal" && len(msg.Arguments) > 0 {
				var eventos []eventoSignalR
				if err := json.Unmarshal(msg.Arguments[0], &eventos); err != nil {
					continue
				}

				// Convertir a Sismo
				result := make([]Sismo, 0, len(eventos))
				for _, evt := range eventos {
					t, _ := time.Parse(time.RFC3339, evt.GMTOT+"Z")
					result = append(result, Sismo{
						Fecha:        t.Format("2/1/2006, 3:04:05 p. m."),
						Fases:        evt.Fases,
						Latitud:      evt.Latitud,
						Longitud:     evt.Longitud,
						Profundidad:  evt.Profundidad,
						Magnitud:     evt.M,
						Localizacion: "Localizado " + evt.Region,
						RMS:          evt.RMS,
						Estado:       evt.Estado,
					})
				}
				return result, nil
			}
		}
	}
}

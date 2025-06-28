package scraping

import (
	"fmt"
	"os"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Estructura para almacenar los datos del sismo
type Sismo struct {
	Fecha        string `json:"fecha"`
	Fases        string `json:"fases"`
	Latitud      string `json:"latitud"`
	Longitud     string `json:"longitud"`
	Profundidad  string `json:"profundidad"`
	Magnitud     string `json:"magnitud"`
	Localizacion string `json:"localizacion"`
	RMS          string `json:"rms"`
	Estado       string `json:"estado"`
}

// Función para scrapear los sismos
func ScrapeSismos() ([]Sismo, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("error al iniciar Playwright: %w", err)
	}
	defer pw.Stop()

	// Lanzar navegador
	// Usar headless mode basado en variable de entorno (por defecto true para producción)
	headless := os.Getenv("BROWSER_HEADLESS") != "false"
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
		Timeout:  playwright.Float(60000),
	})
	if err != nil {
		return nil, fmt.Errorf("error al lanzar navegador: %w", err)
	}
	defer browser.Close()

	// Crear contexto con user-agent
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	})
	if err != nil {
		return nil, fmt.Errorf("error al crear contexto del navegador: %w", err)
	}

	// Abrir nueva página
	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("error al abrir página: %w", err)
	}
	defer page.Close()

	// Configurar timeout de navegación
	page.SetDefaultTimeout(60000)
	page.SetDefaultNavigationTimeout(60000)

	// Navegar a la página
	_, err = page.Goto("https://srt.snet.gob.sv/rtsismos", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(60000),
	})
	if err != nil {
		return nil, fmt.Errorf("error al navegar a la página de sismos: %w", err)
	}

	// Esperar a que la tabla aparezca
	_, err = page.WaitForSelector("#tblEventos", playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		return nil, fmt.Errorf("error al esperar la tabla de sismos: %w", err)
	}

	// Esperar un poco más para asegurar que los datos se carguen
	time.Sleep(5 * time.Second)

	// Extraer los datos de la tabla
	sismos, err := page.Evaluate(`() => {
		const rows = document.querySelectorAll("#tblEventos tbody tr:not(:first-child)");
		return Array.from(rows).map(row => {
			const cells = row.querySelectorAll("td");
			return {
				fecha: cells[0]?.textContent?.trim() || "",
				fases: cells[1]?.textContent?.trim() || "",
				latitud: cells[2]?.textContent?.trim() || "",
				longitud: cells[3]?.textContent?.trim() || "",
				profundidad: cells[4]?.textContent?.trim() || "",
				magnitud: cells[5]?.textContent?.trim() || "",
				localizacion: cells[6]?.textContent?.trim() || "",
				rms: cells[7]?.textContent?.trim() || "",
				estado: cells[8]?.textContent?.trim() || ""
			};
		});
	}`)
	if err != nil {
		return nil, fmt.Errorf("error al extraer datos de la tabla: %w", err)
	}

	// Convertir los datos a una estructura de Go
	var result []Sismo
	for _, row := range sismos.([]interface{}) {
		data := row.(map[string]interface{})
		result = append(result, Sismo{
			Fecha:        fmt.Sprintf("%v", data["fecha"]),
			Fases:        fmt.Sprintf("%v", data["fases"]),
			Latitud:      fmt.Sprintf("%v", data["latitud"]),
			Longitud:     fmt.Sprintf("%v", data["longitud"]),
			Profundidad:  fmt.Sprintf("%v", data["profundidad"]),
			Magnitud:     fmt.Sprintf("%v", data["magnitud"]),
			Localizacion: fmt.Sprintf("%v", data["localizacion"]),
			RMS:          fmt.Sprintf("%v", data["rms"]),
			Estado:       fmt.Sprintf("%v", data["estado"]),
		})
	}

	return result, nil
}

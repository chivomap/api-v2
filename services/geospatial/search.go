package geospatial

import (
	"fmt"

	"chivomap.com/interfaces"
	"chivomap.com/types"
)



// mapToSlice convierte las claves de un mapa a un slice de strings.
func mapToSlice(m map[string]struct{}) []string {
	slice := make([]string, 0, len(m))
	for k := range m {
		slice = append(slice, k)
	}
	return slice
}

// GetMunicipios filtra las features por el valor exacto en la propiedad especificada ("D", "M" o "NAM").
func GetMunicipios(staticCache interfaces.StaticCacheService, query, whatIs string) (*types.GeoFeatureCollection, error) {
	if whatIs != "D" && whatIs != "M" && whatIs != "NAM" {
		return nil, fmt.Errorf("parámetro whatIs inválido '%s': debe ser 'M', 'D' o 'NAM'", whatIs)
	}
	
	// Usar cache estático en lugar de leer desde disco
	geo, err := staticCache.GetGeoData()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo datos geoespaciales para filtro %s=%s: %w", whatIs, query, err)
	}
	// Preallocar slice para mejor performance
	filteredFeatures := make([]types.GeoFeature, 0, len(geo.Features)/10) // Estimación conservadora
	for _, feat := range geo.Features {
		if feat.Properties == nil {
			continue
		}
		propVal, ok := feat.Properties[whatIs].(string)
		if !ok {
			continue
		}
		if propVal == query {
			filteredFeatures = append(filteredFeatures, feat)
		}
	}
	return &types.GeoFeatureCollection{
		Type:     "FeatureCollection",
		Features: filteredFeatures,
	}, nil
}

// GetGeoData extrae nombres únicos de departamentos, municipios y distritos a partir del TopoJSON.
func GetGeoData(staticCache interfaces.StaticCacheService) (*types.GeoData, error) {
	// Usar cache estático en lugar de leer desde disco
	geo, err := staticCache.GetGeoData()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo datos geoespaciales para extracción de nombres: %w", err)
	}
	// Preallocar maps con capacidad estimada para mejor performance
	estimatedSize := len(geo.Features) / 50 // Estimación basada en datos de El Salvador
	departamentos := make(map[string]struct{}, estimatedSize)
	municipios := make(map[string]struct{}, estimatedSize*5)
	distritos := make(map[string]struct{}, estimatedSize*20)
	for _, feat := range geo.Features {
		if feat.Properties == nil {
			continue
		}
		if d, ok := feat.Properties["D"].(string); ok && d != "" {
			departamentos[d] = struct{}{}
		}
		if m, ok := feat.Properties["M"].(string); ok && m != "" {
			municipios[m] = struct{}{}
		}
		if nam, ok := feat.Properties["NAM"].(string); ok && nam != "" {
			distritos[nam] = struct{}{}
		}
	}
	return &types.GeoData{
		Departamentos: mapToSlice(departamentos),
		Municipios:    mapToSlice(municipios),
		Distritos:     mapToSlice(distritos),
	}, nil
}

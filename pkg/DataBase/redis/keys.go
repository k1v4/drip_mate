package redispkg

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	catalogItemKey = "catalog:item:%s" // catalog:item:{uuid}
	userProfileKey = "profile:user:%s" // profile:user:{uuid}
	weatherCityKey = "weather:city:%s" // weather:city:{city}
)

func GetCatalogItemKey(id uuid.UUID) string {
	return fmt.Sprintf(catalogItemKey, id.String())
}

func GetUserProfileKey(id uuid.UUID) string {
	return fmt.Sprintf(userProfileKey, id.String())
}

func GetWeatherCityKey(city string) string {
	return fmt.Sprintf(weatherCityKey, city)
}

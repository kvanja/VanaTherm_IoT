package openweatherdata

type OpenWeatherData struct {
	Cord       Coord     `json:"coord"`
	Weather    []Weather `json:"weather"`
	Base       string    `json:"base"`
	Main       Main      `json:"main"`
	Visibility int       `json:"visibility"`
	Wind       Wind      `json:"wind"`
	Clouds     Clouds    `json:"clouds"`
	Dt         int       `json:"dt"`
	Sys        Sys       `json:"sys"`
	Timezone   int       `json:"timezone"`
	Id         int       `json:"id"`
	Name       string    `json:"name"`
	Cod        int       `json:"cod"`
}

type DataForApp struct {
	WeatherStatus string  `json:"weatherStatus"`
	Icon          string  `json:"icon"`
	Temperature   float32 `json:"temp"`
}

type Main struct {
	Temperature float32 `json:"temp"`
	FeelsLike   float32 `json:"feels_like"`
	TempMin     float32 `json:"temp_min"`
	TempMax     float32 `json:"temp_max"`
	Pressure    float32 `json:"pressure"`
	Humidity    float32 `json:"humidity"`
	SeaLevel    float32 `json:"sea_level"`
	GrndLevel   float32 `json:"grnd_level"`
}

type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Coord struct {
	Lon float32 `json:"lon"`
	Lat float32 `json:"lat"`
}

type Wind struct {
	Speed float32 `json:"speed"`
	Deg   int     `json:"deg"`
	Gust  float32 `json:"gust"`
}

type Clouds struct {
	All int `json:"all"`
}

type Sys struct {
	Type    int    `json:"type"`
	Id      int    `json:"id"`
	Country string `json:"country"`
	Sunrise int    `json:"sunrise"`
	Sunset  int    `json:"sunset"`
}

type City struct {
	City string `json:"city"`
}

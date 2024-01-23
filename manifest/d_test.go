package main_test

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"testing"
	"wxbot/internal/dto"
)

var ctx = context.Background()

func TestName(t *testing.T) {
	client := g.Client()

	resp, err := client.Get(ctx, "https://restapi.amap.com/v3/weather/weatherInfo?city=440981&key=901c4be03b0e1776938ee5a555dc7166")
	if err != nil {
		return
	}
	defer resp.Close()
	weatherRes := &dto.WeatherRes{}
	gconv.Scan(resp.ReadAllString(), weatherRes)

	if weatherRes.Info == "OK" {
		live := weatherRes.Lives[0]
		g.Dump(fmt.Sprintf("地点: %v\n  天气: %v\n 气温: %v\n 风向: %v\n 风力: %v\n 湿度: %v ", live.Province+live.City, live.Weather, live.Temperature, live.Winddirection, live.Windpower, live.Humidity))
	}
	g.Dump(weatherRes)

}

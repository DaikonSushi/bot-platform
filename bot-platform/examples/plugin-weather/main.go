// Example weather plugin - demonstrates how to create an external plugin
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/DaikonSushi/bot-platform/pkg/pluginsdk"
)

// WeatherPlugin is an example external plugin
type WeatherPlugin struct {
	bot *pluginsdk.BotClient
}

func (p *WeatherPlugin) Info() pluginsdk.PluginInfo {
	return pluginsdk.PluginInfo{
		Name:              "weather",
		Version:           "1.0.0",
		Description:       "Get weather information for cities",
		Author:            "YourName",
		Commands:          []string{"weather", "天气"},
		HandleAllMessages: false,
	}
}

func (p *WeatherPlugin) OnStart(bot *pluginsdk.BotClient) error {
	p.bot = bot
	bot.Log("info", "Weather plugin started")
	return nil
}

func (p *WeatherPlugin) OnStop() error {
	return nil
}

func (p *WeatherPlugin) OnMessage(ctx context.Context, bot *pluginsdk.BotClient, msg *pluginsdk.Message) bool {
	// This plugin doesn't handle general messages
	return false
}

func (p *WeatherPlugin) OnCommand(ctx context.Context, bot *pluginsdk.BotClient, cmd string, args []string, msg *pluginsdk.Message) bool {
	if len(args) == 0 {
		bot.Reply(msg, pluginsdk.Text("用法: /weather <城市名>\n例如: /weather 北京"))
		return true
	}

	city := strings.Join(args, " ")
	weather, err := getWeather(city)
	if err != nil {
		bot.Reply(msg, pluginsdk.Text(fmt.Sprintf("获取天气失败: %v", err)))
		return true
	}

	bot.Reply(msg, pluginsdk.Text(weather))
	return true
}

// getWeather fetches weather info (using a free weather API as example)
func getWeather(city string) (string, error) {
	// Using wttr.in free weather API
	url := fmt.Sprintf("https://wttr.in/%s?format=j1", city)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var data struct {
		CurrentCondition []struct {
			TempC       string `json:"temp_C"`
			TempF       string `json:"temp_F"`
			Humidity    string `json:"humidity"`
			WeatherDesc []struct {
				Value string `json:"value"`
			} `json:"weatherDesc"`
			WindspeedKmph string `json:"windspeedKmph"`
		} `json:"current_condition"`
		NearestArea []struct {
			AreaName []struct {
				Value string `json:"value"`
			} `json:"areaName"`
			Country []struct {
				Value string `json:"value"`
			} `json:"country"`
		} `json:"nearest_area"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if len(data.CurrentCondition) == 0 {
		return "", fmt.Errorf("no weather data found")
	}

	current := data.CurrentCondition[0]
	location := city
	if len(data.NearestArea) > 0 && len(data.NearestArea[0].AreaName) > 0 {
		location = data.NearestArea[0].AreaName[0].Value
		if len(data.NearestArea[0].Country) > 0 {
			location += ", " + data.NearestArea[0].Country[0].Value
		}
	}

	desc := "Unknown"
	if len(current.WeatherDesc) > 0 {
		desc = current.WeatherDesc[0].Value
	}

	return fmt.Sprintf(
		"🌍 %s 天气\n"+
			"━━━━━━━━━━━━\n"+
			"🌡️ 温度: %s°C (%s°F)\n"+
			"💧 湿度: %s%%\n"+
			"💨 风速: %s km/h\n"+
			"☁️ 状况: %s",
		location,
		current.TempC, current.TempF,
		current.Humidity,
		current.WindspeedKmph,
		desc,
	), nil
}

func main() {
	pluginsdk.Run(&WeatherPlugin{})
}

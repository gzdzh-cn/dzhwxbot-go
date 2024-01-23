package dto

type BuiltData struct {
	GfVersion    *string `json:"gfVersion"`
	GoVersion    *string `json:"goVersion"`
	BuiltGit     *string `json:"builtGit"`
	BuiltTime    *string `json:"builtTime"`
	BuiltVersion *string `json:"builtVersion"`
}

type WxBotCfg struct {
	Storage string `json:"storage"`
}

type ChatGptCfg struct {
	Mode          string `json:"mode"`
	RequestUrl    string `json:"requestUrl"`
	AccessToken   string `json:"accessToken"`
	Goframe       string `json:"goframe"`
	KeyWordPrefix string `json:"keyWordPrefix"`
}

type ChatReq struct {
	Action          string    `json:"action"`
	Messages        []Message `json:"messages"`
	ConversationId  *string   `json:"conversation_id"`
	ParentMessageId string    `json:"parent_message_id"`
	Model           string    `json:"model"`
}

type Message struct {
	Id      string  `json:"id"`
	Author  Author  `json:"author"`
	Content Content `json:"content"`
}

type Author struct {
	Role string `json:"role"`
}

type Content struct {
	ContentType string        `json:"content_type"`
	Parts       []interface{} `json:"parts"`
}

type ChatRes struct {
	Message struct {
		Id      string `json:"id"`
		Content struct {
			Parts []string `json:"parts"`
		} `json:"content"`
		Metadata struct {
			ParentId string `json:"parent_id"`
		}
	} `json:"message"`

	ConversationId *string `json:"conversation_id"`
}

type ChatHistory struct {
	Id              string  `json:"id"`
	ConversationId  *string `json:"conversation_id"`
	ParentMessageId string  `json:"parent_message_id"`
	*WeatherPosition
}

type WeatherPosition struct {
	SubscribeStatus bool   `json:"subscribeStatus"`
	City            string `json:"city"`
	Adcode          string `json:"adcode"`
}

type WeatherRes struct {
	Info     string `json:"info"`
	Infocode string `json:"infocode"`
	Lives    []struct {
		Weather          string `json:"weather"`
		Temperature      string `json:"temperature"`
		HumidityFloat    string `json:"humidity_float"`
		Province         string `json:"province"`
		City             string `json:"city"`
		Adcode           string `json:"adcode"`
		Winddirection    string `json:"winddirection"`
		Windpower        string `json:"windpower"`
		Humidity         string `json:"humidity"`
		Reporttime       string `json:"reporttime"`
		TemperatureFloat string `json:"temperature_float"`
	} `json:"lives"`
	Status string `json:"status"`
	Count  string `json:"count"`
}

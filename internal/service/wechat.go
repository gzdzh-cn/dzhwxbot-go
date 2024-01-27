package service

import (
	"context"
	"fmt"
	"github.com/go-ego/gse"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/skip2/go-qrcode"
	"log"
	"runtime"
	"strings"
	"time"
	"wxbot/internal/dto"
	"wxbot/internal/utils"

	"github.com/eatmoreapple/openwechat"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/util/grand"
)

// MsgHandler description
type MsgHandler struct {
	bot *openwechat.Bot
}

var (
	times = 0
	msgs  = []string{"hi", "你好", "余额", "收入", "支出"}
	seg   gse.Segmenter
)

// NewHandler description
func NewHandler(bot *openwechat.Bot) *MsgHandler {
	return &MsgHandler{
		bot: bot,
	}
}

// QrCodeCallBack 登录扫码回调，
func (h *MsgHandler) QrCodeCallBack(uuid string) {
	EmailDataSetting.SendEMail(GetQrcodeMsg("https://login.weixin.qq.com/l/"+uuid), "微信登录二维码", nil)
	if runtime.GOOS == "windows" {
		// 运行在Windows系统上
		openwechat.PrintlnQrcodeUrl(uuid)
	} else {
		glog.Debugf(context.Background(), "login in linux")
		q, _ := qrcode.New("https://login.weixin.qq.com/l/"+uuid, qrcode.Low)
		glog.Debugf(context.Background(), q.ToString(true))
	}
}

// KeepAlive 保活
func (h *MsgHandler) KeepAlive(ctx context.Context) {
	self, _ := bot.GetCurrentUser()
	//self, err := h.bot.GetCurrentUser()
	//if err != nil {
	//	glog.Errorf(ctx, "get current user error : %v", err)
	//	return
	//}
	glog.Debugf(ctx, "self : %+v", *self.User)
	mp, err := self.Mps(false)
	if err != nil {
		glog.Errorf(ctx, "get friends error : %v", err)
		return
	}

	now := time.Now()
	if times == 0 || (times == 24 && now.Hour() == 0) {
		times = now.Hour()
	}
	glog.Debugf(ctx, "保活: %d", times)
	if now.Hour() == times {
		glog.Debugf(ctx, "now : %+v", now)
		if chat := mp.GetByNickName("微信支付"); chat != nil {

			chat.SendText(msgs[grand.N(0, len(msgs)-1)])
		} else {
			glog.Errorf(ctx, "注意，没有找到保活对象")
			fs, err := self.Friends(true)
			if err != nil {
				glog.Errorf(ctx, "get friends error : %v", err)
				return
			}
			// 实在不行就改一个好友备注为ping，然后发送
			if chat := fs.GetByRemarkName("ping"); chat != nil {
				chat.SendText("余额")
			}
		}
		times++
	}
}

// 心跳
func (h *MsgHandler) SyncCheckCallback(resp openwechat.SyncCheckResponse) {
	ctx := gctx.New()
	glog.Debugf(ctx, "RetCode:%s  Selector:%s", resp.RetCode, resp.Selector)

}

// Handler 全局处理入口
func (h *MsgHandler) Handler(msg *openwechat.Message) {
	ctx := gctx.New()
	glog.Debugf(ctx, "hadler Received 消息类型 :%s", msg.MsgType)

	// 处理群消息
	if msg.IsSendByGroup() {
		h.GroupMsg(ctx, msg)
		return
	}

	// 好友申请
	if msg.IsFriendAdd() {

		_, err := msg.Agree("你好！我是AI引擎开发的微信机器人，你可以向我提问任何问题。")
		if err != nil {
			glog.Errorf(ctx, "add friend agree error : %v", err)
			return
		}
	}

	// 私聊
	h.UserMsg(ctx, msg)
}

func (h *MsgHandler) GroupMsg(ctx context.Context, msg *openwechat.Message) error {

	requestText := msg.Content
	//拍一拍
	if msg.IsTickledMe() {
		glog.Debug(ctx, "拍一拍 requestText", requestText)
		reply := "@" + utils.SubStr(requestText) + " \n\n" + "您好！请问拍我，是需要什么帮助吗"
		return ReplyMsg(ctx, msg, reply)
	}

	return ReadGroupMsg(ctx, msg, requestText)

}

func (h *MsgHandler) UserMsg(ctx context.Context, msg *openwechat.Message) error {

	switch msg.MsgType {
	case openwechat.MsgTypeText:
		return ReadPersonMsg(ctx, msg, msg.Content)
	}

	return nil
}

// 批量推送天气到群
func (h *MsgHandler) MultGetWeaher(ctx context.Context) {
	self, _ := bot.GetCurrentUser()
	groups, _ := self.Groups()
	glog.Debug(ctx, "查询全部群组")
	var groupsSlice []dto.WeatherPosition
	path := wxBotCfg.Storage + "/weather"
	list, _ := gfile.ScanDirFile(path, "*.json", true)
	//拿到开启订阅的群
	for _, v := range list {
		fmt.Println(gfile.GetContents(v))
		load, _ := gjson.LoadJson(gfile.GetContents(v))
		subscribeStatus := load.Get("subscribeStatus").Bool()
		if subscribeStatus {
			weatherPosition := dto.WeatherPosition{}
			err := load.Scan(&weatherPosition)
			if err != nil {
				glog.Error(ctx, err.Error())
			}
			groupsSlice = append(groupsSlice, weatherPosition)
		}
	}
	glog.Debug(ctx, "groupsSlice", groupsSlice)
	for _, v := range groupsSlice {
		result := groups.SearchByID(v.GroupsId)
		reply := GetWeather(ctx, v.Adcode)
		_ = result.SendText(reply)
	}
}

// 批量推送群通知
func (h *MsgHandler) MultNotice(ctx context.Context) {
	self, _ := bot.GetCurrentUser()
	groups, _ := self.Groups()
	glog.Debug(ctx, "查询全部群组")
	var groupsSlice []dto.WeatherPosition
	path := wxBotCfg.Storage + "/weather"
	list, _ := gfile.ScanDirFile(path, "*.json", true)
	for _, v := range list {
		fmt.Println(gfile.GetContents(v))
		load, _ := gjson.LoadJson(gfile.GetContents(v))
		subscribeStatus := load.Get("subscribeStatus").Bool()
		if subscribeStatus {
			weatherPosition := dto.WeatherPosition{}
			err := load.Scan(&weatherPosition)
			if err != nil {
				glog.Error(ctx, err.Error())
			}
			groupsSlice = append(groupsSlice, weatherPosition)
		}
		//groupsSlice = append(groupsSlice, weatherPosition)
	}
	glog.Debug(ctx, "groupsSlice", gconv.String(groupsSlice))

	results := groups.Search(len(groupsSlice), func(group *openwechat.Group) bool {
		for _, v := range groupsSlice {
			if group.ID() == v.GroupsId {
				return true
			}
		}

		return false
	})
	_ = results.SendText("订阅")
	glog.Debug(ctx, "results", gconv.String(results))
}

// 天气请求
func GetWeather(ctx context.Context, code string) string {
	glog.Warning(ctx, "查询天气")

	resp, err := client.Get(ctx, "https://restapi.amap.com/v3/weather/weatherInfo?city="+code+"&key=901c4be03b0e1776938ee5a555dc7166")
	if err != nil {
		glog.Debugf(ctx, "天气查询失败: %v", err.Error())
		return ""
	}
	defer resp.Close()
	weatherRes := &dto.WeatherRes{}
	gconv.Scan(resp.ReadAllString(), weatherRes)
	timeDate := gtime.New(time.Now())
	if weatherRes.Info == "OK" {
		live := weatherRes.Lives[0]
		Date := fmt.Sprintf("%v年%v月%v日", timeDate.Year(), timeDate.Month(), timeDate.Day())
		data := fmt.Sprintf("%v\n地点: %v\n天气: %v\n气温: %v\n风向: %v\n风力: %v\n湿度: %v\n查询时间:%v\n", Date, live.Province+live.City, live.Weather, live.Temperature+"摄氏度", live.Winddirection, live.Windpower+"级", live.Humidity, gtime.Now())
		glog.Debug(ctx, gjson.MustEncodeString(data))
		return data
	}
	return ""
}

// 单个对话框
func ReadPersonMsg(ctx context.Context, msg *openwechat.Message, requestText string) error {

	reply := ""
	requestText = strings.TrimSpace(requestText)
	newSegment, _ := gse.New("zh", "alpha")
	// 加载默认词典
	err1 := seg.LoadDict()
	if err1 != nil {
		glog.Error(ctx, "词典加载失败")
	}

	//分词
	slice := newSegment.Cut(requestText, true)
	glog.Debug(ctx, gconv.String(slice))
	if gstr.Contains(gconv.String(slice), "天气") && (gstr.Contains(gconv.String(slice), "查") || gstr.Contains(gconv.String(slice), "了解") || gstr.Contains(gconv.String(slice), "知道") || gstr.Contains(gconv.String(slice), "获取")) {
		//查询天气
		reply = FindWeather(ctx, slice)

	} else if gstr.Contains(gconv.String(slice), "天气") && (gstr.Contains(gconv.String(slice), "订阅")) {
		handler := NewHandler(bot)
		handler.MultGetWeaher(ctx)

		//self, _ := bot.GetCurrentUser()

		//for _, v := range groupsSlice {
		//	result := groups.SearchByID(v.GroupsId)
		//	r := GetWeather(ctx, v.Adcode)
		//	_ = result.SendText(r)
		//}

		//订阅天气
		reply = "不支持个人订阅"
	} else {
		reply = NewChatGpt().SendConent(ctx, msg, "请用中文回答："+requestText)
	}

	return ReplyMsg(ctx, msg, reply)
}

// 处理群聊文本
func ReadGroupMsg(ctx context.Context, msg *openwechat.Message, requestText string) error {

	requestText = strings.TrimSpace(msg.Content)
	requestText = strings.Trim(requestText, "\n")
	glog.Debug(ctx, "requestText", requestText)
	if requestText != "" {
		glog.Debugf(ctx, "群聊 requestText:%s", requestText)
		//群聊 @消息
		if msg.IsAt() {
			receiver, _ := msg.Receiver()
			requestText = gstr.SubStrFromEx(requestText, receiver.NickName)
			requestText = strings.TrimSpace(requestText)
			glog.Debugf(ctx, "群聊 @消息 截取后 requestText:%s", requestText)
			reply := ""
			newSegment, _ := gse.New("zh", "alpha")
			// 加载默认词典
			err1 := seg.LoadDict()
			if err1 != nil {
				glog.Error(ctx, "词典加载失败")
			}

			//分词
			slice := newSegment.Cut(requestText, true)
			glog.Debug(ctx, gconv.String(slice))

			//查询天气
			if gstr.Contains(gconv.String(slice), "天气") && (gstr.Contains(gconv.String(slice), "查") || gstr.Contains(gconv.String(slice), "了解") || gstr.Contains(gconv.String(slice), "知道") || gstr.Contains(gconv.String(slice), "获取")) {
				reply = FindWeather(ctx, slice)
			} else if gstr.Contains(gconv.String(slice), "天气") && (gstr.Contains(gconv.String(slice), "订阅")) {
				//订阅天气
				reply = SubscribeWeather(ctx, msg, slice)

			} else {
				reply = NewChatGpt().SendConent(ctx, msg, "请用中文回答："+requestText)
			}

			return ReplyMsg(ctx, msg, reply)
		} else {
			//群聊 关键词触发消息
			if strings.HasPrefix(requestText, chatGptCfg.Goframe+" ") {
				requestText = strings.TrimPrefix(requestText, chatGptCfg.Goframe+" ")
				glog.Debugf(ctx, "群聊 关键词 gf 截取后 requestText:%s", requestText)
				reply := Search(ctx, requestText)
				return ReplyMsg(ctx, msg, reply)
			} else if strings.HasPrefix(requestText, chatGptCfg.KeyWordPrefix+" ") || strings.HasPrefix(requestText, chatGptCfg.KeyWordPrefix+"\n") {
				requestText = strings.TrimSpace(gstr.SubStrFromEx(requestText, chatGptCfg.KeyWordPrefix))
				glog.Debugf(ctx, "群聊 关键词 小智 截取后 requestText:%s", requestText)
				reply := ""

				newSegment, _ := gse.New("zh", "alpha")
				// 加载默认词典
				err1 := seg.LoadDict()
				if err1 != nil {
					glog.Error(ctx, "词典加载失败")
				}

				//分词
				slice := newSegment.Cut(requestText, true)
				glog.Debug(ctx, gconv.String(slice))
				if gstr.Contains(gconv.String(slice), "天气") && (gstr.Contains(gconv.String(slice), "查") || gstr.Contains(gconv.String(slice), "了解") || gstr.Contains(gconv.String(slice), "知道") || gstr.Contains(gconv.String(slice), "获取")) {
					reply = FindWeather(ctx, slice)

				} else {
					reply = NewChatGpt().SendConent(ctx, msg, "请用中文回答："+requestText)
				}

				return ReplyMsg(ctx, msg, reply)
			}
		}
	}

	return nil
}

// 查询天气
func FindWeather(ctx context.Context, slice g.SliceStr) string {

	weatherCode, err := FilterWeather(ctx, slice)
	if err != nil {
		return err.Error()
	}

	return GetWeather(ctx, weatherCode.Adcode)
}

// 订阅天气
func SubscribeWeather(ctx context.Context, msg *openwechat.Message, slice g.SliceStr) string {
	weatherCode, err := FilterWeather(ctx, slice)
	if err != nil {
		return err.Error()
	}
	sender, err := msg.Sender()
	if err != nil {
		glog.Error(ctx, err.Error())
		err = gerror.New("异常")
		return err.Error()
	}
	glog.Debugf(ctx, "Received User%s[%s] Text Msg : %v", sender.NickName, sender.ID(), msg.Content)
	path := wxBotCfg.Storage + "/weather/" + sender.ID() + ".json"

	var weatherPosition = &dto.WeatherPosition{
		SubscribeStatus: true,
		City:            weatherCode.Name,
		Adcode:          weatherCode.Adcode,
		GroupsId:        sender.ID(),
		GroupsNickName:  sender.NickName,
		GroupsUserName:  sender.UserName,
	}
	_, err = gfile.Create(path)
	if err != nil {
		glog.Error(ctx, err.Error())
		err = gerror.New("异常")
		return err.Error()
	}

	//记录上下文到本地
	gfile.PutContents(path, gconv.String(weatherPosition))
	reply := "订阅成功，每天早上8点定时推送当天天气预报\n退订，请发送：退订天气"
	return reply
}

// 匹配数据条件找到acode和city
func FilterWeather(ctx context.Context, slice g.SliceStr) (data *dto.WeatherCode, err error) {

	var (
		weatherCodeArr   = &[]dto.WeatherCode{}
		weatherMathSlice []dto.WeatherCode
		path             = wxBotCfg.Storage
	)

	weatherJson, _ := gjson.Load(path + "/weatherCode.json")
	err = weatherJson.Scan(weatherCodeArr)
	if err != nil {
		glog.Error(ctx, "天气转换异常："+err.Error())
		if err != nil {
			err = gerror.New("异常")
			return
		}
	}

	title := "查询"
	if gstr.Contains(gconv.String(slice), "订阅") {
		title = "订阅"
	}

	time1 := time.Now()
	for _, v := range slice {
		for _, code := range *weatherCodeArr {
			if gstr.Contains(code.Name, v) {
				weatherMathSlice = append(weatherMathSlice, code)
			}
		}
	}
	midTime := time.Since(time1).Nanoseconds()
	mathCount := len(weatherMathSlice)
	log.Printf("花费时间：%v, weatherMathSlice: %v, 数量：%v", midTime, gconv.String(weatherMathSlice), mathCount)
	if mathCount > 0 {

		if mathCount == 1 {

			data = &dto.WeatherCode{
				Adcode: weatherMathSlice[0].Adcode,
				Name:   weatherMathSlice[0].Name,
			}
			return
		}
		if mathCount > 1 {

			log.Printf("这里匹配到有%v个", mathCount)
			replyMultWeather := ""
			for _, weather := range weatherMathSlice {
				replyMultWeather += title + "天气" + weather.Name + "\n"
			}

			err = gerror.New(fmt.Sprintf("匹配到有以下%v个，请发送：\n%v\n", mathCount, replyMultWeather))

		}
	} else {
		if title == "查询" {
			err = gerror.New("没匹配到，发送格式：查询北京天气")
		} else {
			err = gerror.New("没匹配到，发送格式：订阅北京天气")
		}

	}

	return
}

// 回复聊天
func ReplyMsg(ctx context.Context, msg *openwechat.Message, reply string) error {

	if msg.IsSendByGroup() {
		if msg.IsTickledMe() {

		} else {
			groupSender, err := msg.SenderInGroup()
			if err != nil {
				glog.Debugf(ctx, "get sender in group error :%v \n", err)
				return err
			}
			glog.Debugf(ctx, "groupSender: %v", gconv.String(*groupSender))
			replyName := groupSender.NickName
			reply = "@" + replyName + " 您好！我是智能机器人，我来回答你: " + " \n\n" + reply
		}

	}

	_, err := msg.ReplyText(reply)
	if err != nil {
		glog.Debugf(ctx, "response user error: %v \n", err)
		return err
	}
	return nil
}

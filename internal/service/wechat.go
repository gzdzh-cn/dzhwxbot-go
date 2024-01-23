package service

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/skip2/go-qrcode"
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
	self, err := h.bot.GetCurrentUser()
	if err != nil {
		glog.Errorf(ctx, "get current user error : %v", err)
		return
	}
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

// GroupMsg description
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

// 查询天气
func (h *MsgHandler) GetWeather(ctx context.Context) string {

	resp, err := client.Get(ctx, "https://restapi.amap.com/v3/weather/weatherInfo?city=440981&key=901c4be03b0e1776938ee5a555dc7166")
	if err != nil {
		glog.Debugf(ctx, "天气查询失败: %v", err.Error())
		return ""
	}
	defer resp.Close()
	weatherRes := &dto.WeatherRes{}
	gconv.Scan(resp.ReadAllString(), weatherRes)

	if weatherRes.Info == "OK" {
		live := weatherRes.Lives[0]
		data := fmt.Sprintf("地点: %v\n  天气: %v\n 气温: %v\n 风向: %v\n 风力: %v\n 湿度: %v ", live.Province+live.City, live.Weather, live.Temperature, live.Winddirection, live.Windpower, live.Humidity)
		glog.Debug(ctx, data)
		return data
	}
	return ""
}

func GetWeather(ctx context.Context) string {
	glog.Warning(ctx, "查询天气")

	resp, err := client.Get(ctx, "https://restapi.amap.com/v3/weather/weatherInfo?city=440981&key=901c4be03b0e1776938ee5a555dc7166")
	if err != nil {
		glog.Debugf(ctx, "天气查询失败: %v", err.Error())
		return ""
	}
	defer resp.Close()
	weatherRes := &dto.WeatherRes{}
	gconv.Scan(resp.ReadAllString(), weatherRes)

	if weatherRes.Info == "OK" {
		live := weatherRes.Lives[0]
		data := fmt.Sprintf("地点: %v\n天气: %v\n气温: %v\n风向: %v\n风力: %v\n湿度: %v", live.Province+live.City, live.Weather, live.Temperature+"摄氏度", live.Winddirection, live.Windpower+"级", live.Humidity)
		glog.Debug(ctx, data)
		return data
	}
	return ""
}

// 单个对话框
func ReadPersonMsg(ctx context.Context, msg *openwechat.Message, requestText string) error {

	reply := ""
	requestText = strings.TrimSpace(requestText)
	if requestText == "天气" {
		reply = GetWeather(ctx)
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
			if requestText == "天气" {
				reply = GetWeather(ctx)
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
				if requestText == "天气" {
					reply = GetWeather(ctx)
				} else {
					reply = NewChatGpt().SendConent(ctx, msg, "请用中文回答："+requestText)
				}

				return ReplyMsg(ctx, msg, reply)
			}
		}

	}

	return nil
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

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"wxbot/internal/dto"
)

type ChatGpt struct {
	content *string
}

var (
	chatReq *dto.ChatReq
)

func init() {

	chatReq = &dto.ChatReq{
		Action: "next",
		// Messages 切片初始化
		Messages: []dto.Message{
			{
				Id: "aaa2a222-af79-4d01-ae7e-cd7259b1f1be",
				Author: dto.Author{
					Role: "user",
				},
				Content: dto.Content{
					ContentType: "text",
				},
			},
		},
		Model: "text-davinci-002-render-sha",
	}
}

func NewChatGpt() *ChatGpt {
	return &ChatGpt{}
}

// 发送
func (c *ChatGpt) SendConent(ctx context.Context, msg *openwechat.Message, requestText string) string {

	reply, err := Request(ctx, msg, requestText)
	if err != nil {
		glog.Error(ctx, err.Error())
		return ""
	}
	sendMsg := c.ReplyMsg(ctx, msg, reply)
	return sendMsg
}

// 回复
func (c *ChatGpt) ReplyMsg(ctx context.Context, msg *openwechat.Message, reply string) string {

	//glog.Debugf(ctx, "Received  Text Msg : %v", msg.Content)

	return reply
}

// 请求chatgpt
func Request(ctx context.Context, msg *openwechat.Message, requestText string) (data string, err error) {

	var (
		r     = g.RequestFromCtx(ctx)
		reply = "请稍等......"
	)

	sender, err := msg.Sender()
	if err != nil {
		glog.Error(ctx, err.Error())
		return
	}
	glog.Debugf(ctx, "Received User%s[%s] Text Msg : %v", sender.NickName, sender.ID(), msg.Content)

	path := wxBotCfg.Storage + "/history/" + sender.ID() + ".json"
	exists := gfile.Exists(path)
	var contents *gjson.Json
	if exists {
		contents, _ = gjson.Load(path)
		chatReq.ParentMessageId = contents.Get("id").String()
		conversationId := contents.Get("conversation_id").String()
		chatReq.ConversationId = &conversationId

	} else {
		_, err = gfile.Create(path)
		if err != nil {
			glog.Error(ctx, err.Error())
			return
		}
		chatReq.ConversationId = nil
	}
	// 生成一个新的 UUID
	uuid := uuid.New().String()
	chatReq.Messages[0].Id = uuid
	chatReq.Messages[0].Content.Parts = g.SliceAny{requestText}
	glog.Debug(ctx, "chatReq", *chatReq)

	client.SetHeader("Authorization", "Bearer "+chatGptCfg.AccessToken)
	client.SetHeader("Content-Type", "application/json")
	resp, err := client.Post(ctx, chatGptCfg.RequestUrl+"/backend-api/conversation", chatReq)
	if err != nil {
		g.Log().Error(ctx, err)
		r.Response.WriteStatusExit(500)
	}
	defer resp.Close()
	defer resp.Body.Close()
	// 检查 resp.Body 是否已关闭
	select {
	case <-ctx.Done():
		glog.Debug(ctx, "Context canceled. Do not process response.")
		// 可以添加其他处理逻辑，例如记录错误或返回默认值
		return getDefaultReply("信息超时"), nil
	default:
		// 继续处理 response
	}

	// 复制 resp.Body 到 reply
	reply, err = copyResponseToReply(resp.Body)
	if err != nil {
		glog.Error(ctx, err.Error())
		// 可以添加其他处理逻辑，例如记录错误或返回默认值
		return getDefaultReply("文本转发失败"), nil
	}

	//截取最后一个流
	res := strLen(reply)

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("恢复", err)
			data = "网络异常，请重新提问。如果再次出现问题，请联系管理员！微信号：liliewin"
		}
	}()

	chatHistory := &dto.ChatHistory{
		Id:              uuid,
		ParentMessageId: chatReq.ParentMessageId,
		ConversationId:  res.ConversationId,
		WeatherPosition: &dto.WeatherPosition{
			SubscribeStatus: false,
			City:            "高州",
			Adcode:          "440981",
		},
	}

	gfile.PutContents(path, gconv.String(chatHistory))

	// 返回 reply
	return res.Message.Content.Parts[0], nil
}

// copyResponseToReply 将 respBody 复制到 reply 中
func copyResponseToReply(respBody io.Reader) (string, error) {
	// 读取 resp.Body 的内容并返回字符串
	bodyBytes, err := ioutil.ReadAll(respBody)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

// getDefaultReply 返回默认的 reply
func getDefaultReply(str string) string {
	// 返回默认值，或者进行其他处理
	return str
}

// 截取最后一个流
func strLen(dataString string) *dto.ChatRes {

	var resData *dto.ChatRes

	// 使用 strings 包的 Split 函数将字符串分割为多个 JSON 对象
	msgSlice := gstr.Split(dataString, "\ndata: ")

	glog.Debugf(ctx, "信息长度：%v", len(msgSlice))
	lastObject := ""

	if len(msgSlice) > 2 {
		// 获取最后一个 JSON 对象
		lastObject = msgSlice[len(msgSlice)-2]
		content, _ := gjson.LoadContent(lastObject)
		glog.Debugf(ctx, "lastObject: %v", content)
		// 解析 JSON
		err := json.Unmarshal([]byte(lastObject), &resData)
		if err != nil {
			glog.Error(ctx, "解析 JSON 失败:", err.Error())
			return resData
		}
	}
	//glog.Debugf(ctx, "resData: %v", resData)

	return resData
}

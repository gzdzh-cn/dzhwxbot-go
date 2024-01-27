package service

import (
	"context"
	"github.com/eatmoreapple/openwechat"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gbuild"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/util/gconv"
	"wxbot/internal/dto"
)

var (
	ctx        = context.Background()
	client     = g.Client()
	chatGptCfg = &dto.ChatGptCfg{}
	wxBotCfg   = &dto.WxBotCfg{}
	RunMode    = "dev"
	//self       *openwechat.Self
	bot *openwechat.Bot
)

func init() {
	g.Log().Debug(ctx, "service init")
	gconv.Scan(g.Cfg().MustGetWithEnv(ctx, "chatGpt").Map(), chatGptCfg)
	gconv.Scan(g.Cfg().MustGetWithEnv(ctx, "wxBot").Map(), wxBotCfg)

	gBuildData := gbuild.Data()
	buildData := &dto.BuiltData{}
	err := gconv.Scan(gBuildData, buildData)
	if err != nil {
		return
	}

	if buildData.BuiltTime != nil {
		RunMode = "prod"
	}

	if RunMode == "prod" {
		g.Cfg().GetAdapter().(*gcfg.AdapterFile).SetFileName("config_prod.yaml")
	}
	g.Log().Debugf(ctx, "RunMode:%v", RunMode)

}

// 启动机器人进程
func Boot(ctx context.Context) {
	// init log
	glog.SetDefaultLogger(g.Log())
	glog.SetFlags(glog.F_TIME_STD | glog.F_FILE_SHORT)

	// email init
	emailSetting, err := g.Cfg().Get(ctx, "emailSetting")
	if err == nil {
		err = emailSetting.Scan(&EmailDataSetting)
		if err != nil {
			glog.Fatal(ctx, err)
		}
		glog.Debug(ctx, EmailDataSetting)
	}

	// doc init
	token, err := g.Cfg().Get(ctx, "doctoken")
	if err != nil {
		glog.Fatal(ctx, err)
	}
	if token.String() != "" {
		glog.Debug(ctx, token.String())
		//NewSearchApi(ctx, token.String())
	}

	// wechat
	go RunWechat(ctx)
}

// 启动wx机器人
func RunWechat(ctx context.Context) {
	//bot := openwechat.DefaultBot()
	bot = openwechat.DefaultBot(openwechat.Desktop) // 桌面模式，上面登录不上的可以尝试切换这种模式

	handler := NewHandler(bot)

	// 注册消息处理函数
	bot.MessageHandler = handler.Handler

	// 注册登陆二维码回调
	bot.UUIDCallback = handler.QrCodeCallBack

	bot.SyncCheckCallback = handler.SyncCheckCallback

	_ = TimeTask(handler)

	// 创建热存储容器对象
	reloadStorage := openwechat.NewFileHotReloadStorage(wxBotCfg.Storage + "/botStorage.json")
	glog.Warning(ctx, "reloadStorage", reloadStorage)

	defer reloadStorage.Close()
	// 执行热登录
	err := bot.HotLogin(reloadStorage)
	//免扫码登录
	//err := bot.PushLogin(reloadStorage, openwechat.NewRetryLoginOption())
	if err != nil {
		if err = bot.Login(); err != nil {
			glog.Errorf(ctx, "login error: %v \n", err)
			return
		}
	}
	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}

// 定时任务
func TimeTask(handler *MsgHandler) (err error) {

	_, err = gcron.Add(ctx, "0 0 6 * * *", func(ctx context.Context) {

		exists := gfile.Exists(wxBotCfg.Storage + "/botStorage.json")
		if exists {
			self, _ := bot.GetCurrentUser()
			groups, _ := self.Groups(true)
			mps, _ := self.Mps(true)
			friends, _ := self.Friends(true)
			_ = gfile.PutContents(wxBotCfg.Storage+"/history/groups.json", gjson.MustEncodeString(groups))
			_ = gfile.PutContents(wxBotCfg.Storage+"/history/mps.json", gjson.MustEncodeString(mps))
			_ = gfile.PutContents(wxBotCfg.Storage+"/history/friends.json", gjson.MustEncodeString(friends))
			glog.Debug(ctx, "定时下载用户群，好友，公众号数据")
		}

	}, "GetData")

	_, err = gcron.Add(ctx, "0 0 8 * * *", func(ctx context.Context) {
		//批量推送天气
		handler.MultGetWeaher(ctx)
	}, "GetWeather")
	//gtimer.SetTimeout(ctx, 3*time.Second, func(ctx context.Context) {
	//fmt.Println("SetTimeout:", gtime.Now())
	//handler.MultNotice(ctx)
	//})

	return err
}

// MyWrite
type MyWrite struct {
}

// Write
func (w *MyWrite) Write(p []byte) (n int, err error) {
	glog.Skip(1).Debug(context.Background(), string(p)[20:])
	return len(p), nil
}

package service

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/gogf/gf/v2/os/glog"
	goconfluence "github.com/hailaz/confluence-go-api"
)

var basePath = "https://goframe.org"
var apiPath = basePath + "/rest/api"
var cAPI *goconfluence.API
var qaPath = basePath + "/pages/viewpage.action?pageId=7296348"
var replaceStr = [][2]string{
	{"@@@hl@@@", "["},
	{"@@@endhl@@@", "]"},
	{"][", ""},
}

// NewSearchApi description
func NewSearchApi(ctx context.Context, token string) {
	// goconfluence.SetDebug(true)
	// initialize a new api instance
	api, err := goconfluence.NewAPI(apiPath, "", token)
	if err != nil {
		glog.Fatal(ctx, err)
	}
	cAPI = api
	// get current user information
	currentUser, err := api.CurrentUser()
	if err != nil {
		glog.Fatal(ctx, err)
	}
	glog.Debugf(ctx, "%+v\n", currentUser)
}

// Search description
func Search(ctx context.Context, key string) string {
	glog.Debug(ctx, "search key: ", key)
	resStr := ""
	cql := fmt.Sprintf("siteSearch ~ '%s' AND space in ('%s')", key, "gf")
	if cAPI == nil {
		return "没有配置gf文档搜索api~"
	}
	res, err := cAPI.Search(goconfluence.SearchQuery{
		CQL:   cql,
		Limit: 3,
	})
	if err != nil {
		glog.Error(ctx, err)
		return "哎呀，出错啦~"
	}
	// g.Dump(res)
	if len(res.Results) > 0 {
		resStr = "搜索结果：\n"
		for _, v := range res.Results {
			glog.Debug(ctx, v)
			resStr += v.Title + "\n"
			resStr += basePath + v.Content.Links.WebUI + "\n"

			glog.Debug(ctx, v.Title)
			glog.Debug(ctx, v.Content.Title)
			glog.Debug(ctx, basePath+v.Content.Links.WebUI)

		}
		glog.Debug(ctx, resStr)
		resStr = strings.Replace(resStr, replaceStr[0][0], replaceStr[0][1], -1)
		resStr = strings.Replace(resStr, replaceStr[1][0], replaceStr[1][1], -1)
		resStr = strings.Replace(resStr, replaceStr[2][0], replaceStr[2][1], -1)
		resStr = html.UnescapeString(resStr)
		glog.Debug(ctx, resStr)
		resStr += "其它常见问题：\n" + qaPath
	} else {
		resStr = "没有找到呢~换个关键字试试^_^\n也可以访问最新文档\n" + basePath
	}

	return resStr
}

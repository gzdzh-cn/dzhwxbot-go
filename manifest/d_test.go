package main_test

import (
	"context"
	"github.com/go-ego/gse"
	"github.com/go-ego/gse/hmm/pos"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/util/gconv"
	"log"
	"testing"
	"wxbot/internal/dto"
)

var ctx = context.Background()
var (
	seg              gse.Segmenter
	posSeg           pos.Segmenter
	keyWord          = "我想知道东山的天气"
	text             = "订阅茂名天气"
	weatherMathSlice []dto.WeatherCode
	newSegment, _    = gse.New("zh", "alpha")
)

func Member(id int, MemberFunc func(fid_ int)) {
	MemberFunc(id)
}

func TestFor(t *testing.T) {

	var funcSlice []func(id int)
	for i := 0; i < 3; i++ {

		Member(i, func(id_ int) {
			log.Println("Member i", i)
		})

		//i := i
		funcSlice = append(funcSlice, func(id int) {
			log.Println("funcSlice i", i)
		})

		funcSlice[len(funcSlice)-1](i)

	}

	log.Println("分割线.....")

	for i := 0; i < 3; i++ {
		y := i
		for _, f := range funcSlice {
			f(y)
		}
	}

}

type Group struct {
	Name string `json:"name"`
}

func Search(g *Group, limit int, searchFunc func(group *Group) bool) bool {
	log.Println("search", limit)
	if searchFunc(g) {
		log.Println("search searchFunc")
	}
	return true
}

func MembersSearch(limit int, searchFuncList ...func(group *Group) bool) bool {
	group := &Group{}
	return Search(group, limit, func(group *Group) bool {
		log.Println("MembersSearch", limit)
		for _, searchFunc := range searchFuncList {
			if !searchFunc(group) {
				return false
			}
		}
		return true
	})
}

func GroupsSearch(limit int, searchFuncList ...func(group *Group) bool) bool {
	return MembersSearch(limit, func(group *Group) bool {
		log.Println("GroupsSearch", limit)
		for _, searchFunc := range searchFuncList {
			if !searchFunc(group) {
				return false
			}
		}
		return true
	})
}

func TestFunc(t *testing.T) {
	slice := g.SliceStr{"name", "sex"}
	GroupsSearch(1, func(group *Group) bool {
		for _, v := range slice {
			log.Println("TestFunc func1 v", v)
		}
		return true
	}, func(group *Group) bool {
		for _, v := range slice {
			log.Println("TestFunc func2 v", v)
		}
		return true
	}, func(group *Group) bool {
		for _, v := range slice {
			log.Println("TestFunc func3 v", v)
		}
		return true
	})

	var funcSlice []func(group *Group) bool
	for i := 0; i < 3; i++ {
		y := i
		funcSlice = append(funcSlice, func(group *Group) bool {
			for _, v := range slice {
				log.Printf("TestFunc func%v v:%v", y, v)
			}
			return true
		})
	}
	GroupsSearch(1, funcSlice...)

}

func TestMap(t *testing.T) {

	type Student struct {
		UserName string `json:"userName"`
		Sex      string `json:"sex"`
		Age      int    `json:"age"`
	}
	type School struct {
		Name     string `json:"name"`
		Children []Student
	}
	schoolJson := `
 [
  {
    "name": "School1",
    "children": [
      {
        "userName": "Alice",
        "age": 18
      },
      {
        "userName": "Bob",
        "sex": "Male",
        "age": 20
      }
    ]
  }
]
`
	//var schools *[]School   //错误的写法
	//var schools []*School    //正确的写法
	var schools = &[]School{} //正确的写法

	err := gconv.Scan(schoolJson, &schools)
	if err != nil {
		glog.Error(ctx, err.Error())
		return
	}

	g.Dump(schools)

}

func TestName(t *testing.T) {

	seg.LoadDict()
	hmm := newSegment.Cut(text, true)
	glog.Debug(ctx, "cut use hmm: ", hmm)

	hmm = newSegment.CutSearch(text, true)
	glog.Debug(ctx, "cut search use hmm: ", hmm)

	// 分词文本
	tb := []byte(text)

	// 处理分词结果
	glog.Debug(ctx, "输出分词结果, 类型为字符串, 使用搜索模式: ", seg.String(text, true))
	glog.Debug(ctx, "输出分词结果, 类型为 slice: ", seg.Slice(text))

	segments := seg.Segment(tb)
	// 处理分词结果, 普通模式
	glog.Debug(ctx, "普通模式", gse.ToString(segments))

	segments1 := seg.Segment([]byte(text))
	// 搜索模式
	glog.Debug(ctx, "搜索模式", gse.ToString(segments1, true))
}

package ACGImageClassify

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/tidwall/gjson"
)

var (
	botpath, _ = os.Getwd() // 当前bot运行目录
	datapath   = botpath + "/data/acgimage/"
	cache_file = datapath + "cache"
	cache_uri  = "file:///" + cache_file
	head       = "http://sayuri.fumiama.top:62002/dice?class=9&url="
	lastvisit  = time.Now().Unix()
	comments   = []string{
		"[0]这啥啊",
		"[1]普通欸",
		"[2]有点可爱",
		"[3]不错哦",
		"[4]很棒",
		"[5]我好啦!",
		"[6]影响不好啦!",
		"[7]太涩啦，🐛了!",
		"[8]已经🐛不动啦...",
	}
)

func init() {
	os.RemoveAll(datapath) //清除缓存
	err := os.MkdirAll(datapath, 0755)
	if err != nil {
		panic(err)
	}
}

func Flush() {
	lastvisit = time.Now().Unix()
}

func CanVisit(delay int64) bool {
	if time.Now().Unix()-lastvisit > delay {
		Flush()
		return true
	}
	return false
}

func Classify(ctx *zero.Ctx, targeturl string, noimg bool) {
	lv := lastvisit
	if targeturl[0] != '&' {
		targeturl = url.QueryEscape(targeturl)
	}
	get_url := head + targeturl
	if noimg {
		get_url += "&noimg=true"
	}
	resp, err := http.Get(get_url)
	if err != nil {
		ctx.Send(fmt.Sprintf("ERROR: %v", err))
	} else {
		if noimg {
			data, err1 := ioutil.ReadAll(resp.Body)
			if err1 == nil {
				dhash := gjson.GetBytes(data, "img").String()
				class := int(gjson.GetBytes(data, "class").Int())
				replyClass(ctx, dhash, class, noimg, lv)
			} else {
				ctx.Send(fmt.Sprintf("ERROR: %v", err1))
			}
		} else {
			class, err1 := strconv.Atoi(resp.Header.Get("Class"))
			dhash := resp.Header.Get("DHash")
			if err1 != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err1))
			}
			defer resp.Body.Close()
			// 写入文件
			data, _ := ioutil.ReadAll(resp.Body)
			f, _ := os.OpenFile(cache_file+strconv.FormatInt(lv, 10), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
			defer f.Close()
			f.Write(data)
			replyClass(ctx, dhash, class, noimg, lv)
		}
	}
}

func replyClass(ctx *zero.Ctx, dhash string, class int, noimg bool, lv int64) {
	img := message.Image(cache_uri + strconv.FormatInt(lv, 10))
	if class > 5 {
		if dhash != "" && !noimg {
			b14, err3 := url.QueryUnescape(dhash)
			if err3 == nil {
				ctx.Send(comments[class] + "给你点提示哦:" + b14)
			}
			ctx.Event.GroupID = 0
			ctx.SendGroupMessage(0, img)
			ctx.SendChain(message.Text("偷偷发给你啦，不要和别人说哦"), img)
		} else {
			ctx.Send(comments[class])
		}
	} else {
		comment := message.Text(comments[class])
		if !noimg {
			ctx.SendChain(message.Image(cache_uri+strconv.FormatInt(lv, 10)), comment)
		} else {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), comment)
		}
	}
}

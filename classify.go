package ACGImageClassify

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
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
			ctx.Send(message.Text("偷偷发给你啦，不要和别人说哦"))
			ctx.SendGroupMessage(0, img.Add("cache", "1"))
			ctx.Send(bigPic(cache_file+strconv.FormatInt(lv, 10), comments[class]))
		} else {
			ctx.Send(comments[class])
		}
	} else {
		comment := message.Text(comments[class])
		if !noimg {
			ctx.SendChain(img, comment)
		} else {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), comment)
		}
	}
}

// PicHash 返回图片的 md5 值
func picHash(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.ToUpper(fmt.Sprintf("%x", md5.Sum(data)))
}

// BigPic 返回一张XML大图CQ码
func bigPic(file string, comment string) string {
	var hash = picHash(file)
	return fmt.Sprintf(`[CQ:xml,data=<?xml version='1.0' 
encoding='UTF-8' standalone='yes' ?><msg serviceID="5" 
templateID="12345" action="" brief="不够涩！" 
sourceMsgId="0" url="" flag="0" adverSign="0" multiMsgFlag="0">
<item layout="0" advertiser_id="0" aid="0"><image uuid="%s.jpg" md5="%s" 
GroupFiledid="2235033681" filesize="81322" local_path="%s.jpg" 
minWidth="200" minHeight="200" maxWidth="500" maxHeight="1000" />
</item><source name="%s" icon="" 
action="" appid="-1" /></msg>]`,
		hash,
		hash,
		hash,
		comment,
	)
}

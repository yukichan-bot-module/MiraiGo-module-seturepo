package seturepo

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"path"
	"sync"
	"time"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"gopkg.in/yaml.v2"
)

var instance *seturepo
var logger = utils.GetModuleLogger("com.aimerneige.seturepo")
var repo map[string][2]string
var blacklistUser []int64
var allowedList []int64

type seturepo struct {
}

func init() {
	instance = &seturepo{}
	bot.RegisterModule(instance)
}

func (s *seturepo) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "com.aimerneige.seturepo",
		Instance: instance,
	}
}

// Init 初始化过程
// 在此处可以进行 Module 的初始化配置
// 如配置读取
func (s *seturepo) Init() {
	path := config.GlobalConfig.GetString("aimerneige.seturepo.path")
	if path == "" {
		path = "./seturepo.yaml"
	}
	bytes := utils.ReadFile(path)
	err := yaml.Unmarshal(bytes, &repo)
	if err != nil {
		logger.WithError(err).Errorf("Unable to read config file in %s", path)
	}
	blacklistUserSlice := config.GlobalConfig.GetIntSlice("aimerneige.seturepo.blacklist")
	for _, user := range blacklistUserSlice {
		blacklistUser = append(blacklistUser, int64(user))
	}
	allowedListSlice := config.GlobalConfig.GetIntSlice("aimerneige.seturepo.allowed")
	for _, user := range allowedListSlice {
		allowedList = append(allowedList, int64(user))
	}
}

// PostInit 第二次初始化
// 再次过程中可以进行跨 Module 的动作
// 如通用数据库等等
func (s *seturepo) PostInit() {
}

// Serve 注册服务函数部分
func (s *seturepo) Serve(b *bot.Bot) {
	b.GroupMessageEvent.Subscribe(func(c *client.QQClient, msg *message.GroupMessage) {
		if !isAllowed(msg.GroupCode) {
			return
		}
		if msg.Sender.IsAnonymous() {
			return
		}
		if inBlacklist(msg.Sender.Uin) {
			return
		}
		for k, v := range repo {
			if msg.ToString() == k {
				c.SendGroupMessage(msg.GroupCode, message.NewSendingMessage().Append(message.NewText(v[0])))
				sendSetu(c, msg.Sender.Uin, v[1])
			}
		}
	})
	b.PrivateMessageEvent.Subscribe(func(c *client.QQClient, msg *message.PrivateMessage) {
		for k, v := range repo {
			if msg.ToString() == k {
				sendSetu(c, msg.Sender.Uin, v[1])
			}
		}
	})
}

// Start 此函数会新开携程进行调用
// ```go
//
//	go exampleModule.Start()
//
// ```
// 可以利用此部分进行后台操作
// 如 http 服务器等等
func (s *seturepo) Start(b *bot.Bot) {
}

// Stop 结束部分
// 一般调用此函数时，程序接收到 os.Interrupt 信号
// 即将退出
// 在此处应该释放相应的资源或者对状态进行保存
func (s *seturepo) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
}

func sendSetu(c *client.QQClient, id int64, dir string) {
	imgData, err := getSetuImg(dir)
	if err != nil {
		logger.WithError(err).Error("Unable to get img.")
	}
	imgMsgElement, err := c.UploadPrivateImage(id, imgData)
	if err != nil {
		logger.WithError(err).Error("Unable to Upload img.")
	}
	imgMsg := message.NewSendingMessage().Append(imgMsgElement)
	c.SendPrivateMessage(id, imgMsg)
}

func getSetuImg(dir string) (io.ReadSeeker, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logger.WithError(err).Errorf("Fail to read dir %s", dir)
	}
	rand.Seed(time.Now().Unix())
	imgFile := files[rand.Intn(len(files))]
	// 检测是否读到文件夹，如果是则重试三次，否则报错
	for i := 0; i < 3 && imgFile.IsDir(); i++ {
		imgFile = files[rand.Intn(len(files))]
	}
	if imgFile.IsDir() {
		return nil, fmt.Errorf("Fail to get a file in dir %s", dir)
	}
	imgBytes, err := ioutil.ReadFile(path.Join(dir, imgFile.Name()))
	if err != nil {
		logger.WithError(err).Errorf("Fail to read img file %s", imgFile.ModTime())
	}
	return bytes.NewReader(imgBytes), nil
}

func inBlacklist(userUin int64) bool {
	for _, v := range blacklistUser {
		if userUin == v {
			return true
		}
	}
	return false
}

func isAllowed(groupCode int64) bool {
	for _, v := range allowedList {
		if groupCode == v {
			return true
		}
	}
	return false
}

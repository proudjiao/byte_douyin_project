# 抖音极简版

<!-- PROJECT SHIELDS -->

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]

<!-- links -->

[your-project-path]: proudjiao/byte_douyin_project
[contributors-shield]: https://img.shields.io/github/contributors/proudjiao/byte_douyin_project.svg?style=flat-square
[contributors-url]: https://github.com/proudjiao/byte_douyin_project/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/proudjiao/byte_douyin_project.svg?style=flat-square
[forks-url]: https://github.com/proudjiao/byte_douyin_project/network/members
[stars-shield]: https://img.shields.io/github/stars/proudjiao/byte_douyin_project.svg?style=flat-square
[stars-url]: https://github.com/proudjiao/byte_douyin_project/stargazers
[issues-shield]: https://img.shields.io/github/issues/proudjiao/byte_douyin_project.svg?style=flat-square
[issues-url]: https://img.shields.io/github/issues/proudjiao/byte_douyin_project.svg
[license-shield]: https://img.shields.io/github/license/proudjiao/byte_douyin_project?style=flat-square
[license-url]: https://github.com/proudjiao/byte_douyin_project/blob/master/LICENSE

- [数据库说明](#数据库说明)
  - [数据库关系说明](#数据库关系说明)
  - [数据库建立说明](#数据库建立说明)
- [架构说明](#架构说明)
  - [各模块代码详细说明](#各模块代码详细说明)
    - [Handlers](#handlers)
    - [Service](#service)
    - [Models](#models)
- [遇到的问题及对应解决方案](#遇到的问题及对应解决方案)
  - [返回 json 数据的完整性和前端要求的一致性](#返回json数据的完整性和前端要求的一致性)
  - [is_favorite 和 is_follow 字段的更新](#is_favorite和is_follow字段的更新)
  - [视频的保存和封面的切片](#视频的保存和封面的切片)
    - [视频的保存](#视频的保存)
    - [封面的截取](#封面的截取)
- [可改进的地方](#可改进的地方)
- [项目运行](#项目运行)

## 数据库说明

![database.png](https://p6-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/892fbbe46695467ebe4fb4a12ebd412e~tplv-k3u1fbpfcp-watermark.image?)

> 单纯看上面的图会感觉很混乱，现在我们来将关系拆解。

### 数据库关系说明

**关系图如下：**

![database_relation.png](https://p6-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/f08918db1ea84126bc21d23fe9401a75~tplv-k3u1fbpfcp-watermark.image?)

> 所有的表都有自己的 id 主键为唯一的标识。

user_logins：存下用户的用户名和密码

user_infos：存下用户的基本信息

videos：存下视频的基本信息

comment：存下每个评论的基本信息

**具体的关系索引：**

所有的一对一和一对多关系，只需要在一个表中加入对方的 id 索引。

- 比如 user_infos 和 user_logins 的一对一关系，在 user_logins 中加入 user_id 字段设为外键存储 user_infos 中对应的行的 id 信息。
- 比如 user_infos 和和 videos 的一对多关系，在 videos 中加入 user_id 字段设为外键存储 user_infos 中对应的行的 id 信息。

所有的多对多关系，需要多建立一张表，用该表作为媒介存储两个对象的 id 作为关系的产生，而它们各自表中不需要再存下额外的字段。

- 比如 user_infos 和 videos 的多对多关系，创建一张 user_favor_videos 中间表，然后将该表的字段均设为外键，分别存下 user_infos 和 videos 对应行的 id。如 id 为 1 的用户对 id 为 2 的视频点了个赞，那么就把这个 1 和 2 存入中间表 user_favor_videos 即可。

### 数据库建立说明

数据库各表的建立无需自己实现额外的建表操作，一切都由 gorm 框架自动建表，具体逻辑在 models 层的代码中。

> gorm 官方文档链接：[链接](https://gorm.io/zh_CN/docs/index.html)

建表和初始化操作由 init_db.go 来执行：

```go
func InitDB() {
	var err error
	DB, err = gorm.Open(mysql.Open(config.DBConnectString()), &gorm.Config{
		PrepareStmt:            true, //缓存预编译命令
		SkipDefaultTransaction: true, //禁用默认事务操作
		//Logger:                 logger.Default.LogMode(logger.Info), //打印sql语句
	})
	if err != nil {
		panic(err)
	}
	err = DB.AutoMigrate(&UserInfo{}, &Video{}, &Comment{}, &UserLogin{})
	if err != nil {
		panic(err)
	}
}
```

## 架构说明

![architecture.png](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/ae11d82b8de74787a258ef36f4cf2557~tplv-k3u1fbpfcp-watermark.image?)

> 以用户登录为例共需要经过以下过程：

1. 进入中间件 SHAMiddleWare 内的函数逻辑，得到 password 明文加密后再设置 password。具体需要调用 gin.Context 的 Set 方法设置 password。随后调用 next()方法继续下层路由。
2. 进入 UserLoginHandler 函数逻辑，获取 username，并调用 gin.Context 的 Get 方法得到中间件设置的 password。再调用 service 层的 QueryUserLogin 函数。
3. 进入 QueryUserLogin 函数逻辑，执行三个过程：checkNum，prepareData，packData。也就是检查参数、准备数据、打包数据，准备数据的过程中会调用 models 层的 UserLoginDAO。
4. 进入 UserLoginDAO 的逻辑，执行最终的数据库请求过程，返回给上层。

### 各模块代码详细说明

我开发的过程中是以单个函数为单个文件进行开发，所以代码会比较长，故我根据数据库内的模型对函数文件进行了如下分包：

![handlers.png](https://p1-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/6dc222793d6f4038b1bf2435053bfee4~tplv-k3u1fbpfcp-watermark.image?)

service 层的分包也是一样的。

#### Handlers

对于 handlers 层级的所有函数实现有如下规范：

所有的逻辑由代理对象进行，完成以下两个逻辑

1. 解析得到参数。
2. 开始调用下层逻辑。

例如一个关注动作触发的逻辑：

```go
NewProxyPostFollowAction().Do()
//其中Do主要包含以下两个逻辑，对应两个方法
p.parseNum() //解析参数
p.startAction() //开始调用下层逻辑
```

#### Service

对于 service 层级的函数实现由如下规范：

同样由一个代理对象进行，完成以下三个或两个逻辑

当上层需要返回数据信息，则进行三个逻辑：

1. 检查参数。
2. 准备数据。
3. 打包数据。

当上层不需要返回数据信息，则进行两个逻辑：

1. 检查参数。
2. 执行上层指定的动作。

例如关注动作在 service 层的逻辑属于第二类：

```go
NewPostFollowActionFlow(...).Do()
//Do中包含以下两个逻辑
p.checkNum() //检查参数
p.publish() //执行动作
```

#### Models

对于 models 层的各个操作，没有像 service 和 handler 层针对前端发来的请求就行对应的处理，models 层是面向于数据库的增删改查，不需要考虑和上层的交互。

而 service 层根据上层的需要来调用 models 层的不同代码请求数据库内的内容。

## 遇到的问题及对应解决方案

### 返回 json 数据的完整性和前端要求的一致性

由于数据库内的一对一、一对多、多对多关系是根据 id 进行映射，所以 models 层请求得到的字段并不包含前端所需要的直接数据，比如前端要求 Comment 结构中需要包含 UserInfo，而我的 Comment 结构如下：

```go
type Comment struct {
	Id         int64     `json:"id"`
	UserInfoId int64     `json:"-"` //用于一对多关系的id
	VideoId    int64     `json:"-"` //一对多，视频对评论
	User       UserInfo  `json:"user" gorm:"-"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"-"`
	CreateDate string    `json:"create_date" gorm:"-"`
}
```

很明显，为了与数据库中设计的表一一对应，在原数据的基础上加了几个字段，且在 gorm 屏蔽了 User 字段，所以 service 调用 models 层得到是 Comment 数据中 User 字段还未被填充，还需再填充这部分内容，好在由对应的 UserId，故可以正确填充该字段。

为了重用以及不破坏代码的一致性，将填充逻辑写入 util 包内，比如以上的字段填充函数，同时前端要求的日期格式也能够按要求设置：

```go
func FillCommentListFields(comments *[]*models.Comment) error {
	size := len(*comments)
	if comments == nil || size == 0 {
		return errors.New("util.FillCommentListFields comments为空")
	}
	dao := models.NewUserInfoDAO()
	for _, v := range *comments {
		_ = dao.QueryUserInfoById(v.UserInfoId, &v.User) //填充这条评论的作者信息
		v.CreateDate = v.CreatedAt.Format("1-2")         //转为前端要求的日期格式
	}
	return nil
}
```

这里举了 Comment 这一个例子，其他的 Video 也是同理。

### is_favorite 和 is_follow 字段的更新

每次为视频点赞都会在数据库的 user_favor_videos 表中加入用户的 id 和视频的 id，很明显 is_favorite 字段是针对每个用户来判断的，而我所设计的数据库中的 videos 表也是包含这个字段的，但这个字段很明显不能直接进行复用，而是需要每次判断用户和此视频的关系来重新更新。

这个更新过程放入 util 包的填充函数中即可，为了点赞过程的迅速响应，我采取了 nosql 的方式存储了这个点赞的映射，也就是 userId 和 videoId 的映射，也就是用 nosql 代替了这个中间表的功效。

具体代码逻辑在 cache 包内。

### 视频的保存和封面的切片

#### 视频的保存

在本地建立 static 文件夹存储视频和封面图片。

具体逻辑如下：

1. 检查视频格式
2. 根据 userId 和该作者发布的视频数量产生唯一的名称，如 id 为 1 的用户发布了 0 个视频，那么本次上传的名称为 1-0.mp4
3. 截取第一帧画面作为封面
4. 保存视频基本信息到数据库（包括视频链接和封面链接

#### 封面的截取

使用 ffmpeg 调用命令行对视频进行截取。

设计 ffmpeg 请求类 Video2Image，通过对它内部的参数设置来构造对应的命令行字符串。具体请看 util 包内的 ffmpeg.go 的实现。

由于我设计的命令请求字符串是直接的一行字符串，而 go 语言 exec 包里面的 Command 函数执行所需的仅仅是一个个参数。

所以此处我想到用 cgo 直接调用 system(args)来解决。

代码如下：

```go
//#include <stdlib.h>
//int startCmd(const char* cmd){
//	  return system(cmd);
//}
import "C"

func (v *Video2Image) ExecCommand(cmd string) error {
	if v.debug {
		log.Println(cmd)
	}
	cCmd := C.CString(cmd)
	defer C.free(unsafe.Pointer(cCmd))
	status := C.startCmd(cCmd)
	if status != 0 {
		return errors.New("视频切截图失败")
	}
	return nil
}
```

## 可改进的地方

1. 写到后面发现很多 mysql 的数据可以用 redis 优化。
2. 很多执行逻辑可以通过并行优化。
3. 路由分组可以更为详实。
4. ...

## 项目运行

> 本项目运行不需要手动建表，项目启动后会自动建表。

**运行所需环境**：

- mysql 5.7 及以上
- redis 5.0.14 及以上
- ffmepg（已放入 lib 自带，用于对视频切片得到封面
- 需要 gcc 环境（主要用于 cgo，windows 请将 mingw-w64 设置到环境变量

**运行需要更改配置**：

> 进入 config 目录更改对应的 mysql、redis、server、path 信息。

- mysql：mysql 相关的配置信息
- redis：redis 相关配置信息
- server：当前服务器（当前启动的机器）的配置信息，用于生成对应的视频和图片链接
- path：其中 ffmpeg_path 为 lib 里的文件路径，static_source_path 为本项目的 static 目录，这里请根据本地的绝对路径进行更改

> 完成 config 配置文件的更改后，需要再更改 conf.go 里的解析文件路径为 config.toml 文件的绝对路径，内容如下：
>
> ```go
> if _, err := toml.DecodeFile("你的绝对路径\\config.toml", &Info); err != nil {
> 		panic(err)
> 	}
> ```

**运行所需命令**：

```shell
cd .\byte_douyin_project\
go run main.go
```

***模拟机注意事项***

此repo只提供后端，完整运行需要下载前端apk

若在Android Studio上运行该模拟器，要链接至后端需要双击右下角“我”，并设置baseURL为服务器地址

(If you want to refer to the computer which is running the Android simulator, use the IP address 10.0.2.2 instead)



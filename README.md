# mtRSSConverter

一个专为M-Team RSS订阅设计的转换器，解决群晖Synology Download Station无法正常下载M-Team种子的问题。

## 问题背景

M-Team的RSS订阅链接存在以下问题，导致无法在群晖Synology Download Station中正常使用：

1. **302重定向问题**：种子链接的302重定向导致Download Station无法正确获取种子文件
2. **时间戳参数重复**：种子链接包含时间戳参数，Download Station会将其识别为不同的种子，导致重复下载失败

## 解决方案

mtRSSConverter通过以下方式解决这些问题：

1. **代理转发**：作为中间代理，处理M-Team RSS订阅的请求
2. **URL重写**：将种子链接重写为本地服务地址，避免302重定向
3. **去重处理**：通过GUID标识符避免重复下载
4. **数据库缓存**：使用SQLite数据库存储种子信息，提高响应速度

## 主要功能

- 🚀 代理M-Team RSS订阅请求
- 🔄 自动重写种子下载链接
- 💾 SQLite数据库缓存种子信息
- 🐳 Docker容器化部署
- 🌐 支持自定义服务地址配置

## 技术架构

- **后端框架**：Gin (Go)
- **数据库**：SQLite + GORM
- **容器化**：Docker
- **编程语言**：Go 1.24+

## 快速开始

### 方法一：自行编译部署（推荐）

1. **克隆项目**
```bash
git clone https://github.com/srt180/mtRSSConverter.git
cd mtRSSConverter
```

2. **编译Docker镜像**
```bash
docker build -t mtrssconverter .
```

3. **运行容器**
```bash
docker run -e SQLITE_PATH="./data/mtrssconverter.db" \
           -e BASE_ADDR="http://192.168.2.10:8080" \
           -d -p 8080:8080 \
           -v ./data:/root/data \
           --name mtc \
           mtrssconverter
```

### 方法二：使用预构建镜像

```bash
docker run -e SQLITE_PATH="./data/mtrssconverter.db" \
           -e BASE_ADDR="http://192.168.2.10:8080" \
           -d -p 8080:8080 \
           -v ./data:/root/data \
           --name mtc \
           registry.cn-hangzhou.aliyuncs.com/srt180/mtrssconverter:latest
```

## 环境变量配置

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `SQLITE_PATH` | `./mtRSSConverter.db` | SQLite数据库文件路径 |
| `BASE_ADDR` | `http://localhost:8080` | 服务基础地址（用于生成fetch URL） |

## 使用方法

### 1. 获取M-Team RSS订阅链接

在M-Team网站获取您的RSS订阅链接，例如：
```
https://rss.m-team.cc/api/rss/fetch?categories=xxx&dl=1&onlyFav=1&pageSize=10&sign=xxxxx&t=xxxxx&tkeys=ttitle&uid=xxxxx
```

### 2. 转换为本地服务地址

在M-Team RSS链接前添加您的服务地址：
```
http://192.168.2.10:8080/rss/https://rss.m-team.cc/api/rss/fetch?categories=xxx&dl=1&onlyFav=1&pageSize=10&sign=xxxxx&t=xxxxx&tkeys=ttitle&uid=xxxxx
```

### 3. 在Download Station中添加订阅

将转换后的链接添加到群晖Synology Download Station的RSS订阅中即可正常使用。

## API接口

### RSS转换接口
- **路径**：`GET /rss/*url`
- **功能**：获取并转换M-Team RSS订阅内容
- **参数**：`*url` - M-Team RSS订阅链接

### 种子下载接口
- **路径**：`GET /fetch/:guid`
- **功能**：下载指定GUID对应的种子文件
- **参数**：`:guid` - 种子的唯一标识符

## 部署说明

### 群晖NAS部署

1. 在群晖DSM中安装Docker套件
2. 使用提供的`run.sh`脚本或手动创建容器
3. 确保端口8080未被占用
4. 配置数据卷挂载以持久化数据库

### 注意事项

- 确保NAS的IP地址配置正确
- 数据库文件路径需要持久化存储
- 建议使用固定IP或DDNS地址

## 项目结构

```
mtRSSConverter/
├── config/          # 配置管理
│   ├── config.go    # 配置结构定义
│   └── db.go        # 数据库初始化
├── handlers/         # HTTP处理器
│   └── rss.go       # RSS相关处理逻辑
├── models/           # 数据模型
│   └── item.go      # 种子条目模型
├── services/         # 业务逻辑
│   └── rss.go       # RSS转换服务
├── main.go           # 主程序入口
├── Dockerfile        # Docker构建文件
├── go.mod            # Go模块依赖
└── run.sh            # 快速运行脚本
```

## 开发说明

### 本地开发环境

```bash
# 安装依赖
go mod tidy

# 运行开发服务器
go run main.go
```

### 构建二进制文件

```bash
# 构建Linux版本
CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o mtRSSConverter .

# 构建macOS版本
go build -o mtRSSConverter .
```

## 许可证

本项目采用MIT许可证，详见LICENSE文件。

## 贡献

欢迎提交Issue和Pull Request来改进这个项目。

## 联系方式

如有问题或建议，请通过GitHub Issues联系。

---

**注意**：请确保您有合法的M-Team账号和RSS订阅权限。本工具仅用于解决技术问题，请遵守相关网站的使用条款。

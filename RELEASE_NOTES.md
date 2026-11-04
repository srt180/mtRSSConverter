# v0.1.2 — Download Station RSS 兼容性修复

本次基于旧版 RSS 转换服务发布补丁，尝试解决 issue #2 中“将排除无效或已封锁的链接”的问题。新版 Web 界面及下载管理功能仍在开发中，不包含在这次补丁里。

## 修复内容

- RSS 的 `link` 和 `enclosure` 统一指向带 `.torrent` 后缀的代理下载地址。
- 下载响应提供种子类型、附件文件名和正确长度，支持无响应体的 HEAD 探测；旧的无后缀地址继续可用。
- 为缺少 GUID 的条目生成独立且稳定的映射，避免下载链接为空或多个条目覆盖同一个映射。
- 上游 HTTP 错误不会再作为种子返回；容器增加 HTTPS 根证书。
- 提供 linux/amd64、linux/arm64 镜像。

## 更新方案

1. 先记录原容器的环境变量、端口和数据目录挂载；停止容器后备份原数据目录。不要删除数据库。
2. 拉取版本镜像：

   ```sh
   docker pull registry.cn-hangzhou.aliyuncs.com/srt180/mtrssconverter:v0.1.2
   ```

3. 使用新镜像重新创建容器，保留原有端口、`SQLITE_PATH`、数据库挂载和其他配置。群晖 Container Manager 可下载该镜像后按原配置重建。数据库格式和容器内旧版路径保持兼容。
4. 将 `BASE_ADDR` 设置为群晖可访问的转换服务地址，例如 `http://192.168.1.20:8080`。这里的 IP 是运行转换服务的设备 IP；不要使用 `localhost` 或 Docker 内部地址。
5. 启动后，在 Download Station 中刷新 RSS 订阅，再尝试手动下载或自动下载规则。

原先按 README 部署的用户，可在原部署目录中参考以下命令（先将示例 IP、端口和路径改成自己的原配置）：

```sh
docker stop mtc
# 此时另行备份 ./data；保留旧容器方便回退
docker rename mtc mtc-before-v0.1.2
docker run -d --name mtc -p 8080:8080 \
  -e SQLITE_PATH=./data/mtrssconverter.db \
  -e BASE_ADDR=http://192.168.1.20:8080 \
  -v "$PWD/data:/root/data" \
  registry.cn-hangzhou.aliyuncs.com/srt180/mtrssconverter:v0.1.2
```

## 验证范围

已通过 Go 测试与静态检查；回归测试覆盖 RSS 转换、缺少 GUID 的条目、302 重定向、签名参数、GET/HEAD、旧链接及上游错误。尚未进行群晖实机验证，欢迎在 issue #2 反馈结果。反馈时请隐藏 RSS 链接中的签名、token 等敏感参数。

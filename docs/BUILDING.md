# 构建

开发依赖：Go 1.27、Node 24、Wails CLI 2.14。

```powershell
npm --prefix frontend ci
npm --prefix frontend run typecheck
npm --prefix frontend run build
gofmt -w .
go test ./...
go vet ./...
wails build -clean -trimpath -platform windows/amd64
```

GitHub Actions 使用六个原生 runner 构建 Windows x64/ARM64、macOS Intel/Apple Silicon、Linux x64/ARM64。Linux runner 安装 GTK3/WebKitGTK 4.1，并以 Wails v2 的 `webkit2_41` build tag 编译。CI 成功只证明对应 runner 可以编译和打包；不证明真实串口、驱动、权限或烧录效果。

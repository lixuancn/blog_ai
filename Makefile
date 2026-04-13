build_aliyun:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o blogapp cmd/server/main.go

scp_aliyun:
	scp -r blogapp root@120.27.119.218:/usr/local/www/blog_ai/
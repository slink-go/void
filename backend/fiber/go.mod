module backend/fiber

go 1.22.3

require (
	github.com/gofiber/fiber/v2 v2.52.4
	github.com/gofiber/template/html/v2 v2.1.1
	github.com/slink-go/api-gateway v0.0.0-00010101000000-000000000000
	github.com/slink-go/logging v0.0.2
	github.com/valyala/fasthttp v1.53.0
)

replace github.com/slink-go/api-gateway => ./../../api-gateway

require (
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/gofiber/template v1.8.3 // indirect
	github.com/gofiber/utils v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/slink-go/disco-go v0.0.18 // indirect
	github.com/slink-go/disco/common v0.0.8 // indirect
	github.com/slink-go/go-eureka-client v1.1.1 // indirect
	github.com/slink-go/httpclient v0.0.8 // indirect
	github.com/slink-go/logger v0.0.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
)

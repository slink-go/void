module backend

go 1.22.3

require (
	github.com/gofiber/fiber/v2 v2.52.4
	github.com/slink-go/api-gateway v0.0.0
	github.com/slink-go/logging v0.0.2
)

replace github.com/slink-go/api-gateway => ./../api-gateway

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rs/zerolog v1.32.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
)

package main

import (
	"context"
	"strings"

	"github.com/miebyte/goutils/cores"
	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/logging"

	"github.com/superwhys/novel/api"
	"github.com/superwhys/novel/config"
	"github.com/superwhys/novel/content"
	"github.com/superwhys/novel/internal/library"
)

var serverConfigFlag = flags.Struct("server", (*config.ServerConfig)(nil), "novel server config")

func main() {
	flags.Parse()

	serverConfig := new(config.ServerConfig)
	logging.PanicError(serverConfigFlag(serverConfig))

	lib, err := library.Load(content.Files)
	logging.PanicError(err)

	logging.Infow(
		context.Background(),
		"novel library loaded",
		logging.String("title", lib.Novel().Title),
		logging.Int("chapters", lib.Novel().ChapterCount),
	)

	srv := cores.NewCores(
		cores.WithHttpHandler("/api", api.New(lib)),
	)

	if err := cores.Start(srv, serverConfig.Addr); err != nil && !isGracefulStop(err) {
		logging.PanicError(err)
	}
}

func isGracefulStop(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Signal:")
}

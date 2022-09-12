package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	"github.com/nitwhiz/no-load/internal/cold"
	"os"
	"path"
)

type CLIOptions struct {
	IgnoreHeader []string `short:"i" long:"ignore-header" description:"header name to ignore when hashing"`
}

func main() {
	cliOpts := CLIOptions{}

	args, err := flags.Parse(&cliOpts)

	targetUrl := args[0]
	dataDir := args[1]

	if err != nil {
		panic(err)
	}

	opts := cold.Options{
		TargetUrl:     targetUrl,
		DataDir:       dataDir,
		IgnoreHeaders: cliOpts.IgnoreHeader,
	}

	wd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(path.Join(wd, "data/"), 0777)

	if err != nil {
		panic(err)
	}

	r := gin.Default()

	r.Use(cors.Default())

	r.Any("/*proxyPath", func(c *gin.Context) {
		creq, err := cold.NewRequest(c.Request)

		if err != nil {
			panic(err)
		}

		cresp, err := creq.ToResponse(&opts)

		if err != nil {
			panic(err)
		}

		for k, v := range cresp.Headers {
			c.Header(k, v)
		}

		c.Data(cresp.Status, cresp.ContentType, cresp.Body)
	})

	err = r.Run("0.0.0.0:3700")

	if err != nil {
		panic(err)
	}
}

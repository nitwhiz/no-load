package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nitwhiz/no-load/internal/cold"
	"os"
	"path"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("not enough arguments.")
	}

	targetUrl := os.Args[1]
	dataDir := os.Args[2]

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

		cresp, err := creq.ToResponse(targetUrl, dataDir)

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

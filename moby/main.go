package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/moby/tool/src/moby"
)

type Create struct {
	Config string `form:"config"`
}

func createImage(config []byte, filename string) error {
	m, err := moby.NewConfig(config)
	if err != nil {
		log.Fatalf("Invalid config: %v", err)
	}
	buf := new(bytes.Buffer)
	if err := moby.Build(m, buf, false, ""); err != nil {
		log.Fatalf("%v", err)
	}
	image := buf.Bytes()
	if err := moby.Outputs(filename, image, []string{"iso-bios", "iso-efi"}, 1024, false); err != nil {
		log.Fatalf("Error writing outputs: %v", err)
	}
	return nil
}

func main() {
	router := gin.Default()

	port := os.Getenv("PORT")
	outputDir := os.Getenv("OUTPUT_DIR")
	moby.MobyDir = os.Getenv("MOBY_DIR")
	router.POST("/create", func(c *gin.Context) {
		var form Create

		err := c.Bind(&form)
		if err != nil {
			c.JSON(400, gin.H{
				"error": "Bad Request",
			})
			log.Fatal(err)
		}

		log.Printf("%v", c.PostForm("config"))
		config := []byte(form.Config)

		// Create Filename Hash
		h := md5.New()
		h.Write([]byte(config))
		filename := outputDir + hex.EncodeToString(h.Sum(nil))

		if err := createImage(config, filename); err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to createImage()",
			})
		} else {
			c.JSON(200, gin.H{
				"path": filename + ".iso",
			})

		}

	})
	router.Run(":" + port)
}

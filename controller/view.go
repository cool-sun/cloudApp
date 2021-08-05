package controller

import (
	"github.com/gin-gonic/gin"
)

type View struct {
}

func (v View) Index(c *gin.Context) {
	c.File("./view/index.html")
}

package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"strconv"
)

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, "0 "+strconv.Itoa(int(time.Now().Unix())))
}

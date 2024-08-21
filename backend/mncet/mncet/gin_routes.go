package mncet

import (
	"mncet/mncet/databases"
	"mncet/mncet/tools"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddHost(c *gin.Context, database databases.Databases) {
	var HostInfo tools.Hosts
	if err := c.ShouldBindJSON(&HostInfo); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	// database.AddHosts([]tools.Hosts{HostInfo})

	c.JSON(http.StatusOK, gin.H{"status": 200})
}

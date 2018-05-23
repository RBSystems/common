package log

import (
	"net/http"

	"github.com/labstack/echo"
)

//SetLogLevel depends on a :level parameter in the endpoint
func SetLogLevel(c echo.Context) error {
	lvl := c.Param("level")

	L.Info("Setting log level to %v")
	err := SetLevel(lvl)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	L.Info("Log level set to %v", lvl)
	return c.JSON(http.StatusOK, "ok")
}

func GetLogLevel(c echo.Context) error {

	L.Info("Getting log level.")
	lvl, err := GetLevel()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	L.Infof("Log level is %v", lvl)

	m := make(map[string]string)
	m["log-level"] = lvl

	return c.JSON(http.StatusOK, m)
}

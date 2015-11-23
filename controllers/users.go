package controllers

import (
	"net/http"

	"github.com/asartalo/axyaserve/models"
	"github.com/gin-gonic/gin"
)

type Users struct {
	AppDb models.AppDb
}

func (ctrl *Users) NewUser(c *gin.Context) {
	responder := NewResponder(c)
	var creds Credentials
	err := c.BindJSON(&creds)
	if err != nil {
		responder.Error(
			http.StatusBadRequest,
			"Make sure credentials are complete.",
			err.Error(),
		)
		return
	}
	user, err := ctrl.AppDb.NewUser(creds.Name, creds.Password)
	if err != nil {
		responder.Error(
			http.StatusInternalServerError,
			"Error: "+err.Error(),
			err.Error(),
		)
		return
	}
	responder.Created("User created", gin.H{"name": user.Name})
}

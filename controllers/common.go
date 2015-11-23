package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Credentials struct {
	Name     string `form:"name" json:"name" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type Responder struct {
	context *gin.Context
}

func NewResponder(c *gin.Context) *Responder {
	return &Responder{c}
}

func (j *Responder) Error(status int, message, details string) {
	j.context.JSON(status, gin.H{"message": message, "details": details})
}

func (j *Responder) Okay(message string, payload gin.H) {
	j.context.JSON(http.StatusOK, gin.H{"message": message, "payload": payload})
}

func (j *Responder) Created(message string, payload gin.H) {
	j.context.JSON(http.StatusCreated, gin.H{"message": message, "payload": payload})
}

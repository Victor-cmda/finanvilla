package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"finanvilla/internal/domain/entities"
	"finanvilla/internal/domain/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user entities.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.CreateUser(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	users, total, err := h.userService.List(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *UserHandler) UpdateSettings(c *gin.Context) {
	// Pegar o ID diretamente da URL
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	var settings entities.UserSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validação básica dos campos
	if err := validateSettings(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdateSettings(
		c.Request.Context(),
		userID,
		&settings,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Settings updated successfully",
		"settings": settings,
	})
}

func validateSettings(settings *entities.UserSettings) error {
	if settings.Theme != "dark" && settings.Theme != "light" {
		return fmt.Errorf("theme must be 'dark' or 'light'")
	}

	// Validação básica do formato da linguagem (xx-XX)
	if !regexp.MustCompile(`^[a-z]{2}-[A-Z]{2}$`).MatchString(settings.Language) {
		return fmt.Errorf("invalid language format")
	}

	// Lista de moedas suportadas
	validCurrencies := []string{"BRL", "USD", "EUR"}
	currencyValid := false
	for _, c := range validCurrencies {
		if c == settings.Currency {
			currencyValid = true
			break
		}
	}
	if !currencyValid {
		return fmt.Errorf("unsupported currency")
	}

	return nil
}

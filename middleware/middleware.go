package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	b64 "encoding/base64"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	jwtSigningMethod = jwt.SigningMethodHS256
	jwtSignatureKey  = []byte(os.Getenv("APP_SECRET_KEY_JWT"))
)

// BasicAuth - authentication with basic auth
func BasicAuth(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")

	if !strings.Contains(authHeader, "Basic") {
		result := gin.H{
			"status":  http.StatusForbidden,
			"message": "invalid token",
			"href":    c.Request.RequestURI,
		}
		c.JSON(http.StatusForbidden, result)
		c.Abort()
		return
	}

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	tokenString := strings.Replace(authHeader, "Basic ", "", -1)
	myToken := clientID + ":" + clientSecret
	myBasicAuth := b64.StdEncoding.EncodeToString([]byte(myToken))
	if tokenString != myBasicAuth {
		result := gin.H{
			"status":  http.StatusUnauthorized,
			"message": "Unauthorized user",
			"href":    c.Request.RequestURI,
		}
		c.JSON(http.StatusUnauthorized, result)
		c.Abort()
		return
	}
}

// JWTAuth - auth token jwt
func JWTAuth(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	if !strings.Contains(authHeader, "Bearer") {
		result := gin.H{
			"status":  http.StatusForbidden,
			"message": "invalid token",
			"href":    c.Request.RequestURI,
		}
		c.JSON(http.StatusForbidden, result)
		c.Abort()
		return
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", -1)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Signing method invalid")
		} else if method != jwtSigningMethod {
			return nil, fmt.Errorf("Signing method invalid")
		}

		return jwtSignatureKey, nil
	})

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if err != nil {
		log.Error().Msg(err.Error())
		result := gin.H{
			"status":  http.StatusUnauthorized,
			"message": err.Error(),
		}
		c.JSON(http.StatusUnauthorized, result)
		c.Abort()
		return
	}

	_, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		result := gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Error check payload token",
		}
		c.JSON(http.StatusInternalServerError, result)
		c.Abort()
		return
	}
}

// JwtAuthWithHeader - auth token jwt with header
func JwtAuthWithHeader(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	myUserID := c.Request.Header.Get("userid")
	if myUserID == "" {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":   http.StatusBadRequest,
				"messages": "Header cant empty",
			},
		)
		c.Abort()
		return
	}

	userID, _ := strconv.Atoi(myUserID)

	if !strings.Contains(authHeader, "Bearer") {
		result := gin.H{
			"status":   http.StatusForbidden,
			"messages": "invalid token",
			"href":     c.Request.RequestURI,
		}
		c.JSON(http.StatusForbidden, result)
		c.Abort()
		return
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", -1)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Signing method invalid")
		} else if method != jwtSigningMethod {
			return nil, fmt.Errorf("Signing method invalid")
		}

		return jwtSignatureKey, nil
	})

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if err != nil {
		log.Error().Msg(err.Error())
		result := gin.H{
			"status":   http.StatusUnauthorized,
			"messages": err.Error(),
		}
		c.JSON(http.StatusUnauthorized, result)
		c.Abort()
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok {
		myID := claims["user_id"].(float64)
		if userID != int(myID) {
			result := gin.H{
				"status":   http.StatusUnauthorized,
				"messages": "Unauthorize user",
			}
			c.JSON(http.StatusUnauthorized, result)
			c.Abort()
			return
		}
	}
}

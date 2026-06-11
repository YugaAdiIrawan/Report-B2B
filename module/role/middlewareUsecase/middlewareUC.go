package middlewareUsecase

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/YugaAdiIrawan/module/role/miiddlewareRepo"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Middleware interface {
	MiddlewareRouteRoles(context *gin.Context)
}
type middlewareUsecase struct {
	middlewareRepository miiddlewareRepo.MiddlewareRepository
}

func NewMiddlewareUsecase(middlewareRepo miiddlewareRepo.MiddlewareRepository) Middleware {
	return &middlewareUsecase{
		middlewareRepo,
	}
}

// MiddlewareRouteRoles - middleware for insert audit trail
func (uc *middlewareUsecase) MiddlewareRouteRoles(c *gin.Context) {
	myUserID := c.Request.Header.Get("userid")
	moduleName := c.FullPath()
	methodName := c.Request.Method
	var (
		//desc, reqBody string
		userID    int
		errUserID error
	)

	simplifiedEndpoint := simplifyEndpoint(moduleName)

	// Check if route is already exist on dataLake
	isRouteRegistered, errIsrouteRegistered := uc.middlewareRepository.RetrieveRouteAlreadyRegistered(methodName, simplifiedEndpoint)
	if errIsrouteRegistered != nil {
		log.Error().Msg(errIsrouteRegistered.Error())
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":   http.StatusBadRequest,
				"messages": "something wrong on checked registered route",
			},
		)
		c.Abort()
		return
	}

	if !isRouteRegistered {
		log.Info().Msg("Allowed with notes, " + simplifyEndpoint(moduleName) + " not registered ")
		c.Next()
		return
	}

	if myUserID != "" {
		userID, errUserID = strconv.Atoi(myUserID)
		if errUserID != nil {
			log.Print("MiddlewareRouteRoles errUserID =", errUserID.Error())
			return
		}
	}

	userDetail, errUserDetail := uc.middlewareRepository.RetrieveUserByUserID(userID)
	if errUserDetail != nil {
		log.Error().Msg(errUserDetail.Error())
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":   http.StatusBadRequest,
				"messages": "User not found",
			},
		)
		c.Abort()
		return
	}
	if userDetail == nil {
		log.Warn().Int("userID", userID).Msg("MiddlewareRouteRoles: userDetail is nil")
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":   http.StatusUnauthorized,
			"messages": "User not found",
		})
		c.Abort()
		return
	}

	log.Debug().Msg(strconv.Itoa(userID))
	log.Debug().Msg(strconv.Itoa(userDetail.RoleID))
	log.Debug().Msg(methodName)
	log.Debug().Msg(simplifyEndpoint(moduleName))

	isAllowed, errIsAllowed := uc.middlewareRepository.RetrieveUserRoleAllowanceIsExist(userDetail.RoleID, methodName, simplifiedEndpoint)
	if errIsAllowed != nil {
		log.Error().Msg(errIsAllowed.Error())
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":   http.StatusBadRequest,
				"messages": "something error",
			},
		)
		c.Abort()
		return
	}

	if !isAllowed {
		result := gin.H{
			"status":  http.StatusNotAcceptable,
			"message": "Unauthorized user - plaese ask admin to allow you to access : " + simplifyEndpoint(moduleName),
		}
		c.JSON(http.StatusNotAcceptable, result)
		c.Abort()
		return
	}

}

// simplifyEndpoint replaces parameters in the endpoint path with placeholders
func simplifyEndpoint(endpoint string) string {
	parts := strings.Split(endpoint, "/")
	for i, part := range parts {
		if part != "" && (strings.HasPrefix(part, ":") || isDynamic(part)) {
			parts[i] = "*"
		}
	}

	log.Debug().Msg("joined string " + strings.Join(parts, "/"))
	return strings.Join(parts, "/")
}

// isDynamic checks if a string is a parameter (e.g., numeric or alphanumeric)
func isDynamic(s string) bool {
	return !isStaticSegment(s)
}

// isStaticSegment checks if a segment is static (not dynamic)
func isStaticSegment(s string) bool {
	// Check if segment is alphanumeric with special characters allowed in static paths
	for _, char := range s {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_') {
			return false
		}
	}
	return true
}

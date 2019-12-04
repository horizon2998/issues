package controller

import (
	"issues/model"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	jwt_lib "github.com/dgrijalva/jwt-go"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

/*
const (
	STATIC_PATH = "upload"
)
*/

type Controller struct {
	// DB instance
	DB *gorm.DB

	// Cấu hình config
	Config model.Config
}

func NewController() *Controller {
	var c Controller
	return &c
}

func tokenGenerate(user model.Users) (string, error) {
	token := jwt_lib.New(jwt_lib.GetSigningMethod("HS256"))

	token.Claims = jwt_lib.MapClaims{
		"userId": user.ID,
		//"Role":   user.Role,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}
	//fmt.Println(user.ID)
	return token.SignedString([]byte(model.SecretKey))

}

type userJSON struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func (gc *Controller) LoginJSON(c *gin.Context) {
	var loginInfo userJSON

	err := c.BindJSON(&loginInfo)

	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorMesssage{
			Message: "Thông tin đăng nhập không hợp lệ",
		})
		return
	}

	//
	var user model.Users

	err = gc.DB.Raw(`
			SELECT *FROM Users
			WHERE phone = ? and password = ?
		`, loginInfo.Phone, loginInfo.Password).Scan(&user).Error

	//fmt.Println(loginInfo.Phone)

	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorMesssage{
			Message: "Can't login =((",
		})
		// log.Println(err)
		return
	}
	log.Println("---------------", user)

	//var userLogin model.Users
	var token string

	if token, err = tokenGenerate(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error,cannot create login session!",
		})
		return
	}

	log.Println("------", user.ID)

	c.JSON(http.StatusOK, gin.H{
		"jwt_token": token,
	})

	return
}
func (gc *Controller) ListIssues(c *gin.Context) {

	var issuesdefault []model.IssueDefault
	errGetIssues := gc.DB.Raw(`
		SELECT  *FROM Issues
	`).Scan(&issuesdefault).Error

	if errGetIssues != nil {
		log.Println(errGetIssues)
		c.JSON(http.StatusInternalServerError, model.ErrorMesssage{
			Message: "Lỗi server",
		})
		return
	}
	log.Println("---------------", issuesdefault)

	c.JSON(http.StatusOK, issuesdefault)
}

func (gc *Controller) GetProfile(c *gin.Context) {
	var user model.Users
	var headerInfo model.AuthorizationHeader

	if err := c.ShouldBindHeader(&headerInfo); err != nil {
		c.JSON(200, err)
	}

	tokenFromHeader := strings.Replace(headerInfo.Token, "Bearer ", "", -1)

	log.Println("-----", tokenFromHeader)

	claims := jwt_lib.MapClaims{}
	tkn, err := jwt_lib.ParseWithClaims(tokenFromHeader, claims, func(token *jwt_lib.Token) (interface{}, error) {
		return []byte(model.SecretKey), nil
	})

	if err != nil {
		if err == jwt_lib.ErrSignatureInvalid {
			log.Println("error 1")
			c.JSON(http.StatusUnauthorized, model.ErrorMesssage{
				Message: "Token không hợp lệ",
			})
			return
		}
		log.Println("error 2", err)
		c.JSON(http.StatusBadRequest, model.ErrorMesssage{
			Message: "Bad request",
		})
		return
	}
	if !tkn.Valid {
		log.Println("error 3")
		c.JSON(http.StatusUnauthorized, model.ErrorMesssage{
			Message: "Token không hợp lệ",
		})
		return
	}

	var IdFromToken string
	//var RoleFromToken string

	log.Println("---------", claims)

	for k, v := range claims {
		if k == "userId" {
			IdFromToken = v.(string)
		}
	}
	log.Println("-----------", IdFromToken)

	err = gc.DB.Raw(`
			SELECT *FROM Users
			WHERE id = ? 
		`, IdFromToken).Scan(&user).Error

	//fmt.Println(loginInfo.Phone)

	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorMesssage{
			Message: "Can Find ID From Token =((",
		})
		log.Println(err)
		return

	}
	c.JSON(http.StatusOK, user)

	return
}

//type

//type postissueJSON struct {}

func (gc *Controller) PostIssue(c *gin.Context) {

}

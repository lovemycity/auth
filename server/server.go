package server

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/globalsign/mgo/bson"
	"github.com/lovemycity/auth/tpl"
	"github.com/lovemycity/auth/user"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func Start() {
	var (
		err error
	)
	cnf, err = getConfig()
	if err != nil {
		log.Fatal(err)
	}
	mCl, err := mongo.NewClient(options.Client().ApplyURI(cnf.MongoDSN))
	if err != nil {
		log.Fatal(err)
	}
	if err := mCl.Connect(context.Background()); err != nil {
		log.Fatal(err)
	}
	db = mCl.Database("auth").Collection("users")
	srv := gin.Default()
	srv.Use(func(ctx *gin.Context) {
		ctx.Header("server", "@lovemycity/auth")
	})
	srv.NoRoute(noRoute)
	store, err := redis.NewStore(10,
		cnf.RedisNetwork,
		cnf.RedisAddr,
		cnf.RedisPassword,
		[]byte(cnf.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		Domain:   cnf.SessionDomain,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
		MaxAge:   60 * 60 * 24 * 30,
	})
	if err != nil {
		log.Fatal(err)
	}
	srv.Use(sessions.Sessions(cnf.SessionName, store))
	srv.Use(static.ServeRoot("/", "./ui/public"))
	srv.Use(mwCors)
	srv.OPTIONS("*any", handleOptions)
	srv.GET("/", handleStaticLogin)
	srv.GET("/register", handleStaticRegister)
	api := srv.Group("/api")
	{
		api.GET("/me", handleGetUser)
		api.POST("/register", handleRegister)
		api.POST("/login", handleLogin)
	}
	if err := srv.Run(":" + cnf.Port); err != nil {
		log.Fatal(err)
	}
}

var (
	db  *mongo.Collection
	cnf *Config
)

func mwCors(ctx *gin.Context) {
	ctx.Header("server", "@lovemycity/auth")
	ctx.Header("access-control-allow-origin", ctx.GetHeader("origin"))
	ctx.Header("access-control-allow-methods", "GET, POST")
	ctx.Header("access-control-allow-headers", "content-type,authorization")
	ctx.Header("access-control-allow-credentials", "true")
}

func handleOptions(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}

func noRoute(ctx *gin.Context) {
	if strings.HasPrefix(ctx.FullPath(), "/api") {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "invalid route",
		})
		return
	}
	ctx.String(http.StatusNotFound, "Not Found")
}

func handleStaticLogin(ctx *gin.Context) {
	tpl.WriteLayout(ctx.Writer, &tpl.LoginPage{
		BasePage: tpl.BasePage{
			SessionDomain: cnf.SessionDomain,
		},
	})
}

func handleStaticRegister(ctx *gin.Context) {
	tpl.WriteLayout(ctx.Writer, &tpl.RegisterPage{
		BasePage: tpl.BasePage{
			SessionDomain: cnf.SessionDomain,
		},
	})
}

func handleGetUser(ctx *gin.Context) {
	s := sessions.Default(ctx)
	if us := s.Get("@user"); us != nil {
		if us == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"message": "not authorized",
			})
			return
		}
		u := us.(*user.User)
		if u.Picture == "" {
			hasher := md5.Sum([]byte(u.Email))
			hash := hex.EncodeToString(hasher[:])
			u.Picture = "https://www.gravatar.com/avatar/" + hash
		}
		ctx.JSON(http.StatusOK, us)
	} else {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "not authorized",
		})
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`

	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`

	DateOfBirth time.Time `json:"date_of_birth"`
}

func (rr *registerRequest) validate() error {
	if strings.Trim(rr.Email, " ") == "" {
		return errors.Errorf("%s is an invalid email", rr.Email)
	}
	if rr.Password == "" || len(rr.Password) < 8 {
		return errors.New("invalid password")
	}
	return nil
}

func handleRegister(ctx *gin.Context) {
	req := new(registerRequest)
	if err := ctx.BindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": errors.Wrap(err,
				"failed to parse json").Error(),
		})
		return
	}
	if err := req.validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	pass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": errors.Wrap(err, "failed to crypt password"),
		})
		return
	}
	u := &user.User{
		ID:          primitive.NewObjectID(),
		Email:       req.Email,
		Password:    string(pass),
		DateOfBirth: primitive.NewDateTimeFromTime(req.DateOfBirth),
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		MiddleName:  req.MiddleName,
		CreatedAt:   primitive.NewDateTimeFromTime(time.Now()),
		CreatedIP:   ctx.ClientIP(),
	}
	_, err = db.InsertOne(ctx, u)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": errors.Wrap(err, "failed to save user").Error(),
		})
		return
	}
	s := sessions.Default(ctx)
	s.Set("@user", u)
	if err := s.Save(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": errors.Wrap(err, "failed to save session").Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, u)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (lr *loginRequest) validate() error {
	if lr.Email == "" {
		return errors.New("invalid email")
	}
	if lr.Password == "" {
		return errors.New("invalid password")
	}
	return nil
}

func handleLogin(ctx *gin.Context) {
	req := new(loginRequest)
	if err := ctx.BindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": errors.Wrap(err,
				"failed to parse json").Error(),
		})
		return
	}
	if err := req.validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": errors.Wrap(err,
				"validation failed").Error(),
		})
		return
	}
	u := new(user.User)
	res := db.FindOne(ctx, bson.M{"email": req.Email})
	if res.Err() != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": errors.Wrap(res.Err(),
				"user not found").Error(),
		})
		return
	}
	if err := res.Decode(u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": errors.Wrap(err,
				"failed to decode user").Error(),
		})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": errors.Wrap(err,
				"invalid password").Error(),
		})
		return
	}
	s := sessions.Default(ctx)
	s.Set("@user", u)
	if err := s.Save(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": errors.Wrap(err,
				"failed to save session").Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, u)
}

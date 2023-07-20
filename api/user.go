package api

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/r-scheele/escout/db/sqlc"
	"github.com/r-scheele/escout/util"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	ID                int64     `json:"id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	// Parse JSON request body
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Hash password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var user db.User
	// Prepare params for CreateUser method
	arg := db.CreateUserParams{
		Username:  req.Username,
		Password:  hashedPassword,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}

	user, err = server.store.GetUser(ctx, db.GetUserParams{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			user, err = server.store.CreateUser(ctx, arg)
			if err != nil {
				log.Printf("Error creating user: %v\n", err)
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}
		}
	} else {
		ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("user already exists")))
		return
	}
	rsp := newUserResponse(user)

	// Return success response
	ctx.JSON(http.StatusOK, rsp)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, db.GetUserParams{
		Username: req.Username,
	})
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.CheckPassword(req.Password, user.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, rsp)
}

type getUserRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type getUserResponse struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (server *Server) getUser(ctx *gin.Context) {
	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUserByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := getUserResponse{
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
	ctx.JSON(http.StatusOK, rsp)
}

type getUsersRequest struct {
	Page int32 `form:"page" binding:"required,min=1"`
	Size int32 `form:"size" binding:"required,min=1,max=100"`
}

type getUsersResponse struct {
	Users []getUserResponse `json:"users"`
	Total int32             `json:"total"`
}

func (server *Server) getUsers(ctx *gin.Context) {
	var req getUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	users, err := server.store.ListUsers(ctx, db.ListUsersParams{
		Limit:  req.Size,
		Offset: (req.Page - 1) * req.Size,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := getUsersResponse{
		Users: make([]getUserResponse, 0, len(users)),
		Total: int32(len(users)),
	}
	for _, user := range users {
		rsp.Users = append(rsp.Users, getUserResponse{
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		})
	}
	ctx.JSON(http.StatusOK, rsp)
}
